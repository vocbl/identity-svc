package redisclient

import (
	"context"

	"github.com/redis/go-redis/v9"
	redisclient "github.com/vocbl/shared/infrastructure/redis"
	appcfg "github.com/vocbl/users-svc/internal/infrastructure/cfg"

	"go.uber.org/zap"
)

func Connect(ctx context.Context, cfg appcfg.Redis, log *zap.Logger) (*redis.Client, error) {
	return redisclient.Connect(ctx, cfg.Redis, nil, log)
}
