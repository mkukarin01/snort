package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	ChiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/mkukarin01/snort/internal/config"
	"github.com/mkukarin01/snort/internal/handlers"
	"github.com/mkukarin01/snort/internal/logger"
	InternalMiddleware "github.com/mkukarin01/snort/internal/middleware"
	"github.com/mkukarin01/snort/internal/service"
)

// NewRouter - создаем роутер chi
func NewRouter(cfg *config.Config) http.Handler {
	shortener := service.NewURLShortener()
	r := chi.NewRouter()

	// инициализуем собственный логгер синглтончик => мидлварь
	log := logger.InitLogger()
	r.Use(logger.LoggingMiddleware(log))
	// подсунем gzip реализацию
	r.Use(InternalMiddleware.GzipDecompressionMiddleware) // для входящего трафика
	r.Use(InternalMiddleware.GzipMiddleware)              // для исходящего

	// есть какие-то встроенные мидлвари, позовем их
	r.Use(ChiMiddleware.Recoverer)

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleShorten(w, r, shortener, cfg.BaseURL)
	})

	r.Post("/api/shorten", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleShortenJSON(w, r, shortener, cfg.BaseURL)
	})

	if cfg.BasePath == "" {
		r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
			handlers.HandleRedirect(w, r, shortener)
		})
	} else {
		r.Route(cfg.BasePath, func(r chi.Router) {
			r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
				handlers.HandleRedirect(w, r, shortener)
			})
		})
	}

	return r
}
