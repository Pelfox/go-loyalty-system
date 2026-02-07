package internal

// SessionCookieName определяет название для cookie токена сессии.
const SessionCookieName = "session_token"

// UserKey создаёт обозначение для ключа context.Context, который несёт в себе
// информацию об авторизованном и аутентифицированном пользователе.
type UserKey struct{}
