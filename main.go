package main

import (
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	ENV_KEY_LOG_FORMAT   = "LOG_FORMAT"
	ENV_KEY_SERVER_ADDR  = "SERVER_ADDR"
	ENV_KEY_SECRET       = "SECRET"
	ENV_KEY_DATABASE_URL = "DATABASE_URL"
	AUTHORIZATION_HEADER = "Authorization"
	BEARER_PREFIX        = "Bearer "
)

func main() {
	logFormat := os.Getenv(ENV_KEY_LOG_FORMAT)
	if logFormat == "console" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	addr := os.Getenv(ENV_KEY_SERVER_ADDR)
	if addr == "" {
		addr = ":80"
	}

	log.Info().Str("addr", addr).Msg("starting server")
	r := configureRouter()
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Error().Err(err).Msg("server error")
	}
}

func configureRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		secret := os.Getenv(ENV_KEY_SECRET)
		if secret == "" {
			log.Error().Msg("secret key not configured")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		header := r.Header.Get(AUTHORIZATION_HEADER)
		if header == "" || !strings.HasPrefix(header, BEARER_PREFIX) {
			log.Error().Msg("invalid authorization")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		token := strings.TrimPrefix(header, BEARER_PREFIX)
		if token != secret {
			log.Error().Msg("authorization failed")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		databaseUrl := os.Getenv(ENV_KEY_DATABASE_URL)
		if databaseUrl == "" {
			log.Error().Msg("database url not available")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// w.Write([]byte(databaseUrl))
		w.Write([]byte("hello world"))
	})

	return r
}
