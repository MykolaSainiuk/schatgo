package server

import (
	"log/slog"
	"os"

	"github.com/go-chi/httplog/v2"
)

func SetupLogger(env string) *httplog.Logger {
	isProd := os.Getenv("NODE_ENV") == "production"
	var tags map[string]string = nil
	if isProd {
		tags = map[string]string{
			"version": "v1.0",
			"env":     env,
		}
	}
	return httplog.NewLogger("schat-server", httplog.Options{
		JSON:             isProd,
		LogLevel:         slog.LevelDebug,
		Concise:          !isProd,
		RequestHeaders:   isProd,
		MessageFieldName: "message",
		// TimeFieldFormat: time.RFC850,
		Tags: tags,
		QuietDownRoutes: []string{
			"/",
			"/favicon.ico",
			"/favicon",
			"/health",
			"/health-check",
		},
		// QuietDownPeriod: 10 * time.Second,
		// SourceFieldName: "source",
	})
}

var LogPathsToSkip = []string{
	"/",
	"/favicon.ico",
	"/favicon",
	"/health",
	"/health-check",
	"/ping",
	"/swagger/index.html",
	"/swagger/swagger-ui.css",
	"/swagger/swagger-ui-standalone-preset.js",
	"/swagger/swagger-ui-bundle.js",
	"/swagger/favicon-32x32.png",
	"/swagger/doc.json",
}
