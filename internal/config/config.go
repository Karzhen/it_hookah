package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App    AppConfig
	DB     DBConfig
	JWT    JWTConfig
	Server ServerConfig
}

type AppConfig struct {
	Name string
	Env  string
	Port string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type JWTConfig struct {
	AccessSecret     string
	RefreshSecret    string
	AccessTTLMinutes int
	RefreshTTLHours  int
}

type ServerConfig struct {
	ReadTimeoutSeconds     int
	WriteTimeoutSeconds    int
	IdleTimeoutSeconds     int
	ShutdownTimeoutSeconds int
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		App: AppConfig{
			Name: getEnv("APP_NAME", "lk-backend"),
			Env:  getEnv("APP_ENV", "development"),
			Port: getEnv("APP_PORT", "8080"),
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "lk_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			AccessSecret:  getEnv("JWT_ACCESS_SECRET", ""),
			RefreshSecret: getEnv("JWT_REFRESH_SECRET", ""),
		},
		Server: ServerConfig{
			ReadTimeoutSeconds:     getEnvAsInt("HTTP_READ_TIMEOUT_SECONDS", 10),
			WriteTimeoutSeconds:    getEnvAsInt("HTTP_WRITE_TIMEOUT_SECONDS", 10),
			IdleTimeoutSeconds:     getEnvAsInt("HTTP_IDLE_TIMEOUT_SECONDS", 60),
			ShutdownTimeoutSeconds: getEnvAsInt("HTTP_SHUTDOWN_TIMEOUT_SECONDS", 10),
		},
	}

	cfg.JWT.AccessTTLMinutes = getEnvAsInt("JWT_ACCESS_TTL_MINUTES", 15)
	cfg.JWT.RefreshTTLHours = getEnvAsInt("JWT_REFRESH_TTL_HOURS", 720)

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	missing := make([]string, 0)

	if strings.TrimSpace(c.JWT.AccessSecret) == "" {
		missing = append(missing, "JWT_ACCESS_SECRET")
	}
	if strings.TrimSpace(c.JWT.RefreshSecret) == "" {
		missing = append(missing, "JWT_REFRESH_SECRET")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required env vars: %s", strings.Join(missing, ", "))
	}

	if c.JWT.AccessTTLMinutes <= 0 {
		return errors.New("JWT_ACCESS_TTL_MINUTES must be greater than 0")
	}
	if c.JWT.RefreshTTLHours <= 0 {
		return errors.New("JWT_REFRESH_TTL_HOURS must be greater than 0")
	}

	return nil
}

func (c Config) HTTPAddr() string {
	return ":" + c.App.Port
}

func (c Config) ReadTimeout() time.Duration {
	return time.Duration(c.Server.ReadTimeoutSeconds) * time.Second
}

func (c Config) WriteTimeout() time.Duration {
	return time.Duration(c.Server.WriteTimeoutSeconds) * time.Second
}

func (c Config) IdleTimeout() time.Duration {
	return time.Duration(c.Server.IdleTimeoutSeconds) * time.Second
}

func (c Config) ShutdownTimeout() time.Duration {
	return time.Duration(c.Server.ShutdownTimeoutSeconds) * time.Second
}

func (c DBConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", c.User, c.Password, c.Host, c.Port, c.Name, c.SSLMode)
}

func getEnv(key string, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
