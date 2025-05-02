package client

import (
	"context"
	"fmt"

	"github.com/AlexMickh/speak-auth/internal/domain/models"
	"github.com/AlexMickh/speak-protos/pkg/api/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client user.UserClient
}

func New(addr string) (*Client, error) {
	const op = "grpc.client.New"

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	client := user.NewUserClient(conn)

	return &Client{
		conn:   conn,
		client: client,
	}, nil
}

func (c *Client) CreateUser(
	ctx context.Context,
	email string,
	username string,
	password string,
	description string,
	profileImage []byte,
) (string, error) {
	const op = "grpc.client.CreateUser"

	res, err := c.client.CreateUser(ctx, &user.CreateUserRequest{
		Email:        email,
		Username:     username,
		Password:     password,
		Description:  description,
		ProfileImage: profileImage,
	})
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return res.GetId(), nil
}

func (c *Client) GetUserInfo(ctx context.Context, email string) (models.User, error) {
	const op = "grpc.client.GetUser"

	res, err := c.client.GetUser(ctx, &user.GetUserRequest{
		Email: email,
	})
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return models.User{
		ID:       res.GetId(),
		Username: res.GetUsername(),
		Password: res.GetPassword(),
	}, nil
}

func (c *Client) VerifyEmail(ctx context.Context, id string) error {
	const op = "grpc.server.VerifyEmail"

	_, err := c.client.VerifyEmail(ctx, &user.VerifyEmailRequest{
		Id: id,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Client) Close() {
	c.conn.Close()
}
