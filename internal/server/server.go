package server

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/contrib/secure"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"geodbsvc/internal/controllers/api"
	"geodbsvc/internal/database"
	"geodbsvc/internal/embedfs"
	"geodbsvc/internal/loggerx"
	"geodbsvc/web"
)

type Server struct {
	cfg     *viper.Viper
	db      database.DB
	httpSrv *http.Server
	apiCtrl *api.Controller
}

// New
func New(cfg *viper.Viper, db database.DB, logger *loggerx.Logger) *Server {
	srv := new(Server)
	srv.cfg = cfg
	srv.db = db

	// Создаём HTTP-сервер
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = logger

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// Некоторые настройки, связанные с безопасностью
	router.Use(secure.Secure(secure.Options{
		FrameDeny:          true, // Запрещает показывать сайт во фрейме
		ContentTypeNosniff: true, //
		BrowserXssFilter:   true, //
	}))

	// Шаблоны
	templ := template.Must(template.New("").ParseFS(web.FS, "templates/*.gohtml"))
	router.SetHTMLTemplate(templ)

	router.GET("/", func(ctx *gin.Context) {
		ip := ctx.ClientIP()
		geo, err := srv.db.Lookup(ip)
		if err != nil {
			geo = &database.GeoRecord{}
		}
		ctx.HTML(http.StatusOK, "index.gohtml", gin.H{"ip": ip, "geo": geo})
	})

	// Статические файлы
	router.Use(static.Serve("/", embedfs.EmbedFolder(web.FS, "static")))

	// API
	srv.apiCtrl = api.NewController(srv)

	api := router.Group("/api")
	{
		api.GET("/lookup/:ip", srv.apiCtrl.Lookup())
	}

	srv.httpSrv = &http.Server{
		Handler:      router,
		WriteTimeout: 5 * time.Minute, // Таймаут ответа от сервера
	}

	return srv
}

// Run запускает HTTP-сервер
func (s *Server) Run(addr string) {
	s.httpSrv.Addr = addr
	go func() {
		if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error: %s\n", err)
		}
	}()
}

// Shutdown останавливает HTTP-сервер
func (s *Server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.httpSrv.Shutdown(ctx)
}

// GetConfig возвращает указатель на конфиг
func (s *Server) GetConfig() *viper.Viper {
	return s.cfg
}

// GetDB возвращает указатель на БД
func (s *Server) GetDB() database.DB {
	return s.db
}
