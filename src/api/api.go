package api

import (
	"github.com/go-chi/chi/v5"

	"github.com/MykolaSainiuk/schatgo/src/common/types"
	"github.com/MykolaSainiuk/schatgo/src/middleware"

	"github.com/MykolaSainiuk/schatgo/src/api/authapi"
	"github.com/MykolaSainiuk/schatgo/src/api/userapi"
	"github.com/MykolaSainiuk/schatgo/src/api/userapi/contactapi"
)

func apiRouter(srv types.IServer) chi.Router {
	r := chi.NewRouter()

	AuthOnly := middleware.Authorized(srv.GetDB())

	r.Group(func(r chi.Router) {
		authHandler := authapi.NewAuthHandler(srv)

		r.Route("/user", func(r chi.Router) {
			r.Post("/register", authHandler.RegisterUser)
			r.Post("/login", authHandler.LoginUser)
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(AuthOnly)
		userHandler := userapi.NewUserHandler(srv)
		r.Get("/user/me", userHandler.GetUserInfo)
	})

	r.Group(func(r chi.Router) {
		r.Use(AuthOnly)
		contactHandler := contactapi.NewContactHandler(srv)
		r.Route("/user/contact", func(r chi.Router) {
			r.Put("/add", contactHandler.AddContact)
			r.Get("/list/all", contactHandler.ListAllContacts)
			r.Get("/list", contactHandler.ListContactsPaginated)
		})
	})

	return r
}

func InitRoutes(srv types.IServer) {
	r := srv.GetRouter()

	r.Mount("/api", apiRouter(srv))
}
