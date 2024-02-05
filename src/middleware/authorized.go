package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"

	"github.com/MykolaSainiuk/schatgo/src/common/cmnerr"
	"github.com/MykolaSainiuk/schatgo/src/common/types"
	"github.com/MykolaSainiuk/schatgo/src/helper/jwthelper"

	"github.com/MykolaSainiuk/schatgo/src/repo/tokenrepo"
)

var (
	trOnce sync.Once
	repo   *tokenrepo.TokenRepo = nil
)

func Authorized(dbRef types.IDatabase) func(http.Handler) http.Handler {
	trOnce.Do(func() {
		repo = tokenrepo.NewTokenRepo(dbRef)
	})

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if repo == nil {
				cmnerr.Reply500(w, ErrTokenRepoNotInitialized)
				return
			}

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

			ok, err := repo.ExistToken(r.Context(), &accessToken)
			if err != nil || !ok {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

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

var (
	ErrTokenRepoNotInitialized = errors.New("toke repo not initialized")
)
