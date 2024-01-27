package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func InitRoutes(r chi.Router) {
	// test route
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello world!"))
	})
}
