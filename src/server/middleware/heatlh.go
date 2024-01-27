package middleware

import (
	"encoding/json"
	"net/http"
)

func HealthCheck() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(hResponse)
	})
}

var healthyObj = struct {
	Server string `json:"server"`
}{
	Server: "OK",
}
var hResponse, _ = json.Marshal(healthyObj)
