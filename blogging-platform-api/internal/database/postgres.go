package database

import (
	"blogging-platform-api/internal/config"
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var db *sql.DB

func Open(cfg *config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host,
		cfg.Port,
		cfg.Username,
		cfg.Password,
		cfg.DBName,
	)

	var err error
	db, err = sql.Open("pgx", dsn)
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
