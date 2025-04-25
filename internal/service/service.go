package service

import (
	"context"
	"fmt"

	"github.com/AlexMickh/speak-auth/internal/config"
	"github.com/AlexMickh/speak-auth/internal/domain/models"
	"github.com/AlexMickh/speak-auth/internal/lib/jwt"
	"golang.org/x/crypto/bcrypt"
)

type Storage interface {
	SaveUser(ctx context.Context, username, email, password string) (string, error)
	GetUser(ctx context.Context, email, password string) (*models.User, error)
}

type Service struct {
	cfg     config.JwtConfig
	storage Storage
}

func New(cfg config.JwtConfig, storage Storage) *Service {
	return &Service{
		cfg:     cfg,
		storage: storage,
	}
}

func (s *Service) Register(ctx context.Context, username, email, password string) (string, error) {
	const op = "service.Register"

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	id, err := s.storage.SaveUser(ctx, username, email, string(hashPassword))
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (string, string, error) {
	const op = "service.Login"

	user, err := s.storage.GetUser(ctx, email, password)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	accessToken, err := jwt.GenAccess(s.cfg, user.ID.String(), user.Email, user.Username)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	refreshToken, err := jwt.GenRefresh(s.cfg, user.ID.String())
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	return accessToken, refreshToken, nil
}
