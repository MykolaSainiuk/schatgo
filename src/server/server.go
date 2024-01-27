package server

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"

	db "github.com/MykolaSainiuk/schatgo/src/db"
)

type Server struct {
	router chi.Router
	db     *db.DB
}

func Setup() *Server {
	router := SetupRouter()

	dbConn, err := db.ConnectDB(os.Getenv("MONGO_INITDB_DATABASE"))
	if err != nil {
		slog.Error("Failed to connect MongoDB")
		os.Exit(1)
	}

	return &Server{
		router: router,
		db:     dbConn,
	}
}

func (srv *Server) Router() chi.Router {
	return srv.router
}

func (srv *Server) Run() <-chan struct{} {
	closingCh := make(chan struct{}, 1)
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")

	// rest
	go func() {
		slog.Info("Server is up on port", slog.String("port", port))

		err := http.ListenAndServe(host+":"+port, srv.router)
		if err != nil {
			slog.Error("Server has failed to serve", slog.Any("error", err.Error()))
		}

		// reflect.ValueOf(ch).TrySend(reflect.ValueOf(struct{}{}))
		closingCh <- struct{}{}
	}()

	return closingCh
}

func (srv *Server) Shutdown() {
	slog.Info("Closing server gracefully...")

	// drop DB connections
	if srv.db.CancelFn != nil {
		srv.db.CancelFn()
	}
	if srv.db.Client != nil && srv.db.Context != nil {
		if err := srv.db.Client.Disconnect(srv.db.Context); err != nil {
			slog.Error(err.Error())
		}
	}
}

// evn vars first
func init() {
	envFilePath := getEnvFilePath()
	if err := godotenv.Load(envFilePath); err != nil {
		slog.Error(".env file not found")
		os.Exit(1)
	}
	slog.Info(".env file is loaded...")

	// load jwt secret & expr
	// jwthelper.InitJwtData()
}

func getEnvFilePath() string {
	_, b, _, _ := runtime.Caller(0)
	rootPath := filepath.Dir(b)
	return filepath.Join(rootPath, "../../", "./.env")
}
