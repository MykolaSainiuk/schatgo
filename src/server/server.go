package server

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"

	"github.com/MykolaSainiuk/schatgo/src/common/types"
	"github.com/MykolaSainiuk/schatgo/src/db"
	"github.com/MykolaSainiuk/schatgo/src/helper/jwthelper"
	"github.com/MykolaSainiuk/schatgo/src/server/router"
)

type Server struct {
	router chi.Router
	db     types.IDatabase
}

func Setup() types.IServer {
	r := router.SetupRouter()

	dbConn, err := db.ConnectDB(os.Getenv("MONGO_INITDB_DATABASE"))
	if err != nil {
		slog.Error("failed to connect MongoDB")
		os.Exit(1)
	}

	return &Server{
		router: r,
		db:     dbConn,
	}
}

func (srv *Server) Run() <-chan struct{} {
	closingCh := make(chan struct{}, 1)
	host, port := os.Getenv("HOST"), os.Getenv("PORT")

	// rest
	go func() {
		slog.Info("server is up on port", slog.String("port", port))

		err := http.ListenAndServe(host+":"+port, srv.router)
		if err != nil {
			slog.Error("server has failed to serve", slog.Any("error", err.Error()))
		}

		// reflect.ValueOf(ch).TrySend(reflect.ValueOf(struct{}{}))
		closingCh <- struct{}{}
	}()

	return closingCh
}

func (srv *Server) Shutdown() {
	slog.Info("Closing server gracefully")
	srv.db.Shutdown()
}

func (srv *Server) GetRouter() chi.Router {
	return srv.router
}

func (srv *Server) GetDB() types.IDatabase {
	return srv.db
}

func init() {
	// evn vars load
	envFilePath := getEnvFilePath()
	if err := godotenv.Load(envFilePath); err != nil {
		slog.Error(".env file not found")
		os.Exit(1)
	}
	slog.Info(".env file is loaded")

	// load jwt secret & expr
	jwthelper.InitJwtData()
}

func getEnvFilePath() string {
	_, b, _, _ := runtime.Caller(0)
	rootPath := filepath.Dir(b)
	return filepath.Join(rootPath, "../../", "./.env")
}
