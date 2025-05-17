package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"ppo/internal/config"
)

func NewPool(ctx context.Context, dbCfg config.Database) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(dbCfg.DSN)
	if err != nil {
		return nil, err
	}

	poolCfg.MaxConns = dbCfg.MaxConns
	poolCfg.MinConns = dbCfg.MinConns
	poolCfg.MaxConnIdleTime = time.Duration(dbCfg.MaxConnIdleTime) * time.Second
	poolCfg.HealthCheckPeriod = time.Duration(dbCfg.HealthCheckPeriod) * time.Second
	poolCfg.ConnConfig.ConnectTimeout = time.Duration(dbCfg.ConnectTimeout) * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func ClosePool(pool *pgxpool.Pool) {
	pool.Close()
}
