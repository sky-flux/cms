package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/uptrace/bun/migrate"

	"github.com/sky-flux/cms/internal/config"
	"github.com/sky-flux/cms/internal/database"
	"github.com/sky-flux/cms/migrations"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "数据库迁移管理",
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "执行所有待迁移",
	RunE:  runMigrateUp,
}

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "回滚最近一组迁移",
	RunE:  runMigrateDown,
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看迁移状态",
	RunE:  runMigrateStatus,
}

var migrateInitCmd = &cobra.Command{
	Use:   "init",
	Short: "创建迁移元数据表 (bun_migrations, bun_migration_locks)",
	RunE:  runMigrateInit,
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.AddCommand(migrateUpCmd, migrateDownCmd, migrateStatusCmd, migrateInitCmd)
}

func newMigrator(cfg *config.Config) (*migrate.Migrator, func(), error) {
	db, err := database.NewPostgres(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("connect to postgres: %w", err)
	}
	cleanup := func() { db.Close() }
	return migrate.NewMigrator(db, migrations.Migrations), cleanup, nil
}

func runMigrateInit(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	initLogger(cfg)

	migrator, cleanup, err := newMigrator(cfg)
	if err != nil {
		return err
	}
	defer cleanup()

	ctx := context.Background()
	if err := migrator.Init(ctx); err != nil {
		return fmt.Errorf("migrate init: %w", err)
	}
	slog.Info("migration tables created")
	return nil
}

func runMigrateUp(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	initLogger(cfg)

	migrator, cleanup, err := newMigrator(cfg)
	if err != nil {
		return err
	}
	defer cleanup()

	ctx := context.Background()

	// Ensure migration tables exist
	if err := migrator.Init(ctx); err != nil {
		return fmt.Errorf("migrate init: %w", err)
	}

	group, err := migrator.Migrate(ctx)
	if err != nil {
		return fmt.Errorf("migrate up: %w", err)
	}

	if group.IsZero() {
		slog.Info("no new migrations to run")
	} else {
		slog.Info("migrated", "group", group)
	}
	return nil
}

func runMigrateDown(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	initLogger(cfg)

	migrator, cleanup, err := newMigrator(cfg)
	if err != nil {
		return err
	}
	defer cleanup()

	ctx := context.Background()

	group, err := migrator.Rollback(ctx)
	if err != nil {
		return fmt.Errorf("migrate down: %w", err)
	}

	if group.IsZero() {
		slog.Info("no migrations to rollback")
	} else {
		slog.Info("rolled back", "group", group)
	}
	return nil
}

func runMigrateStatus(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	initLogger(cfg)

	migrator, cleanup, err := newMigrator(cfg)
	if err != nil {
		return err
	}
	defer cleanup()

	ctx := context.Background()

	ms, err := migrator.MigrationsWithStatus(ctx)
	if err != nil {
		return fmt.Errorf("migrate status: %w", err)
	}

	fmt.Printf("migrations: %s\n", ms)
	fmt.Printf("unapplied:  %s\n", ms.Unapplied())
	fmt.Printf("last group: %s\n", ms.LastGroup())
	return nil
}
