package postqbuild

import (
	"context"

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
	poolCfg.MaxConnIdleTime = dbCfg.MaxConnIdleTime
	poolCfg.HealthCheckPeriod = dbCfg.HealthCheckPeriod
	poolCfg.ConnConfig.ConnectTimeout = dbCfg.ConnectTimeout

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func ClosePool(pool *pgxpool.Pool) {
	pool.Close()
}
