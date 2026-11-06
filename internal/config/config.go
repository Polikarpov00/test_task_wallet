package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	HTTP HTTPConfig
	DB   DBConfig
}

type HTTPConfig struct {
	Port            string        `env:"APP_PORT"              envDefault:"8080"`
	ReadTimeout     time.Duration `env:"HTTP_READ_TIMEOUT"     envDefault:"5s"`
	WriteTimeout    time.Duration `env:"HTTP_WRITE_TIMEOUT"    envDefault:"10s"`
	ShutdownTimeout time.Duration `env:"HTTP_SHUTDOWN_TIMEOUT" envDefault:"10s"`
}

type DBConfig struct {
	Host     string `env:"DB_HOST,required,notEmpty"`
	Port     string `env:"DB_PORT"      envDefault:"5432"`
	User     string `env:"DB_USER,required,notEmpty"`
	Password string `env:"DB_PASSWORD,required,notEmpty"`
	Name     string `env:"DB_NAME,required,notEmpty"`
	MaxConns int32  `env:"DB_MAX_CONNS" envDefault:"20"`
	MinConns int32  `env:"DB_MIN_CONNS" envDefault:"2"`
}

func Load() (*Config, error) {
	err := godotenv.Load("config.env")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			//просто обработка случая os.ErrNotExist
		} else {
			return nil, fmt.Errorf("can't read config.env: %w", err)
		}
	}

	var cfg Config
	err = env.Parse(&cfg)
	if err != nil {
		return nil, fmt.Errorf("can't parse config.env: %w", err)
	}
	return &cfg, nil
}

func (cfg *DBConfig) DSN() string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     net.JoinHostPort(cfg.Host, cfg.Port),
		Path:     cfg.Name,
		RawQuery: "sslmode=disable",
	}

	return u.String()
}
