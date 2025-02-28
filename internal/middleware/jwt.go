package middleware

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// JWTClaims берем просто существующие claims
type JWTClaims struct {
	jwt.RegisteredClaims
}

// GenerateJWT генерирует токен
func GenerateJWT(userID, issuer, audience, secret string) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,                                   // sub
			Issuer:    issuer,                                   // iss
			Audience:  []string{audience},                       // aud
			ExpiresAt: jwt.NewNumericDate(now.AddDate(0, 0, 7)), // exp
			IssuedAt:  jwt.NewNumericDate(now),                  // iat
			NotBefore: jwt.NewNumericDate(now),                  // nbf
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateJWT парсит, проверяет подпись и валидность токена, вернет Subject aka uID
func ValidateJWT(tokenString, secret string) (string, error) {
	if tokenString == "" {
		return "", errors.New("empty token string")
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(t *jwt.Token) (any, error) {
		// подпись проверили
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return "", errors.New("invalid token claims")
	}

	// aud
	if !claims.VerifyAudience(defaultAudience, true) {
		return "", fmt.Errorf("unexpected token audience: %v", claims.Audience)
	}

	// iss
	if !claims.VerifyIssuer(defaultIssuer, true) {
		return "", fmt.Errorf("unexpected token issuer: %v", claims.Issuer)
	}

	// exp
	if claims.ExpiresAt == nil || claims.ExpiresAt.Time.Before(time.Now()) {
		return "", errors.New("token has expired")
	}

	return claims.Subject, nil
}
