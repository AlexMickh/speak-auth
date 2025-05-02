package server

import (
	"context"
	"log/slog"
	"net/mail"
	"strings"

	"github.com/AlexMickh/speak-auth/internal/config"
	"github.com/AlexMickh/speak-auth/internal/lib/email"

	// "github.com/AlexMickh/speak-auth/internal/storage"
	"github.com/AlexMickh/speak-auth/pkg/logger"
	"github.com/AlexMickh/speak-protos/pkg/api/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Service interface {
	Register(ctx context.Context,
		email string,
		username string,
		password string,
		description string,
		profileImage []byte,
	) (string, error)
	Login(ctx context.Context, email, password string) (string, string, error)
	VerifyEmail(ctx context.Context, id string) error
	UpdateTokens(ctx context.Context, accessToken, refreshToken string) (string, string, error)
}

type Server struct {
	auth.UnimplementedAuthServer
	service Service
	log     *slog.Logger
	cfg     config.MailConfig
}

func New(service Service, log *slog.Logger, cfg config.MailConfig) *Server {
	return &Server{
		service: service,
		log:     log,
		cfg:     cfg,
	}
}

func (s *Server) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	const op = "server.Register"

	log := s.log.With(
		slog.String("op", op),
		slog.String("email", req.GetEmail()),
	)

	if req.GetUsername() == "" {
		log.Error("username is empty")
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}

	if req.GetEmail() == "" {
		log.Error("email is empty")
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	_, err := mail.ParseAddress(req.GetEmail())
	if err != nil {
		log.Error("not real email address")
		return nil, status.Error(codes.InvalidArgument, "not real email address")
	}

	if req.GetPassword() == "" {
		log.Error("password is empty")
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	id, err := s.service.Register(
		ctx,
		req.GetEmail(),
		req.GetUsername(),
		req.GetPassword(),
		req.GetDescription(),
		req.GetProfileImage(),
	)
	if err != nil {
		log.Error("failed to save user", logger.Err(err))
		return nil, status.Error(codes.Internal, "failed to save user")
	}

	err = email.Send(s.cfg, req.GetEmail(), id, req.GetUsername())
	if err != nil {
		log.Error("failed to send email", logger.Err(err))
	}

	return &auth.RegisterResponse{
		Id: id,
	}, nil
}

func (s *Server) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	const op = "server.Login"

	log := s.log.With(
		slog.String("op", op),
		slog.String("email", req.GetEmail()),
	)

	if req.GetEmail() == "" {
		log.Error("email is empty")
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if req.GetPassword() == "" {
		log.Error("password is empty")
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	_, err := mail.ParseAddress(req.GetEmail())
	if err != nil {
		log.Error("not real email address")
		return nil, status.Error(codes.InvalidArgument, "not real email address")
	}

	accessToken, refreshToken, err := s.service.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		// if errors.Is(err, storage.ErrNotFound) {
		// 	log.Error("user not found")
		// 	return nil, status.Error(codes.NotFound, "incorect login or password")
		// }
		// if errors.Is(err, storage.ErrEmailNotVerify) {
		// 	log.Error("email not verified")
		// 	return nil, status.Error(codes.PermissionDenied, "email not verified")
		// }
		log.Error("failed to login user", logger.Err(err))
		return nil, status.Error(codes.Internal, "failed to login user")
	}

	return &auth.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *Server) VerifyEmail(ctx context.Context, req *auth.VerifyEmailRequest) (*emptypb.Empty, error) {
	const op = "server.VerifyEmail"

	log := s.log.With(
		slog.String("op", op),
	)

	if req.GetId() == "" {
		log.Error("id is empry")
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	err := s.service.VerifyEmail(ctx, req.GetId())
	if err != nil {
		// if errors.Is(err, storage.ErrNotFound) {
		// 	log.Error("user not found")
		// 	return nil, status.Error(codes.NotFound, "user not found")
		// }
		log.Error("failed to verify email", logger.Err(err))
		return nil, status.Error(codes.Internal, "failed to verify email")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) UpdateTokens(ctx context.Context, req *auth.UpdateTokensRequest) (
	*auth.UpdateTokensResponse,
	error,
) {
	const op = "server.UpdateTokens"

	log := s.log.With(
		slog.String("op", op),
	)

	if req.GetRefreshToken() == "" {
		log.Error("refresh token is empry")
		return nil, status.Error(codes.InvalidArgument, "refresh token is required")
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Error("failed to get metadata")
		return nil, status.Error(codes.InvalidArgument, "failed to get metadata")
	}

	token, ok := md["authorization"]
	if !ok {
		log.Error("failed to get auth header")
		return nil, status.Error(codes.InvalidArgument, "access token is required")
	}
	if strings.Split(token[0], " ")[0] != "Bearer" {
		log.Error("wrong type of token")
		return nil, status.Error(codes.InvalidArgument, "wrong type of token")
	}

	accessToken, refreshToken, err := s.service.UpdateTokens(
		ctx,
		strings.Split(token[0], " ")[1],
		req.GetRefreshToken(),
	)
	if err != nil {
		log.Error("failed to gen tokens", logger.Err(err))
		return nil, status.Error(codes.Internal, "failed to update tokens")
	}

	return &auth.UpdateTokensResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
