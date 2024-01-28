package middleware

import "net/http"

func Favicon() http.HandlerFunc {
	return NotFound()
}
