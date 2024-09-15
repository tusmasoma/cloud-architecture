package config

import (
	"context"

	"github.com/sethvargo/go-envconfig"

	"github.com/tusmasoma/go-tech-dojo/pkg/log"
)

const (
	dbPrefix = "MYSQL_"
)

type DBConfig struct {
	Host     string `env:"HOST, required"`
	Port     string `env:"PORT, required"`
	User     string `env:"USER, required"`
	Password string `env:"PASSWORD, required"`
	DBName   string `env:"DB_NAME, required"`
}

func NewDBConfig(ctx context.Context) (*DBConfig, error) {
	conf := &DBConfig{}
	pl := envconfig.PrefixLookuper(dbPrefix, envconfig.OsLookuper())
	if err := envconfig.ProcessWith(ctx, conf, pl); err != nil {
		log.Error("Failed to load database config", log.Ferror(err))
		return nil, err
	}
	return conf, nil
}
