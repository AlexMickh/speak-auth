package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/AlexMickh/speak-auth/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

type userInfo struct {
	id       string
	email    string
	username string
	ttl      int64
}

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

func UpdateTokens(cfg config.JwtConfig, accessToken, refreshToken string) (string, string, error) {
	const op = "lib.jwt.UpdateTokens"

	user, err := encodeToken(cfg, accessToken)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	id, ttl, err := encodeRefreshToken(cfg, refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	if user.id != id {
		return "", "", fmt.Errorf("%s: failed to compare id", op)
	}
	if ttl < time.Now().Unix() {
		return "", "", fmt.Errorf("%s: refresh token expired", op)
	}

	accessToken, err = GenAccess(cfg, user.id, user.email, user.username)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	refreshToken, err = GenRefresh(cfg, id)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	return accessToken, refreshToken, nil
}

func encodeToken(cfg config.JwtConfig, token string) (userInfo, error) {
	const op = "lib.jwt.encodeToken"

	claims := jwt.MapClaims{}

	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (any, error) {
		return []byte(cfg.Secret), nil
	})
	if err != nil {
		return userInfo{}, fmt.Errorf("%s: %w", op, err)
	}

	id, ok := claims["uid"].(string)
	if !ok {
		return userInfo{}, fmt.Errorf("%s: failed to get user id", op)
	}
	email, ok := claims["email"].(string)
	if !ok {
		return userInfo{}, fmt.Errorf("%s: failed to get email", op)
	}
	username, ok := claims["username"].(string)
	if !ok {
		return userInfo{}, fmt.Errorf("%s: failed to get username", op)
	}
	ttl, ok := claims["ttl"].(float64)
	if !ok {
		return userInfo{}, fmt.Errorf("%s: failed to get ttl", op)
	}

	return userInfo{
		id:       id,
		email:    email,
		username: username,
		ttl:      int64(ttl),
	}, nil
}

func encodeRefreshToken(cfg config.JwtConfig, token string) (string, int64, error) {
	const op = "lib.jwt.encodeRefreshToken"

	claims := jwt.MapClaims{}

	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (any, error) {
		return []byte(cfg.Secret), nil
	})
	if err != nil {
		return "", -1, fmt.Errorf("%s: %w", op, err)
	}

	id, ok := claims["uid"].(string)
	if !ok {
		return "", -1, fmt.Errorf("%s: failed to get user id", op)
	}
	ttl, ok := claims["ttl"].(float64)
	if !ok {
		return "", -1, fmt.Errorf("%s: failed to get ttl", op)
	}

	return id, int64(ttl), nil
}

func generate(cfg config.JwtConfig, args ...string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	switch len(args) {
	case 1:
		claims["uid"] = args[0]
		claims["ttl"] = time.Now().Add(cfg.RefreshTtl).Unix()
	case 3:
		claims["uid"] = args[0]
		claims["email"] = args[1]
		claims["username"] = args[2]
		claims["ttl"] = time.Now().Add(cfg.AccessTtl).Unix()
	default:
		return "", errors.New("wrong number of arguments")
	}

	tokenString, err := token.SignedString([]byte(cfg.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
