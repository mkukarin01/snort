package app

import (
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"sync" // читать тут https://pkg.go.dev/sync

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mkukarin01/snort/internal/config"
	"github.com/mkukarin01/snort/internal/logger"
)

// URLShortener тип данных сопоставления данных id - ссылка
type URLShortener struct {
	// https://pkg.go.dev/sync#RWMutex
	// хочу чтобы можно было кем угодно читать, но писать одному пока, создал и переиспользуешь на протяжении работы аппы
	sync.RWMutex
	store map[string]string
}

// NewURLShortener создаёт новый экземпляр URLShortener, сделал чтобы меньше писать кода
// вдруг по каким-то причинам захочется разделить потоки данных
func NewURLShortener() *URLShortener {
	return &URLShortener{
		store: make(map[string]string),
	}
}

// Shorten функция для коротких идентификаторов, пока без привязки к урлу
func (us *URLShortener) Shorten(url string) string {
	id := generateID()
	us.Lock()
	us.store[id] = url
	us.Unlock()
	return id
}

// Retrieve юзаем стор чтобы вытащить данные по идентификатору и возвращаем + ок
func (us *URLShortener) Retrieve(id string) (string, bool) {
	us.RLock()
	url, ok := us.store[id]
	us.RUnlock()
	return url, ok
}

// generateID рандомный идентификатор, написал тупую функцию
func generateID() string {
	const length = 8
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// слайс байтов в длину length === 8
	id := make([]byte, length)
	for i := range id {
		// берем случайный индекс
		randomIndex := rand.Intn(len(charset))
		// для позиции i ставим символ из charset
		id[i] = charset[randomIndex]
	}
	return string(id)
}

// NewRouter - создаем роутер chi
func NewRouter(cfg *config.Config) http.Handler {
	shortener := NewURLShortener()
	r := chi.NewRouter()

	// инициализуем собственный логгер синглтончик => мидлварь
	log := logger.InitLogger()
	r.Use(LoggingMiddleware(log))

	// есть какие-то встроенные мидлвари, позовем их
	r.Use(middleware.Recoverer)

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		handlePost(w, r, shortener, cfg.BaseURL)
	})

	// работаем с префиксом, если он есть
	if cfg.BasePath == "" {
		r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
			handleGet(w, r, shortener)
		})
	} else {
		r.Route(cfg.BasePath, func(r chi.Router) {
			r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
				handleGet(w, r, shortener)
			})
		})
	}

	return r
}

// handlePost - просто функция, для пост запроса
func handlePost(w http.ResponseWriter, r *http.Request, shortener *URLShortener, baseURL string) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil || len(bodyBytes) == 0 {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// создам переменную, чтобы потом к ней обращаться несколько раз
	urlStr := string(bodyBytes)

	parsedURL, err := url.ParseRequestURI(urlStr)
	if err != nil {
		http.Error(w, "Invalid URL provided", http.StatusBadRequest)
		return
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		http.Error(w, "Invalid URL: missing scheme or host", http.StatusBadRequest)
		return
	}

	id := shortener.Shorten(urlStr)
	shortURL := baseURL + "/" + id

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

// handlePost - просто функция, для гет запроса
func handleGet(w http.ResponseWriter, r *http.Request, shortener *URLShortener) {
	id := chi.URLParam(r, "id")
	if originalURL, ok := shortener.Retrieve(id); ok {
		http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
	} else {
		http.Error(w, "URL not found", http.StatusBadRequest)
	}
}
