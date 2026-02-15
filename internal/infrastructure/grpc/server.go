package redisclient

import (
	"context"

	grpcserver "github.com/vocbl/shared/infrastructure/grpc"
	logger "github.com/vocbl/shared/logger"
	appcfg "github.com/vocbl/users-svc/internal/infrastructure/cfg"
	"go.uber.org/zap"
)

func Connect(ctx context.Context, cfg appcfg.GRPC, log *zap.Logger) (*grpcserver.Server, error) {
	return grpcserver.NewServer(cfg.Grpc, nil, logger.InterceptorGRPC(log), log)
}
