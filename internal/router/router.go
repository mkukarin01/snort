package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mkukarin01/snort/internal/config"
	"github.com/mkukarin01/snort/internal/handlers"
	"github.com/mkukarin01/snort/internal/logger"
	"github.com/mkukarin01/snort/internal/service"
)

// NewRouter - создаем роутер chi
func NewRouter(cfg *config.Config) http.Handler {
	shortener := service.NewURLShortener()
	r := chi.NewRouter()

	// инициализуем собственный логгер синглтончик => мидлварь
	log := logger.InitLogger()
	r.Use(logger.LoggingMiddleware(log))

	// есть какие-то встроенные мидлвари, позовем их
	r.Use(middleware.Recoverer)

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
