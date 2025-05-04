package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"sf-finances/src/services"
)

type contextKey string
const UserIDKey contextKey = "userID"

type JWTMiddleware struct {
	authService services.UserService
	logger      *logrus.Logger
}

func NewJWTMiddleware(authService services.UserService, logger *logrus.Logger) *JWTMiddleware {
	return &JWTMiddleware{
		authService: authService,
		logger:      logger,
	}
}

func (m *JWTMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Нужна авторизация", http.StatusUnauthorized)
			return
		}

		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			http.Error(w, "Неверный токен", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, bearerPrefix)

		userID, err := m.authService.ParseToken(tokenString)
		if err != nil {
			m.logger.WithError(err).Warn("Ошибка при проверке токена")
			http.Error(w, "Неверный токен", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(ctx context.Context) (int64, error) {
	userID := ctx.Value(UserIDKey).(int64)
	return userID, nil
}

