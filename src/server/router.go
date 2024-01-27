package server

import (
	"log/slog"
	"os"
	"runtime"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v2"
	"github.com/unrolled/secure"

	customMiddleware "github.com/MykolaSainiuk/schatgo/src/server/middleware"
	// ginSwagger "github.com/swaggo/gin-swagger"
	// "github.com/swaggo/swag/example/basic/docs"
)

func SetupRouter() chi.Router {
	coresN := runtime.NumCPU()
	slog.Info("Server server... on ", slog.Int("cores", coresN))
	runtime.GOMAXPROCS(coresN)

	// init router
	r := chi.NewRouter()

	isProd := os.Getenv("NODE_ENV") == "production"
	r.Use(secure.New(secure.Options{
		SSLRedirect: isProd,
	}).Handler)
	r.Use(middleware.RequestID)
	// r.Use(middleware.RealIP)

	logger := SetupLogger(os.Getenv("NODE_ENV"))
	r.Use(httplog.RequestLogger(logger))
	// r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Basic CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// hc := customMiddleware.HealthCheck()
	// r.Handle("/health", hc)
	// r.Handle("/health-check", hc)
	// r.Handle("/ping", hc)
	r.Use(middleware.Heartbeat("/health"))
	r.Use(middleware.Heartbeat("/health-check"))
	r.Use(middleware.Heartbeat("/ping"))

	r.Use(middleware.SetHeader("Content-Type", "application/json"))

	r.Handle("/favicon*", customMiddleware.Favicon())

	r.Get("/readme*", customMiddleware.Readme(r))
	r.Get("/swagger/*", customMiddleware.Swagger())

	r.NotFound(customMiddleware.NotFound())

	// // swagger
	// docs.SwaggerInfo.Schemes = []string{"http"}
	// swaggerHandler := ginSwagger.WrapHandler(swaggerFiles.Handler)
	// server.GET("/swagger/*any", swaggerHandler)

	return r
}
