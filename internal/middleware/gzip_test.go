package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestGzipMiddleware тестируем сжатие
func TestGzipMiddleware(t *testing.T) {
	// хендлер - просто отдаёт текст
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain") // проверка что не слетает заголовок
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, world!"))
	})

	handler := GzipMiddleware(testHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.Header.Get("Content-Encoding") != "gzip" {
		t.Errorf("Expected 'Content-Encoding: gzip', got %q", resp.Header.Get("Content-Encoding"))
	}

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer gz.Close()

	decodedResponse, err := io.ReadAll(gz)
	if err != nil {
		t.Fatalf("Failed to read gzipped content: %v", err)
	}

	expected := "Hello, world!"
	if string(decodedResponse) != expected {
		t.Errorf("Expected response body %q, got %q", expected, decodedResponse)
	}
}

// TestGzipDecompressionMiddleware тестируем разжатие (sic! разархивирование?)
func TestGzipDecompressionMiddleware(t *testing.T) {
	// хендлер - просто читает текст
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusInternalServerError)
			return
		}
		w.Write(body)
	})

	handler := GzipDecompressionMiddleware(testHandler)

	var gzipBody bytes.Buffer
	gz := gzip.NewWriter(&gzipBody)
	_, err := gz.Write([]byte("Compressed request"))
	if err != nil {
		t.Fatalf("Failed to gzip request body: %v", err)
	}
	gz.Close()

	req := httptest.NewRequest("POST", "/", &gzipBody)
	req.Header.Set("Content-Encoding", "gzip")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	decodedResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	expected := "Compressed request"
	if string(decodedResponse) != expected {
		t.Errorf("Expected response body %q, got %q", expected, decodedResponse)
	}
}

// TestGzipMiddlewareWithoutGzipSupport тестим - трафик без сжатия - зеленый свет
func TestGzipMiddlewareWithoutGzipSupport(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Plain response"))
	})

	handler := GzipMiddleware(testHandler)

	req := httptest.NewRequest("GET", "/", nil) // без AE
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if rr.Header().Get("Content-Encoding") != "" {
		t.Errorf("Expected no 'Content-Encoding' header, got %q", rr.Header().Get("Content-Encoding"))
	}

	body, _ := io.ReadAll(resp.Body)
	expected := "Plain response"
	if string(body) != expected {
		t.Errorf("Expected response body %q, got %q", expected, body)
	}
}

// TestGzipDecompressionMiddlewareEmptyBody - мидлварь + пустое тело запроса
func TestGzipDecompressionMiddlewareEmptyBody(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK) // ничего не делаем, просто 200
	})

	handler := GzipDecompressionMiddleware(testHandler)

	req := httptest.NewRequest("POST", "/", nil) // пустое тело
	req.Header.Set("Content-Encoding", "gzip")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}
}
