package natsclient

import (
	"context"

	"github.com/nats-io/nats.go"
	natsclient "github.com/vocbl/shared/infrastructure/nats"
	appcfg "github.com/vocbl/users-svc/internal/infrastructure/cfg"

	"go.uber.org/zap"
)

func Connect(ctx context.Context, cfg appcfg.Nats, log *zap.Logger) (*nats.Conn, error) {
	return natsclient.Connect(ctx, cfg.Nats, nil, log)
}
