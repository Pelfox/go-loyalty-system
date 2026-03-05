package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

var (
	// ErrInvalidContentType обозначает, что полученное содержимое некорректного
	// типа.
	ErrInvalidContentType = errors.New("invalid content type")
	// ErrInvalidBody обозначает, что полученное тело запроса не является
	// валидным.
	ErrInvalidBody = errors.New("the body is invalid")
)

// GetRequestDecoder возвращает декодер для полученного запроса, проверяя
// заголовок `Content-Type` на правильность значения.
func GetRequestDecoder(r *http.Request) (*json.Decoder, error) {
	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		return nil, ErrInvalidContentType
	}
	return json.NewDecoder(r.Body), nil
}

// DecodeAndValidate декодирует и валидирует полученные данные в запросе.
func DecodeAndValidate(r *http.Request, dst any) error {
	decoder, err := GetRequestDecoder(r)
	if err != nil {
		return err
	}

	if err := decoder.Decode(dst); err != nil {
		return ErrInvalidBody
	}

	if err := validate.Struct(dst); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return fmt.Errorf("validation error: %s", ve[0].Field())
		}
		return err
	}

	return nil
}

// WriteJSON записывает переданное значение структуры в качестве JSON в ответ.
func WriteJSON[T any](w http.ResponseWriter, status int, value T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(value)
}

// RoundTo2Decimals округляет полученное значение до двух знаков после запятой.
func RoundTo2Decimals(value float64) float64 {
	return math.Round(value*100) / 100
}
