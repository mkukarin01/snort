package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/mkukarin01/snort/internal/service"
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

	id := shortener.Shorten(req.URL)
	shortURL := baseURL + "/" + id

	resp := URLResponse{Result: shortURL}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// HandleRedirect - обработчик для GET /{id}
func HandleRedirect(w http.ResponseWriter, r *http.Request, shortener *service.URLShortener) {
	id := chi.URLParam(r, "id")
	if originalURL, ok := shortener.Retrieve(id); ok {
		http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
	} else {
		http.Error(w, "URL not found", http.StatusBadRequest)
	}
}
