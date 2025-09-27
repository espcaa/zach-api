package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

type QuoteResponse struct {
	Quote string `json:"quote"`
	Index int    `json:"index"`
}

func main() {
	signal.Ignore(syscall.SIGPIPE)

	godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Println("Starting server on port :" + port)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	r.Get("/quotes/random", func(w http.ResponseWriter, r *http.Request) {
		file, err := os.Open("quotes.json")
		if err != nil {
			http.Error(w, "Could not open quotes file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		var quotes []string
		if err := json.NewDecoder(file).Decode(&quotes); err != nil {
			http.Error(w, "Could not decode quotes file", http.StatusInternalServerError)
			return
		}

		// pick random
		randomIndex := rand.Intn(len(quotes))
		quote := quotes[randomIndex]

		response := QuoteResponse{
			Quote: quote,
			Index: randomIndex,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// quote by index

	r.Get("/quotes/{index}", func(w http.ResponseWriter, r *http.Request) {
		index := chi.URLParam(r, "index")

		file, err := os.Open("quotes.json")
		if err != nil {
			http.Error(w, "Could not open quotes file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		var quotes []string
		if err := json.NewDecoder(file).Decode(&quotes); err != nil {
			http.Error(w, "Could not decode quotes file", http.StatusInternalServerError)
			return
		}

		intIndex := 0
		_, err = fmt.Sscanf(index, "%d", &intIndex)
		if err != nil || intIndex < 0 || intIndex >= len(quotes) {
			http.Error(w, "Invalid index", http.StatusBadRequest)
			return
		}

		quote := quotes[intIndex]

		response := QuoteResponse{
			Quote: quote,
			Index: intIndex,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	r.Get("/photos/{filename}", func(w http.ResponseWriter, r *http.Request) {
		filename := chi.URLParam(r, "filename")
		http.ServeFile(w, r, "photos/"+filename)
	})

	r.Get("/photos/random", func(w http.ResponseWriter, r *http.Request) {
		files, err := os.ReadDir("photos")
		if err != nil || len(files) == 0 {
			http.Error(w, "Could not read photos directory", http.StatusInternalServerError)
			return
		}

		randomIndex := rand.Intn(len(files))
		randomFile := files[randomIndex]

		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		http.ServeFile(w, r, "photos/"+randomFile.Name())
	})

	http.ListenAndServe(":"+port, r)
}
