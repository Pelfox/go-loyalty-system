package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

var (
	// ErrNotRegistered возвращается в том случае, если заказ с указанным
	// номером ещё не был зарегистрирован во внешней системе обработки.
	ErrNotRegistered = errors.New("order with this number is not registered yet")
)

// RateLimitError описывает ошибку, возвращаемую в случае ограничений по
// времени от внешней системы вознаграждений при попытке запроса (rate limit).
type RateLimitError struct {
	// RetryAfter - время (длительность), через которое можно повторить запрос.
	RetryAfter time.Duration
}

func (r RateLimitError) Error() string {
	return "rate limit exceeded"
}

// rateLimitErrorFromHeader создаёт и возвращает новый объект RateLimitError с
// временем повтора (retry after), полученным из заголовка ответа сервера.
func rateLimitErrorFromHeader(header string) RateLimitError {
	value := time.Second
	if seconds, err := strconv.Atoi(header); err == nil {
		value = time.Duration(seconds) * time.Second
	}
	return RateLimitError{RetryAfter: value}
}

// Client это функциональная абстракция для работы с внешней системой учёта
// вознаграждений.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient создаёт и возвращает новый объект Client, настраивая внутренний
// клиент http.Client для работы с внешней системой учёта вознаграждений.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GetOrder возвращает информацию о заказе с указанным номером из внешней
// системы учёта. В случае, если заказ ещё не был зарегистрирован, вернёт
// ErrNotRegistered. Если внешняя система ограничила запрос из-за rate-limit,
// вернёт RateLimitError с временем, через которое необходимо повторить запрос.
func (c *Client) GetOrder(ctx context.Context, number string) (*OrderResponse, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/api/orders/%s", c.baseURL, number),
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var response OrderResponse
		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(&response); err != nil {
			return nil, err
		}
		return &response, nil
	case http.StatusNoContent:
		return nil, ErrNotRegistered
	case http.StatusTooManyRequests:
		retryAfter := resp.Header.Get("Retry-After")
		return nil, rateLimitErrorFromHeader(retryAfter)
	default:
		return nil, fmt.Errorf("accrual service returned %d", resp.StatusCode)
	}
}
