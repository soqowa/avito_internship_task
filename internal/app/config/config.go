package config

import (
	"github.com/caarlos0/env/v9"
)

type Config struct {
	Port     string `env:"PORT" envDefault:"8080"`
	DBDSN    string `env:"DB_DSN" envDefault:"postgres://postgres:postgres@postgres:5432/reviewer?sslmode=disable"`
	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`
}

func Load() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
