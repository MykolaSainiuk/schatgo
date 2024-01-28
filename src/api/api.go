package api

import (
	"github.com/go-chi/chi/v5"

	"github.com/MykolaSainiuk/schatgo/src/middleware"
	"github.com/MykolaSainiuk/schatgo/src/server"

	"github.com/MykolaSainiuk/schatgo/src/api/authapi"
	"github.com/MykolaSainiuk/schatgo/src/api/userapi"
)

func apiRouter(srv *server.Server) chi.Router {
	r := chi.NewRouter()

	AuthOnly := middleware.Authorized(srv.GetDB())

	authHandler := authapi.NewAuthHandler(srv)
	userHandler := userapi.NewUserHandler(srv)

	r.Group(func(r chi.Router) {
		r.Route("/user", func(r chi.Router) {
			r.Post("/register", authHandler.RegisterUser)
			r.Post("/login", authHandler.LoginUser)
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(AuthOnly)
		r.Get("/user/me", userHandler.GetUserInfo)
	})

	return r
}

func InitRoutes(srv *server.Server) {
	r := srv.GetRouter()

	r.Mount("/api", apiRouter(srv))
}
