package middlewares

import (
	"context"
	"net/http"

	"github.com/Pelfox/go-loyalty-system/internal"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// UserKey создаёт обозначение для ключа context.Context, который несёт в себе
// информацию об авторизованном и аутентифицированном пользователе.
type UserKey struct{}

// AuthMiddleware реализует необходимые проверки авторизации/аутентификации.
func AuthMiddleware(jwtSecret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenCookie, err := r.Cookie(internal.SessionCookieName)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// парсим токен и пытаемся его дешифровать через jwtSecret
			token, err := jwt.ParseWithClaims(
				tokenCookie.Value,
				&jwt.RegisteredClaims{},
				func(token *jwt.Token) (any, error) {
					return jwtSecret, nil
				},
				jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
			)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// проверяем что токен вообще валидный
			if !token.Valid {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// парсим ID пользователя, и добавляем его в контекст запроса
			claims := token.Claims.(*jwt.RegisteredClaims)
			userID, err := uuid.Parse(claims.Subject)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserKey{}, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
