package main

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/AlexMickh/speak-auth/internal/config"
	"github.com/AlexMickh/speak-auth/internal/grpc/client"
	"github.com/AlexMickh/speak-auth/internal/grpc/server"
	"github.com/AlexMickh/speak-auth/internal/service"
	"github.com/AlexMickh/speak-auth/pkg/logger"
	"github.com/AlexMickh/speak-protos/pkg/api/auth"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.MustLoad()

	var w io.Writer
	if cfg.Env == "local" {
		w = os.Stdout
	} else {
		file, err := os.OpenFile("log.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		w = file
	}

	log := logger.SetupLogger(w, cfg.Env)
	log.Info("logger is working", slog.String("env", cfg.Env))

	log.Info("starting grpc server", slog.Int("port", cfg.Port))

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", cfg.Port))
	if err != nil {
		log.Error("failed to listen", logger.Err(err))
		os.Exit(1)
	}

	client, err := client.New(cfg.UserServiceAddr)
	if err != nil {
		log.Error("failed to init client", logger.Err(err))
		os.Exit(1)
	}
	defer client.Close()

	service := service.New(cfg.Jwt, client)

	srv := server.New(service, log, cfg.Mail)
	server := grpc.NewServer()
	auth.RegisterAuthServer(server, srv)

	go func() {
		if err := server.Serve(lis); err != nil {
			log.Error("faild to serve", logger.Err(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	server.GracefulStop()

	log.Info("server stoped")
}
