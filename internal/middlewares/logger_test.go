package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
)

// TestLoggerMiddleware_OK проверяет, что LoggerMiddleware корректно пропускает
// запрос к следующему обработчику и не изменяет статус-код ответа.
func TestLoggerMiddleware_OK(t *testing.T) {
	logger := zerolog.Nop()
	middleware := LoggerMiddleware(logger)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
	}
}

// TestLoggerMiddleware_ResponseBody проверяет, что LoggerMiddleware не
// вмешивается в тело HTTP-ответа.
func TestLoggerMiddleware_ResponseBody(t *testing.T) {
	logger := zerolog.Nop()
	middleware := LoggerMiddleware(logger)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	body := w.Body.String()
	if body != "ok" {
		t.Errorf("body = %q, want %q", body, "ok")
	}
}
