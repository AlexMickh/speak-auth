package jwt

import (
	"fmt"
	"time"

	"github.com/AlexMickh/speak-auth/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

func GenAccess(cfg config.JwtConfig, id, email, username string) (string, error) {
	const op = "lib.jwt.GenAccess"

	tokenString, err := generate(cfg, id, email, username)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return tokenString, nil
}

func GenRefresh(cfg config.JwtConfig, id string) (string, error) {
	const op = "lib.jwt.GenRefresh"

	tokenString, err := generate(cfg, id)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return tokenString, nil
}

func generate(cfg config.JwtConfig, args ...string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	if len(args) == 1 {
		claims["uid"] = args[0]
	} else {
		claims["uid"] = args[0]
		claims["email"] = args[1]
		claims["username"] = args[2]
		claims["ttl"] = time.Now().Add(cfg.Ttl).Unix()
	}

	tokenString, err := token.SignedString([]byte(cfg.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
