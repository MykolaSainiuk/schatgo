package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/MykolaSainiuk/schatgo/src/api"
	"github.com/MykolaSainiuk/schatgo/src/server"
)

//	@title			sChat Server API
//	@version		1.0
//	@description	sChat REST server

//	@contact.name	mykola.sainyuk@gmail.com
//	@contact.url	https://www.linkedin.com/in/mykola-sainiuk-3b03168b/

//	@host		localhost:5000
//	@BasePath	/

// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
func main() {
	srv := server.Setup()
	defer srv.Shutdown()

	api.InitRoutes(srv)

	stoppedServerCh := srv.Run()

	signalsCh := make(chan os.Signal, 1)
	signal.Notify(signalsCh, os.Interrupt, syscall.SIGTERM)

	sig := os.Signal(nil)
	select {
	case <-stoppedServerCh:
		slog.Info("Server closed somehow")
	case sig = <-signalsCh:
		slog.Info("SIGTERM signal caught", slog.String("signal", sig.String()))
	}

	if sig != nil {
		// os.Exit(sig)
		os.Exit(1)
	}
	os.Exit(0)
}
