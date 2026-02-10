package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Pelfox/go-loyalty-system/internal/constants"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TestAuthMiddleware_NoCookie проверяет, что при отсутствии cookie
// аутентификации middleware возвращает `HTTP 401 Unauthorized`.
func TestAuthMiddleware_NoCookie(t *testing.T) {
	middleware := AuthMiddleware([]byte("secret"))
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler must not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// TestAuthMiddleware_InvalidToken проверяет, что при некорректном JWT-токене
// middleware возвращает `HTTP 401 Unauthorized`.
func TestAuthMiddleware_InvalidToken(t *testing.T) {
	middleware := AuthMiddleware([]byte("secret"))
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler must not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  constants.SessionCookieName,
		Value: "invalid-token",
	})

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// TestAuthMiddleware_InvalidUserID проверяет, что при некорректном
// идентификаторе пользователя в `Subject` JWT-токена middleware возвращает
// `HTTP 401 Unauthorized`.
func TestAuthMiddleware_InvalidUserID(t *testing.T) {
	secret := []byte("secret")
	claims := jwt.RegisteredClaims{
		Subject: "not-a-uuid",
		ExpiresAt: jwt.NewNumericDate(
			time.Now().Add(time.Minute),
		),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	middleware := AuthMiddleware(secret)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler must not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  constants.SessionCookieName,
		Value: signedToken,
	})

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// TestAuthMiddleware_OK проверяет успешную аутентификацию пользователя. В
// случае валидного JWT-токена middleware должен: вызвать следующий обработчик
// и добавить `userID` в context.Context запроса.
func TestAuthMiddleware_OK(t *testing.T) {
	secret := []byte("secret")
	userID := uuid.New()

	claims := jwt.RegisteredClaims{
		Subject: userID.String(),
		ExpiresAt: jwt.NewNumericDate(
			time.Now().Add(time.Minute),
		),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	middleware := AuthMiddleware(secret)

	called := false
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true

		ctxUserID, ok := r.Context().Value(constants.UserKey{}).(uuid.UUID)
		if !ok {
			t.Fatalf("userID not found in context")
		}
		if ctxUserID != userID {
			t.Errorf("userID = %v, want %v", ctxUserID, userID)
		}

		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  constants.SessionCookieName,
		Value: signedToken,
	})

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if !called {
		t.Fatal("next handler was not called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}
