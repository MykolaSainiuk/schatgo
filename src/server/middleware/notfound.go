package middleware

import "net/http"

func NotFound() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})
}
