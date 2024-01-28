package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/MykolaSainiuk/schatgo/src/common/types"
	"github.com/MykolaSainiuk/schatgo/src/helper/jwthelper"
)

func Authorized(dbRef types.IDatabase) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var accessToken string
			if accessToken = getAuthHeader(r); accessToken == "" {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			var payload *types.TokenPayload
			var err error
			if payload, err = jwthelper.VerifyToken(accessToken); err != nil {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			// TOD: go to DB and check

			// r.Header.Set("UserId", payload.UserID)
			ctx := context.WithValue(r.Context(), types.TokenPayload{}, payload)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func getAuthHeader(r *http.Request) string {
	accessToken := r.Header.Get("Authorization")
	if accessToken == "" {
		return ""
	}

	splitToken := strings.Split(accessToken, "Bearer ")
	if len(splitToken) == 2 {
		return splitToken[1]
	}
	return ""
}
