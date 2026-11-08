package config_test

import (
	"testing"
	"time"

	"github.com/Polikarpov00/test_task_wallet/internal/config"
)

// Сборка строки подключения к Postgres.
// Спецсимволы в пароле (@, /) должны экранироваться,
// иначе DSN распарсится неправильно.
func TestDBConfig_DSN(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		password string
		want     string
	}{
		{
			name: "обычный случай", host: "localhost", password: "secret",
			want: "postgres://wallet:secret@localhost:5432/wallet?sslmode=disable",
		},
		{
			name: "спецсимволы в пароле экранируются", host: "localhost", password: "p@ss/word",
			want: "postgres://wallet:p%40ss%2Fword@localhost:5432/wallet?sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.DBConfig{
				Host: tt.host, Port: "5432",
				User: "wallet", Password: tt.password, Name: "wallet",
			}
			if got := cfg.DSN(); got != tt.want {
				t.Errorf("DSN() = %q, want %q", got, tt.want)
			}
		})
	}
}

// Заданы только обязательные переменные.
// Ожидаем, что они прочитаются, а остальные поля
// получат значения по умолчанию (порт 5432, MaxConns 20, ReadTimeout 5s).
func TestLoad_Success(t *testing.T) {
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_USER", "test_user")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_NAME", "test_db")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if got := cfg.DB.Host; got != "localhost" {
		t.Errorf("DB.Host = %q, want %q", got, "localhost")
	}
	if got := cfg.DB.Port; got != "5432" {
		t.Errorf("DB.Port = %q, want %q", got, "5432")
	}
	if got := cfg.DB.MaxConns; got != 20 {
		t.Errorf("DB.MaxConns = %d, want %d", got, 20)
	}
	if got := cfg.HTTP.ReadTimeout; got != 5*time.Second {
		t.Errorf("HTTP.ReadTimeout = %v, want %v", got, 5*time.Second)
	}
}

// Пустой DB_HOST.
// Ожидаем ошибку при загрузке: сервис не должен стартовать с неполным конфигом.
func TestLoad_MissingRequired(t *testing.T) {
	t.Setenv("DB_HOST", "")
	t.Setenv("DB_USER", "test_user")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_NAME", "test_db")

	_, err := config.Load()
	if err == nil {
		t.Error("Load() with empty DB_HOST: expected error, got nil")
	}
}
