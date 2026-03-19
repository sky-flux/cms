package config

import (
	"fmt"
	"time"

	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Config holds all application configuration. Field names and types are
// identical to the previous Viper-based version to preserve compatibility.
type Config struct {
	Server ServerConfig
	DB     DBConfig
	Redis  RedisConfig
	Meili  MeiliConfig
	JWT    JWTConfig
	TOTP   TOTPConfig
	RustFS RustFSConfig
	Resend ResendConfig
	Log    LogConfig
}

type ServerConfig struct {
	Port        string
	Mode        string
	FrontendURL string
}

type DBConfig struct {
	Host            string
	Port            string
	Name            string
	User            string
	Password        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func (c *DBConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Name, c.SSLMode)
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func (c *RedisConfig) Addr() string {
	return c.Host + ":" + c.Port
}

type MeiliConfig struct {
	URL       string
	MasterKey string
}

type JWTConfig struct {
	Secret        string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

type TOTPConfig struct {
	EncryptionKey string
}

type RustFSConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	Region    string
}

type ResendConfig struct {
	APIKey    string
	FromName  string
	FromEmail string
}

type LogConfig struct {
	Level string
}

// defaults maps flat ENV key names to their default values.
// koanf uses these as the base layer before file and env providers.
var defaults = map[string]any{
	"SERVER_PORT":           "8080",
	"SERVER_MODE":           "debug",
	"FRONTEND_URL":          "http://localhost:3000",
	"DB_HOST":               "localhost",
	"DB_PORT":               "5432",
	"DB_SSLMODE":            "disable",
	"DB_MAX_OPEN_CONNS":     25,
	"DB_MAX_IDLE_CONNS":     5,
	"DB_CONN_MAX_LIFETIME":  "1h",
	"DB_CONN_MAX_IDLE_TIME": "30m",
	"REDIS_HOST":            "localhost",
	"REDIS_PORT":            "6379",
	"REDIS_DB":              0,
	"MEILI_URL":             "http://localhost:7700",
	"JWT_ACCESS_EXPIRY":     "15m",
	"JWT_REFRESH_EXPIRY":    "168h",
	"RUSTFS_ENDPOINT":       "http://localhost:9000",
	"RUSTFS_ACCESS_KEY":     "rustfsadmin",
	"RUSTFS_SECRET_KEY":     "rustfsadmin",
	"RUSTFS_BUCKET":         "cms-media",
	"RUSTFS_REGION":         "us-east-1",
	"RESEND_FROM_NAME":      "Sky Flux CMS",
	"RESEND_FROM_EMAIL":     "noreply@example.com",
	"LOG_LEVEL":             "debug",
}

// Load reads configuration with priority: ENV vars > .env file > built-in defaults.
// cfgFile may be empty string, in which case ".env" in the working directory is
// attempted (failure is silently ignored — env vars alone are sufficient).
func Load(cfgFile string) (*Config, error) {
	k := koanf.New(".")

	// Layer 1: built-in defaults (lowest priority)
	for key, val := range defaults {
		if err := k.Set(key, val); err != nil {
			return nil, fmt.Errorf("set default %s: %w", key, err)
		}
	}

	// Layer 2: .env file (optional, silently ignored if missing)
	envPath := ".env"
	if cfgFile != "" {
		envPath = cfgFile
	}
	_ = k.Load(file.Provider(envPath), dotenv.Parser())

	// Layer 3: environment variables (highest priority among non-flag sources)
	// Pass-through transformer: upper-case env vars map directly to koanf keys.
	if err := k.Load(env.Provider("", ".", func(s string) string { return s }), nil); err != nil {
		return nil, fmt.Errorf("load env vars: %w", err)
	}

	// Parse duration fields
	accessExpiry, err := time.ParseDuration(k.String("JWT_ACCESS_EXPIRY"))
	if err != nil {
		return nil, fmt.Errorf("parse JWT_ACCESS_EXPIRY: %w", err)
	}

	refreshExpiry, err := time.ParseDuration(k.String("JWT_REFRESH_EXPIRY"))
	if err != nil {
		return nil, fmt.Errorf("parse JWT_REFRESH_EXPIRY: %w", err)
	}

	connMaxLifetime, err := time.ParseDuration(k.String("DB_CONN_MAX_LIFETIME"))
	if err != nil {
		return nil, fmt.Errorf("parse DB_CONN_MAX_LIFETIME: %w", err)
	}

	connMaxIdleTime, err := time.ParseDuration(k.String("DB_CONN_MAX_IDLE_TIME"))
	if err != nil {
		return nil, fmt.Errorf("parse DB_CONN_MAX_IDLE_TIME: %w", err)
	}

	cfg := &Config{
		Server: ServerConfig{
			Port:        k.String("SERVER_PORT"),
			Mode:        k.String("SERVER_MODE"),
			FrontendURL: k.String("FRONTEND_URL"),
		},
		DB: DBConfig{
			Host:            k.String("DB_HOST"),
			Port:            k.String("DB_PORT"),
			Name:            k.String("DB_NAME"),
			User:            k.String("DB_USER"),
			Password:        k.String("DB_PASSWORD"),
			SSLMode:         k.String("DB_SSLMODE"),
			MaxOpenConns:    k.Int("DB_MAX_OPEN_CONNS"),
			MaxIdleConns:    k.Int("DB_MAX_IDLE_CONNS"),
			ConnMaxLifetime: connMaxLifetime,
			ConnMaxIdleTime: connMaxIdleTime,
		},
		Redis: RedisConfig{
			Host:     k.String("REDIS_HOST"),
			Port:     k.String("REDIS_PORT"),
			Password: k.String("REDIS_PASSWORD"),
			DB:       k.Int("REDIS_DB"),
		},
		Meili: MeiliConfig{
			URL:       k.String("MEILI_URL"),
			MasterKey: k.String("MEILI_MASTER_KEY"),
		},
		JWT: JWTConfig{
			Secret:        k.String("JWT_SECRET"),
			AccessExpiry:  accessExpiry,
			RefreshExpiry: refreshExpiry,
		},
		TOTP: TOTPConfig{
			EncryptionKey: k.String("TOTP_ENCRYPTION_KEY"),
		},
		RustFS: RustFSConfig{
			Endpoint:  k.String("RUSTFS_ENDPOINT"),
			AccessKey: k.String("RUSTFS_ACCESS_KEY"),
			SecretKey: k.String("RUSTFS_SECRET_KEY"),
			Bucket:    k.String("RUSTFS_BUCKET"),
			Region:    k.String("RUSTFS_REGION"),
		},
		Resend: ResendConfig{
			APIKey:    k.String("RESEND_API_KEY"),
			FromName:  k.String("RESEND_FROM_NAME"),
			FromEmail: k.String("RESEND_FROM_EMAIL"),
		},
		Log: LogConfig{
			Level: k.String("LOG_LEVEL"),
		},
	}

	return cfg, nil
}
