package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mkukarin01/snort/internal/config"
	"github.com/mkukarin01/snort/internal/service"
	"github.com/mkukarin01/snort/internal/storage"
)

// **** utility ****//

// createCfg просто конфиг вернем
func createCfg() *config.Config {
	cfg := &config.Config{
		Port:       "8080",
		BaseDomain: "localhost",
		BasePath:   "",
		Address:    "localhost:8080",
		BaseURL:    "http://localhost:8080",
	}
	return cfg
}

// **** cases ****//

// TestRouter_Routes проверяет, что роутер создаётся и все маршруты присутствуют.
func TestRouter_Routes(t *testing.T) {
	cfg := createCfg()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// так-то теперь можно не мокать, а напрямую дергать memoryStorage и не париться
	// мокаем стораджер
	mockDB := storage.NewMockStorager(ctrl)
	mockDB.EXPECT().Ping().Return(nil)
	mockDB.EXPECT().Load("anyShortID").Return("http://ya.ru", nil)
	mockDB.EXPECT().GetUserURLs(gomock.Any()).Return([]storage.UserURL{}, nil)
	// fanin
	deleter := service.NewURLDeleter(mockDB)

	r := NewRouter(cfg, mockDB, deleter)

	testCases := []struct {
		method string
		path   string
	}{
		{"POST", "/"},
		{"POST", "/api/shorten"},
		{"POST", "/api/shorten/batch"},
		{"GET", "/api/user/urls"},
		{"GET", "/anyShortID"},
		{"GET", "/ping"},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()
		require.NotEqual(t, http.StatusNotFound, res.StatusCode, "Expected route to be registered: %s %s", tc.method, tc.path)
	}
}

// TestRouter_Ping_Success - тест роутера /ping 200
func TestRouter_Ping_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// мокаем стораджер
	mockDB := storage.NewMockStorager(ctrl)
	mockDB.EXPECT().Ping().Return(nil)

	cfg := &config.Config{
		Address: "localhost:8080",
	}

	// fanin
	deleter := service.NewURLDeleter(mockDB)

	router := NewRouter(cfg, mockDB, deleter)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

// TestRouter_Ping_Failed - тест роутера /ping 500
func TestRouter_Ping_Failed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := storage.NewMockStorager(ctrl)
	mockDB.EXPECT().Ping().Return(assert.AnError)

	cfg := &config.Config{
		Address: "localhost:8080",
	}

	// fanin
	deleter := service.NewURLDeleter(mockDB)

	router := NewRouter(cfg, mockDB, deleter)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}

// TestRouter_Ping_NoDB - тест роутера без базы данных (nil)
func TestRouter_Ping_NoDB(t *testing.T) {
	cfg := &config.Config{
		Address: "localhost:8080",
	}

	// fanin
	deleter := service.NewURLDeleter(nil)

	router := NewRouter(cfg, nil, deleter)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}
