package appcfg

import (
	basecfg "github.com/vocbl/shared/infrastructure/config"
	envload "github.com/vocbl/shared/infrastructure/config/envload"
)

type App struct {
	Nats     Nats        `koanf:"nats"`
	GRPC     GRPC        `koanf:"grpc"`
	Postgres PostgresSQL `koanf:"postgresql"`
	Redis    Redis       `koanf:"redis"`
	Cfg      Cfg         `koanf:"cfg"`
}

type Nats struct {
	basecfg.Nats `koanf:",squash"`
}

type GRPC struct {
	basecfg.Grpc `koanf:",squash"`
}

type PostgresSQL struct {
	basecfg.PostgreSQL `koanf:",squash"`
}

type Redis struct {
	basecfg.Redis `koanf:",squash"`
}

type Cfg struct {
	basecfg.App `koanf:",squash"`
}

func Load(envfiles ...string) (*App, error) {
	return envload.Do[App]("IDENTITY_SVC__", envfiles...)
}
