package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/xenptr/go-projects/blogging-platform-api/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Open(cfg *config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host,
		cfg.Port,
		cfg.Username,
		cfg.Password,
		cfg.DBName,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)

	return db, nil
}
