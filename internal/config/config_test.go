package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetViper() {
	viper.Reset()
}

func TestLoad_Defaults(t *testing.T) {
	resetViper()

	cfg, err := Load("")
	require.NoError(t, err)

	assert.Equal(t, "8080", cfg.Server.Port)
	assert.Equal(t, "debug", cfg.Server.Mode)
	assert.Equal(t, "http://localhost:3000", cfg.Server.FrontendURL)
	assert.Equal(t, "localhost", cfg.DB.Host)
	assert.Equal(t, "5432", cfg.DB.Port)
	assert.Equal(t, "disable", cfg.DB.SSLMode)
	assert.Equal(t, 25, cfg.DB.MaxOpenConns)
	assert.Equal(t, 5, cfg.DB.MaxIdleConns)
	assert.Equal(t, time.Hour, cfg.DB.ConnMaxLifetime)
	assert.Equal(t, 30*time.Minute, cfg.DB.ConnMaxIdleTime)
	assert.Equal(t, "localhost", cfg.Redis.Host)
	assert.Equal(t, "6379", cfg.Redis.Port)
	assert.Equal(t, 0, cfg.Redis.DB)
	assert.Equal(t, "http://localhost:7700", cfg.Meili.URL)
	assert.Equal(t, 15*time.Minute, cfg.JWT.AccessExpiry)
	assert.Equal(t, 168*time.Hour, cfg.JWT.RefreshExpiry)
	assert.Equal(t, "http://localhost:9000", cfg.RustFS.Endpoint)
	assert.Equal(t, "cms-media", cfg.RustFS.Bucket)
	assert.Equal(t, "debug", cfg.Log.Level)
	assert.Equal(t, "json", cfg.Log.Format)
}

func TestLoad_FromEnvFile(t *testing.T) {
	resetViper()

	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	content := "SERVER_PORT=9090\nSERVER_MODE=release\nDB_NAME=mydb\nDB_USER=admin\nDB_PASSWORD=secret\n"
	require.NoError(t, os.WriteFile(envFile, []byte(content), 0644))

	cfg, err := Load(envFile)
	require.NoError(t, err)

	assert.Equal(t, "9090", cfg.Server.Port)
	assert.Equal(t, "release", cfg.Server.Mode)
	assert.Equal(t, "mydb", cfg.DB.Name)
	assert.Equal(t, "admin", cfg.DB.User)
	assert.Equal(t, "secret", cfg.DB.Password)
}

func TestLoad_EnvVarOverride(t *testing.T) {
	resetViper()
	t.Setenv("SERVER_PORT", "7777")
	t.Setenv("DB_NAME", "override_db")

	cfg, err := Load("")
	require.NoError(t, err)

	assert.Equal(t, "7777", cfg.Server.Port)
	assert.Equal(t, "override_db", cfg.DB.Name)
}

func TestLoad_InvalidDuration(t *testing.T) {
	resetViper()
	t.Setenv("JWT_ACCESS_EXPIRY", "not-a-duration")

	_, err := Load("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_ACCESS_EXPIRY")
}

func TestLoad_InvalidDBConnMaxLifetime(t *testing.T) {
	resetViper()
	t.Setenv("DB_CONN_MAX_LIFETIME", "bad")

	_, err := Load("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DB_CONN_MAX_LIFETIME")
}

func TestDBConfig_DSN(t *testing.T) {
	cfg := &DBConfig{
		User:     "admin",
		Password: "secret",
		Host:     "db.example.com",
		Port:     "5433",
		Name:     "cms",
		SSLMode:  "require",
	}
	expected := "postgres://admin:secret@db.example.com:5433/cms?sslmode=require"
	assert.Equal(t, expected, cfg.DSN())
}

func TestRedisConfig_Addr(t *testing.T) {
	cfg := &RedisConfig{Host: "redis.local", Port: "6380"}
	assert.Equal(t, "redis.local:6380", cfg.Addr())
}
