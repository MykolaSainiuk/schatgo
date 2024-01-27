package server

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/unrolled/secure"

	// ginSwagger "github.com/swaggo/gin-swagger"
	// "github.com/swaggo/swag/example/basic/docs"

	db "github.com/MykolaSainiuk/schatgo/src/db"
)

// evn vars first
func init() {
	envFilePath := getEnvFilePath()
	slog.Debug("envFilePath", envFilePath)
	if err := godotenv.Load(envFilePath); err != nil {
		slog.Error(".env file not found")
		os.Exit(1)
	}
	slog.Info(".env file is loaded...")

	// load jwt secret & expr
	// jwthelper.InitJwtData()
}

type Server struct {
	router chi.Router
	db     *db.DB
}

func (srv *Server) Router() chi.Router {
	return srv.router
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

func SetupRouter() chi.Router {
	coresN := runtime.NumCPU()
	slog.Info("Server server... on ", coresN, "cores")
	runtime.GOMAXPROCS(coresN)

	isProd := os.Getenv("NODE_ENV") == "production"

	// init router
	r := chi.NewRouter()

	r.Use(secure.New(secure.Options{
		SSLRedirect: isProd,
	}).Handler)
	r.Use(middleware.RequestID)
	// r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
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

	// mode := gin.Mode()
	// if mode != "" || mode == gin.ReleaseMode {
	// 	// default logger - better for debug & development
	// 	server.Use(gin.LoggerWithConfig(gin.LoggerConfig{
	// 		SkipPaths: swaggerPaths,
	// 	}))
	// } else {
	// 	// custom JSON logging, prod-like
	// 	server.Use(middleware.JSONLogger())
	// }

	// // avoid dumb warn msg
	// _ = server.SetTrustedProxies([]string{os.Getenv("HOST")})
	// // custom 500s
	// server.Use(gin.CustomRecovery(middleware.PanicHandler))
	// // set basic security headers
	// server.Use(helmet.Default())
	// // tracking
	// server.Use(middleware.RequestID())
	// // enable cors
	// server.Use(middleware.CORS(middleware.CORSOptions{Origin: "*"}))

	// // tech routes
	// server.GET("/favicon.ico", middleware.Favicon())
	// server.HEAD("/favicon.ico", middleware.Favicon())

	// // health endpoints
	// server.GET("/health", middleware.HealthCheck())
	// server.GET("/health-check", middleware.HealthCheck())

	// // swagger
	// docs.SwaggerInfo.Schemes = []string{"http"}
	// swaggerHandler := ginSwagger.WrapHandler(swaggerFiles.Handler)
	// server.GET("/swagger/*any", swaggerHandler)

	return r
}

func (srv *Server) Run() <-chan struct{} {
	closingCh := make(chan struct{}, 1)
	// host := os.Getenv("HOST")
	port := os.Getenv("PORT")

	// rest
	go func(ch chan<- struct{}) {
		slog.Info("Server is up on port", os.Getenv("PORT"))

		err := http.ListenAndServe(":"+port, srv.router)
		if err != nil {
			slog.Error("Server has failed to serve", slog.Any("error:", err.Error()))
		}

		// reflect.ValueOf(ch).TrySend(reflect.ValueOf(struct{}{}))
		ch <- struct{}{}
	}(closingCh)

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

func getEnvFilePath() string {
	_, b, _, _ := runtime.Caller(0)
	rootPath := filepath.Dir(b)
	return filepath.Join(rootPath, "../../", "./.env")
}
