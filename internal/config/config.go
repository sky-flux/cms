package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server      ServerConfig
	DB          DBConfig
	Redis       RedisConfig
	Meili       MeiliConfig
	JWT         JWTConfig
	TOTP        TOTPConfig
	RustFS      RustFSConfig
	Resend      ResendConfig
	Log         LogConfig
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

func Load(cfgFile string) (*Config, error) {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigFile(".env")
	}
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Defaults
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("SERVER_MODE", "debug")
	viper.SetDefault("FRONTEND_URL", "http://localhost:3000")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("DB_MAX_OPEN_CONNS", 25)
	viper.SetDefault("DB_MAX_IDLE_CONNS", 5)
	viper.SetDefault("DB_CONN_MAX_LIFETIME", "1h")
	viper.SetDefault("DB_CONN_MAX_IDLE_TIME", "30m")
	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("REDIS_DB", 0)
	viper.SetDefault("MEILI_URL", "http://localhost:7700")
	viper.SetDefault("JWT_ACCESS_EXPIRY", "15m")
	viper.SetDefault("JWT_REFRESH_EXPIRY", "168h")
	viper.SetDefault("RUSTFS_ENDPOINT", "http://localhost:9000")
	viper.SetDefault("RUSTFS_ACCESS_KEY", "rustfsadmin")
	viper.SetDefault("RUSTFS_SECRET_KEY", "rustfsadmin")
	viper.SetDefault("RUSTFS_BUCKET", "cms-media")
	viper.SetDefault("RUSTFS_REGION", "us-east-1")
	viper.SetDefault("RESEND_FROM_NAME", "Sky Flux CMS")
	viper.SetDefault("RESEND_FROM_EMAIL", "noreply@example.com")
	viper.SetDefault("LOG_LEVEL", "debug")


	// Read .env file (optional — env vars take precedence)
	_ = viper.ReadInConfig()

	accessExpiry, err := time.ParseDuration(viper.GetString("JWT_ACCESS_EXPIRY"))
	if err != nil {
		return nil, fmt.Errorf("parse JWT_ACCESS_EXPIRY: %w", err)
	}

	refreshExpiry, err := time.ParseDuration(viper.GetString("JWT_REFRESH_EXPIRY"))
	if err != nil {
		return nil, fmt.Errorf("parse JWT_REFRESH_EXPIRY: %w", err)
	}

	connMaxLifetime, err := time.ParseDuration(viper.GetString("DB_CONN_MAX_LIFETIME"))
	if err != nil {
		return nil, fmt.Errorf("parse DB_CONN_MAX_LIFETIME: %w", err)
	}

	connMaxIdleTime, err := time.ParseDuration(viper.GetString("DB_CONN_MAX_IDLE_TIME"))
	if err != nil {
		return nil, fmt.Errorf("parse DB_CONN_MAX_IDLE_TIME: %w", err)
	}

	cfg := &Config{
		Server: ServerConfig{
			Port:        viper.GetString("SERVER_PORT"),
			Mode:        viper.GetString("SERVER_MODE"),
			FrontendURL: viper.GetString("FRONTEND_URL"),
		},
		DB: DBConfig{
			Host:            viper.GetString("DB_HOST"),
			Port:            viper.GetString("DB_PORT"),
			Name:            viper.GetString("DB_NAME"),
			User:            viper.GetString("DB_USER"),
			Password:        viper.GetString("DB_PASSWORD"),
			SSLMode:         viper.GetString("DB_SSLMODE"),
			MaxOpenConns:    viper.GetInt("DB_MAX_OPEN_CONNS"),
			MaxIdleConns:    viper.GetInt("DB_MAX_IDLE_CONNS"),
			ConnMaxLifetime: connMaxLifetime,
			ConnMaxIdleTime: connMaxIdleTime,
		},
		Redis: RedisConfig{
			Host:     viper.GetString("REDIS_HOST"),
			Port:     viper.GetString("REDIS_PORT"),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
		},
		Meili: MeiliConfig{
			URL:       viper.GetString("MEILI_URL"),
			MasterKey: viper.GetString("MEILI_MASTER_KEY"),
		},
		JWT: JWTConfig{
			Secret:        viper.GetString("JWT_SECRET"),
			AccessExpiry:  accessExpiry,
			RefreshExpiry: refreshExpiry,
		},
		TOTP: TOTPConfig{
			EncryptionKey: viper.GetString("TOTP_ENCRYPTION_KEY"),
		},
		RustFS: RustFSConfig{
			Endpoint:  viper.GetString("RUSTFS_ENDPOINT"),
			AccessKey: viper.GetString("RUSTFS_ACCESS_KEY"),
			SecretKey: viper.GetString("RUSTFS_SECRET_KEY"),
			Bucket:    viper.GetString("RUSTFS_BUCKET"),
			Region:    viper.GetString("RUSTFS_REGION"),
		},
		Resend: ResendConfig{
			APIKey:    viper.GetString("RESEND_API_KEY"),
			FromName:  viper.GetString("RESEND_FROM_NAME"),
			FromEmail: viper.GetString("RESEND_FROM_EMAIL"),
		},
		Log: LogConfig{
			Level: viper.GetString("LOG_LEVEL"),
		},
	}

	return cfg, nil
}
