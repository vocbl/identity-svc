package postgresql

import (
	"context"

	postgresqlclient "github.com/vocbl/shared/infrastructure/postgresql"
	appcfg "github.com/vocbl/users-svc/internal/infrastructure/cfg"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Connect(ctx context.Context, cfg appcfg.PostgresSQL, log *zap.Logger) (*gorm.DB, error) {
	return postgresqlclient.Connect(ctx, cfg.PostgreSQL, nil, log)
}
