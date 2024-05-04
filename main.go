package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Environment struct {
	LISTEN_ADDRESS       string
	PLAUSIBLE_SCRIPT_URL string
	PLAUSIBLE_API_URL    string
}

func ParseEnvironment() *Environment {
	LISTEN_ADDRESS := os.Getenv("LISTEN_ADDRESS")
	PLAUSIBLE_SCRIPT_URL := os.Getenv("PLAUSIBLE_SCRIPT_URL")
	PLAUSIBLE_API_URL := os.Getenv("PLAUSIBLE_API_URL")

	if LISTEN_ADDRESS == "" {
		LISTEN_ADDRESS = "localhost:8080"
	}

	if PLAUSIBLE_SCRIPT_URL == "" {
		PLAUSIBLE_SCRIPT_URL = "https://plausible.io/js/%s"
	}

	if PLAUSIBLE_API_URL == "" {
		PLAUSIBLE_API_URL = "https://plausible.io/api/event"
	}

	return &Environment{LISTEN_ADDRESS: LISTEN_ADDRESS, PLAUSIBLE_SCRIPT_URL: PLAUSIBLE_SCRIPT_URL, PLAUSIBLE_API_URL: PLAUSIBLE_API_URL}
}

func buildGetScriptHandler(plausibleScriptUrl string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		scriptExtension := chi.URLParam(r, "name")

		url := fmt.Sprintf(plausibleScriptUrl, scriptExtension)

		response, error := http.Get(url)

		if error != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))

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
		_, error = io.Copy(w, response.Body)

		if error != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		}
	}
}

func buildPostEventHandler(plausibleApiUrl string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		request, error := http.NewRequest(r.Method, plausibleApiUrl, r.Body)

		if error != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))

			return
		}

		for key, values := range r.Header {
			for _, value := range values {
				normalizedKey := strings.ToLower(key)
				isCookieHeader := normalizedKey == "cookie"
				isCloudflareHeader := strings.HasPrefix(normalizedKey, "cf-")

				// Let's not copy the cookie and cloudflare headers
				isAddable := !isCookieHeader && !isCloudflareHeader

				if isAddable {
					fmt.Println(key, value)
					request.Header.Add(key, value)
				}
			}
		}

		client := http.DefaultClient
		response, error := client.Do(request)

		if error != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))

			return
		}

		defer response.Body.Close()

		// Copying headers from the origin response to the final response
		for key, values := range response.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		// Copying the status code from the origin response to the final response
		w.WriteHeader(response.StatusCode)

		// Copying the body from the origin response to the final response
		_, error = io.Copy(w, response.Body)

		if error != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		}
	}
}

func buildGetHealthHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(http.StatusText(http.StatusOK)))
	}
}

func main() {
	r := chi.NewRouter()

	env := ParseEnvironment()

	r.Use(middleware.Logger)

	r.Get("/health", buildGetHealthHandler())
	r.Get("/js/{name}", buildGetScriptHandler(env.PLAUSIBLE_SCRIPT_URL))
	r.Post("/api/event", buildPostEventHandler(env.PLAUSIBLE_API_URL))

	LISTEN_ADDRESS := (*env).LISTEN_ADDRESS

	http.ListenAndServe(LISTEN_ADDRESS, r)
}
