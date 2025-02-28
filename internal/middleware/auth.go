package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/mkukarin01/snort/internal/config"
	"github.com/rs/xid"
)

const (
	jwtCookieName   = "SNORT_AUTH"
	defaultIssuer   = "snort-service"
	defaultAudience = "snort-users"
	// конечно хотелось бы так, но staticcheck вернет ошибку
	// contextUserKey = "snortUserID"
)

type contextKey string

var contextUserKey = contextKey("snortUserID")

// UserAuthMiddleware - мидлварь для проверки jwt
func UserAuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(jwtCookieName)
			var userID string

			// валидация
			if err == nil && cookie != nil && cookie.Value != "" {
				uID, valErr := ValidateJWT(cookie.Value, cfg.SecretKey)
				if valErr == nil && uID != "" {
					// ок
					userID = uID
				}
			}

			// если ничего не вытащили
			if userID == "" {
				userID = xid.New().String()
				newToken, err := GenerateJWT(userID, defaultIssuer, defaultAudience, cfg.SecretKey)
				if err == nil {
					setJWTTokenCookie(w, newToken)
				}
			}

			// контекст нашлепнул
			ctx := context.WithValue(r.Context(), contextUserKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// setJWTTokenCookie - ставит на серверный ответ Set-Cookie
func setJWTTokenCookie(w http.ResponseWriter, jwtToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     jwtCookieName,
		Value:    jwtToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // https => true
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().AddDate(0, 0, 7),
	})
}

// GetUserIDFromContext - достаем uID из контекста
func GetUserIDFromContext(ctx context.Context) string {
	val, _ := ctx.Value(contextUserKey).(string)
	return val
}
