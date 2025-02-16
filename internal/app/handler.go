package app

import (
    "io"
    "math/rand"
    "net/http"
    "sync" // читать тут https://pkg.go.dev/sync
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

// Server структура данных для хранения URLShortener
type Server struct {
    shortener *URLShortener
}

// NewServer просто создадим новый сервер
func NewServer() *Server {
    return &Server{
        shortener: NewURLShortener(),
    }
}

// ServeHTTP просто свитч, вместо варей
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    		case http.MethodPost:
        		s.handlePost(w, r)
    		case http.MethodGet:
        		s.handleGet(w, r)
    		default:
        		http.Error(w, "Unsupported request method", http.StatusBadRequest)
    }
}

// handlePost - хендлер пост запросов
func (s *Server) handlePost(w http.ResponseWriter, r *http.Request) {
    url, err := io.ReadAll(r.Body)
    if err != nil || len(url) == 0 {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    id := s.shortener.Shorten(string(url))
    shortURL := "http://localhost:8080/" + id
    w.WriteHeader(http.StatusCreated)
    w.Write([]byte(shortURL))
}

// handleGet - хендлер гет запросов
func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Path[1:]
    if originalURL, ok := s.shortener.Retrieve(id); ok {
        http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
    } else {
        http.Error(w, "URL not found", http.StatusBadRequest)
    }
}
