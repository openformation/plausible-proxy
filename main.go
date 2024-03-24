package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var LISTEN_ADDR = os.Getenv("LISTEN_ADDR")

func getScript(w http.ResponseWriter, r *http.Request) {
	scriptExtension := chi.URLParam(r, "name")

	url := fmt.Sprintf("https://plausible.io/js/%s", scriptExtension)

	response, err := http.Get(url)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))

		return
	}

	if response.StatusCode != http.StatusOK {
		w.WriteHeader(response.StatusCode)
		w.Write([]byte(http.StatusText(response.StatusCode)))

		return
	}

	defer response.Body.Close()

	// Copying headers from the origin request to the response
	for key, values := range response.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Copying the status code from the origin request to the response
	w.WriteHeader(response.StatusCode)

	// Copying the body from the origin request to the response
	_, err = io.Copy(w, response.Body)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}
}

func main() {
	var r = chi.NewRouter()

	r.Use(middleware.Logger)

	r.Get("/js/{name}", getScript)

	if LISTEN_ADDR == "" {
		LISTEN_ADDR = "localhost:8080"
	}

	http.ListenAndServe(LISTEN_ADDR, r)
}
