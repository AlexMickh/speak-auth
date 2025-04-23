package main

import (
	"context"
	"time"

	"github.com/AlexMickh/logger/pkg/logger"
	"github.com/AlexMickh/speak-auth/internal/config"
	"go.uber.org/zap"
)

func main() {
	cfg := config.MustLoad()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ctx, err := logger.New(ctx, cfg.Env)
	if err != nil {
		panic("failed to init logger")
	}

	logger.GetFromCtx(ctx).Info(ctx, "logger is working", zap.String("env", cfg.Env))

}
