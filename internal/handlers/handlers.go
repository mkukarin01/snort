package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/mkukarin01/snort/internal/middleware"
	"github.com/mkukarin01/snort/internal/service"
	"github.com/mkukarin01/snort/internal/storage"
)

// URLRequest - структура запроса для JSON-POST
type URLRequest struct {
	URL string `json:"url"`
}

// URLResponse - структура ответа для JSON-POST
type URLResponse struct {
	Result string `json:"result"`
}

// HandleShorten - обработчик для POST /
func HandleShorten(w http.ResponseWriter, r *http.Request, shortener *service.URLShortener, baseURL string) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil || len(bodyBytes) == 0 {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

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

	// достанем userID из контекста, если он там есть
	userID := middleware.GetUserIDFromContext(r.Context())

	id, conflict := shortener.Shorten(urlStr, userID)
	shortURL := baseURL + "/" + id

	if conflict {
		// url есть в бд - отдаем 409 Conflict
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(shortURL))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

// HandleShortenJSON - обработчик для POST /api/shorten
func HandleShortenJSON(w http.ResponseWriter, r *http.Request, shortener *service.URLShortener, baseURL string) {
	var req URLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	parsedURL, err := url.ParseRequestURI(req.URL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserIDFromContext(r.Context())

	id, conflict := shortener.Shorten(req.URL, userID)
	shortURL := baseURL + "/" + id

	resp := URLResponse{Result: shortURL}

	w.Header().Set("Content-Type", "application/json")
	if conflict {
		// url есть в бд - отдаем 409 Conflict
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(resp)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// BatchRequest - структурка запроса
type BatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// BatchResponse - структурка ответа
type BatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// HandleShortenBatch - обработчик для POST /api/shorten/batch
func HandleShortenBatch(w http.ResponseWriter, r *http.Request, shortener *service.URLShortener, baseURL string) {
	var req []BatchRequest
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil || len(bodyBytes) == 0 {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// выкинем ошибку если пустой или невалидный
	if err := json.Unmarshal(bodyBytes, &req); err != nil || len(req) == 0 {
		http.Error(w, "Invalid JSON or empty batch", http.StatusBadRequest)
		return
	}

	// каждая ссылка в батче была валидной
	urls := make(map[string]string)
	for _, item := range req {
		parsedURL, err := url.ParseRequestURI(item.OriginalURL)
		if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
			http.Error(w, "Invalid URL in batch", http.StatusBadRequest)
			return
		}
		urls[item.CorrelationID] = item.OriginalURL
	}

	userID := middleware.GetUserIDFromContext(r.Context())
	shortened := shortener.ShortenBatch(urls, userID)

	var res []BatchResponse
	for correlationID, shortID := range shortened {
		res = append(res, BatchResponse{
			CorrelationID: correlationID,
			ShortURL:      baseURL + "/" + shortID,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

// HandleRedirect - обработчик для GET /{id}
func HandleRedirect(w http.ResponseWriter, r *http.Request, shortener *service.URLShortener) {
	id := chi.URLParam(r, "id")
	originalURL, err := shortener.Retrieve(id)
	if err != nil {
		// Если получаем ошибку "удалено", возвращаем 410
		if errors.Is(err, storage.ErrURLDeleted) {
			http.Error(w, "URL is deleted", http.StatusGone)
			return
		}
		// Если "не найдено" или любая другая - 400
		http.Error(w, "URL not found", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}

// HandlePing - обработчик для GET /ping
func HandlePing(w http.ResponseWriter, r *http.Request, db storage.Storager) {
	// db не инициализируем - 500 и работаем дальше
	if db == nil {
		http.Error(w, "Database connection is not configured", http.StatusInternalServerError)
		return
	}

	// db не пингуется - 500 и все равно работаем дальше
	if err := db.Ping(); err != nil {
		http.Error(w, "Database connection failed", http.StatusInternalServerError)
		return
	}

	// 200
	w.WriteHeader(http.StatusOK)
}

// HandleUserURLs - обработчик для GET /api/user/urls
func HandleUserURLs(w http.ResponseWriter, r *http.Request, shortener *service.URLShortener, baseURL string) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userURLs, err := shortener.UserURLs(userID, baseURL)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if len(userURLs) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userURLs)
}

// HandleDeleteUserURLs - обработчик для DELETE /api/user/urls
func HandleDeleteUserURLs(w http.ResponseWriter, r *http.Request, deleter *service.URLDeleter) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var shortIDs []string
	if err := json.NewDecoder(r.Body).Decode(&shortIDs); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Вызываем асинхронное удаление (fanIn)
	deleter.Submit(userID, shortIDs)

	// Возвращаем 202 Accepted
	w.WriteHeader(http.StatusAccepted)
}
