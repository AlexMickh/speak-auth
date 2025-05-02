package service

import (
	"context"
	"fmt"

	"github.com/AlexMickh/speak-auth/internal/config"
	"github.com/AlexMickh/speak-auth/internal/domain/models"
	"github.com/AlexMickh/speak-auth/internal/lib/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Client interface {
	CreateUser(
		ctx context.Context,
		email string,
		username string,
		password string,
		description string,
		profileImage []byte,
	) (string, error)
	GetUserInfo(ctx context.Context, email string) (models.User, error)
	VerifyEmail(ctx context.Context, id string) error
}

type Service struct {
	cfg    config.JwtConfig
	client Client
}

func New(cfg config.JwtConfig, client Client) *Service {
	return &Service{
		cfg:    cfg,
		client: client,
	}
}

func (s *Service) Register(ctx context.Context,
	email string,
	username string,
	password string,
	description string,
	profileImage []byte,
) (string, error) {
	const op = "service.Register"

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	id, err := s.client.CreateUser(
		ctx,
		email,
		username,
		string(hashPassword),
		description,
		profileImage,
	)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (string, string, error) {
	const op = "service.Login"

	user, err := s.client.GetUserInfo(ctx, email)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	accessToken, err := jwt.GenAccess(s.cfg, user.ID, email, user.Username)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	refreshToken, err := jwt.GenRefresh(s.cfg, user.ID)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	return accessToken, refreshToken, nil
}

func (s *Service) VerifyEmail(ctx context.Context, id string) error {
	const op = "service.VerifyEmail"

	err := uuid.Validate(id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = s.client.VerifyEmail(ctx, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Service) UpdateTokens(ctx context.Context, accessToken, refreshToken string) (string, string, error) {
	const op = "service.UpdateTokens"

	accessToken, refreshToken, err := jwt.UpdateTokens(s.cfg, accessToken, refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	return accessToken, refreshToken, nil
}
