package api

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/MykolaSainiuk/schatgo/src/common/types"
	"github.com/MykolaSainiuk/schatgo/src/middleware"

	"github.com/MykolaSainiuk/schatgo/src/api/authapi"
	"github.com/MykolaSainiuk/schatgo/src/api/chatapi"
	"github.com/MykolaSainiuk/schatgo/src/api/chatapi/messageapi"
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

	r.Group(func(r chi.Router) {
		r.Use(AuthOnly)
		chatHandler := chatapi.NewChatHandler(srv)
		r.Route("/chat", func(r chi.Router) {
			r.Put("/new", chatHandler.NewChat)
			r.Get("/list/all", chatHandler.ListAllChats)
			r.Get("/list", chatHandler.ListChatsPaginated)
			r.Delete("/{chatId}/clear", chatHandler.ClearChat)
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(AuthOnly)
		msgHandler := messageapi.NewMessageHandler(srv)
		r.Route("/message", func(r chi.Router) {
			r.Put("/{chatId}/new", msgHandler.NewMessage)
			r.Get("/{chatId}/list/all", msgHandler.ListAllMessages)
			r.Get("/{chatId}/list", msgHandler.ListMessagesPaginated)
		})
	})

	return r
}

func InitRoutes(srv types.IServer) {
	r := srv.GetRouter()

	r.Mount("/api", apiRouter(srv))

	if os.Getenv("NODE_ENV") != "production" {
		// 👇 the walking function 🚶‍♂️ to print routes
		chi.Walk(r, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
			fmt.Printf("[%s]: '%s' has %d middlewares\n", method, route, len(middlewares))
			return nil
		})
	}
}
