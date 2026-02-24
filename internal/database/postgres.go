package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"

	"github.com/sky-flux/cms/internal/config"
)

func NewPostgres(cfg *config.Config) (*bun.DB, error) {
	connector := pgdriver.NewConnector(
		pgdriver.WithDSN(cfg.DB.DSN()),
	)

	sqlDB := sql.OpenDB(connector)
	sqlDB.SetMaxOpenConns(cfg.DB.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.DB.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.DB.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.DB.ConnMaxIdleTime)

	db := bun.NewDB(sqlDB, pgdialect.New(),
		bun.WithDiscardUnknownColumns(),
	)

	db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(cfg.Server.Mode == "debug"),
	))

	if err := db.PingContext(context.Background()); err != nil {
		return nil, fmt.Errorf("postgres ping: %w", err)
	}

	return db, nil
}
