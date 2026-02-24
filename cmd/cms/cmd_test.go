package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"testing"

	"github.com/sky-flux/cms/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Command tree tests ---

func TestRootCommand_SubcommandRegistration(t *testing.T) {
	names := make([]string, 0, len(rootCmd.Commands()))
	for _, cmd := range rootCmd.Commands() {
		names = append(names, cmd.Name())
	}

	assert.Contains(t, names, "serve")
	assert.Contains(t, names, "migrate")
	assert.Contains(t, names, "version")
}

func TestRootCommand_ConfigFlag(t *testing.T) {
	f := rootCmd.PersistentFlags().Lookup("config")
	require.NotNil(t, f)
	assert.Equal(t, "", f.DefValue)
}

func TestMigrateCommand_SubcommandRegistration(t *testing.T) {
	names := make([]string, 0, len(migrateCmd.Commands()))
	for _, cmd := range migrateCmd.Commands() {
		names = append(names, cmd.Name())
	}

	assert.Contains(t, names, "up")
	assert.Contains(t, names, "down")
	assert.Contains(t, names, "status")
	assert.Contains(t, names, "init")
}

func TestServeCommand_Flags(t *testing.T) {
	portFlag := serveCmd.Flags().Lookup("port")
	require.NotNil(t, portFlag)
	assert.Equal(t, "p", portFlag.Shorthand)

	modeFlag := serveCmd.Flags().Lookup("mode")
	require.NotNil(t, modeFlag)
}

// --- Version command test ---

func TestVersionCommand_Output(t *testing.T) {
	buf := new(bytes.Buffer)
	versionCmd.SetOut(buf)

	// Call Run directly to avoid Cobra's root command dispatch
	versionCmd.Run(versionCmd, []string{})

	output := buf.String()
	assert.Contains(t, output, "Sky Flux CMS")
	assert.Contains(t, output, version)
	assert.Contains(t, output, commit)
	assert.Contains(t, output, date)
}

// --- initLogger tests ---
// Note: these tests modify global slog.Default() state.
// Go runs tests in a single package sequentially by default, so no race.

func TestInitLogger_Levels(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		expected slog.Level
	}{
		{"debug", "debug", slog.LevelDebug},
		{"info", "info", slog.LevelInfo},
		{"warn", "warn", slog.LevelWarn},
		{"error", "error", slog.LevelError},
		{"unknown_defaults_to_info", "unknown", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Log: config.LogConfig{Level: tt.level, Format: "text"},
			}
			initLogger(cfg)

			ctx := context.Background()
			assert.True(t, slog.Default().Enabled(ctx, tt.expected),
				"level %s should be enabled", tt.expected)
			if tt.expected > slog.LevelDebug {
				assert.False(t, slog.Default().Enabled(ctx, tt.expected-1),
					"level below %s should be disabled", tt.expected)
			}
		})
	}
}

func TestInitLogger_Formats(t *testing.T) {
	tests := []struct {
		name   string
		format string
		want   string
	}{
		{"json", "json", "*slog.JSONHandler"},
		{"text", "text", "*slog.TextHandler"},
		{"default_is_text", "", "*slog.TextHandler"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Log: config.LogConfig{Level: "info", Format: tt.format},
			}
			initLogger(cfg)

			handler := slog.Default().Handler()
			assert.Equal(t, tt.want, fmt.Sprintf("%T", handler))
		})
	}
}
