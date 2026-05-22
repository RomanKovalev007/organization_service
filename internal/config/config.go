package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	HTTP HTTP
	DB   DB
}

type HTTP struct {
	Addr         string        `env:"HTTP_ADDR"          env-default:":8080"`
	ReadTimeout  time.Duration `env:"HTTP_READ_TIMEOUT"  env-default:"5s"`
	WriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT" env-default:"10s"`
	IdleTimeout  time.Duration `env:"HTTP_IDLE_TIMEOUT"  env-default:"60s"`
}

type DB struct {
	Host            string        `env:"DB_HOST"              env-required:"true"`
	Port            int           `env:"DB_PORT"              env-default:"5432"`
	User            string        `env:"DB_USER"              env-required:"true"`
	Password        string        `env:"DB_PASSWORD"          env-required:"true"`
	Name            string        `env:"DB_NAME"              env-required:"true"`
	SSLMode         string        `env:"DB_SSLMODE"           env-default:"disable"`
	MaxOpenConns    int           `env:"DB_MAX_OPEN_CONNS"    env-default:"10"`
	MaxIdleConns    int           `env:"DB_MAX_IDLE_CONNS"    env-default:"5"`
	ConnMaxLifetime time.Duration `env:"DB_CONN_MAX_LIFETIME" env-default:"1h"`
}

func (d DB) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

func Load() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	return &cfg, nil
}
