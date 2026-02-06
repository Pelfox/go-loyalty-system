package pkg

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestGetRequestDecoder_OK проверяет, что при корректном `Content-Type`
// возвращается JSON-декодер без ошибки.
func TestGetRequestDecoder_OK(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")

	decoder, err := GetRequestDecoder(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if decoder == nil {
		t.Fatal("decoder is nil")
	}
}

// TestGetRequestDecoder_InvalidContentType проверяет, что при некорректном
// `Content-Type` возвращается ErrInvalidContentType.
func TestGetRequestDecoder_InvalidContentType(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "text/plain")

	_, err := GetRequestDecoder(req)
	if !errors.Is(err, ErrInvalidContentType) {
		t.Errorf("error = %v, want %v", err, ErrInvalidContentType)
	}
}

// testRequest используется для тестирования DecodeAndValidate.
type testRequest struct {
	Login string `json:"login" validate:"required"`
}

// TestDecodeAndValidate_OK проверяет, что корректный JSON-запрос успешно
// декодируется и проходит валидацию.
func TestDecodeAndValidate_OK(t *testing.T) {
	body := bytes.NewBufferString(`{"login":"test"}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", "application/json")

	var dst testRequest
	err := DecodeAndValidate(req, &dst)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if dst.Login != "test" {
		t.Errorf("login = %q, want %q", dst.Login, "test")
	}
}

// TestDecodeAndValidate_InvalidBody проверяет, что при некорректном JSON-теле
// запроса возвращается ErrInvalidBody.
func TestDecodeAndValidate_InvalidBody(t *testing.T) {
	body := bytes.NewBufferString(`{`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", "application/json")

	var dst testRequest
	err := DecodeAndValidate(req, &dst)
	if !errors.Is(err, ErrInvalidBody) {
		t.Errorf("error = %v, want %v", err, ErrInvalidBody)
	}
}

// TestDecodeAndValidate_ValidationError проверяет, что при нарушении правил
// валидации возвращается ошибка валидации.
func TestDecodeAndValidate_ValidationError(t *testing.T) {
	body := bytes.NewBufferString(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", "application/json")

	var dst testRequest
	err := DecodeAndValidate(req, &dst)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

// TestWriteJSON_OK проверяет,
// что WriteJSON корректно устанавливает
// статус ответа и заголовок Content-Type.
func TestWriteJSON_OK(t *testing.T) {
	w := httptest.NewRecorder()
	type response struct {
		Status string `json:"status"`
	}

	err := WriteJSON(w, http.StatusCreated, response{Status: "ok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type = %q, want %q", contentType, "application/json")
	}
}
