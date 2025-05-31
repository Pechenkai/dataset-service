package postqbuild

import (
	"context"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"ppo/internal/entities"
	"ppo/internal/repositories"
	"time"
)

type SubscriptionRepo struct {
	db *pgxpool.Pool
}

func NewSubscriptionRepo(pool *pgxpool.Pool) *SubscriptionRepo {
	return &SubscriptionRepo{db: pool}
}

func (r *SubscriptionRepo) Create(ctx context.Context, s *entities.Subscription) error {
	// Если дата создания не задана, инициализируем текущим временем
	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now().UTC()
	}

	// Построение SQL через squirrel
	query := sq.
		Insert("subscriptions").
		Columns("user_id", "dataset_id", "created_at").
		Values(s.UserID, s.DatasetID, s.CreatedAt)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		// Ошибка построения SQL – возвращаем как есть
		return fmt.Errorf("build insert subscription sql: %w", err)
	}

	// Выполняем вставку
	_, err = r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		// Если это ошибка дубликата (unique constraint violation)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return repositories.ErrAlreadySubscribed
		}
		// Любая другая ошибка – оборачиваем
		return fmt.Errorf("subscribe: %w", err)
	}

	return nil
}

func (r *SubscriptionRepo) Unsubscribe(ctx context.Context, userID, datasetID uint64) error {
	// Построение SQL удаления
	query := sq.
		Delete("subscriptions").
		Where(sq.Eq{"user_id": userID, "dataset_id": datasetID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build delete subscription sql: %w", err)
	}

	// Выполняем удаление
	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		return fmt.Errorf("unsubscribe: %w", err)
	}

	// Если ни одна строка не была удалена – возвращаем ErrSubscriptionNotFound
	if cmd.RowsAffected() == 0 {
		return repositories.ErrSubscriptionNotFound
	}

	return nil
}

func (r *SubscriptionRepo) IsSubscribed(ctx context.Context, userID, datasetID uint64) (bool, error) {
	// Построение SQL для проверки наличия записи
	query := sq.
		Select("1").
		From("subscriptions").
		Where(sq.Eq{"user_id": userID, "dataset_id": datasetID}).
		Limit(1)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return false, fmt.Errorf("build is_subscribed sql: %w", err)
	}

	var dummy int
	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(&dummy)
	if err != nil {
		// Если нет строк – значит не подписан, возвращаем (false, nil)
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		// Любая другая ошибка – оборачиваем
		return false, fmt.Errorf("is subscribed: %w", err)
	}
	return true, nil
}

func (r *SubscriptionRepo) GetSubscribers(ctx context.Context, datasetID uint64) ([]uint64, error) {
	// Построение SQL для получения списка подписчиков
	query := sq.
		Select("user_id").
		From("subscriptions").
		Where(sq.Eq{"dataset_id": datasetID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build get subscribers sql: %w", err)
	}

	rows, err := r.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("get subscribers: %w", err)
	}
	defer rows.Close()

	var list []uint64
	for rows.Next() {
		var uid uint64
		if err := rows.Scan(&uid); err != nil {
			return nil, fmt.Errorf("scan subscriber row: %w", err)
		}
		list = append(list, uid)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate subscriber rows: %w", err)
	}

	return list, nil
}

func (r *SubscriptionRepo) GetByUser(ctx context.Context, userID uint64) ([]uint64, error) {
	// Построение SQL для получения всех подписок данного пользователя
	query := sq.
		Select("dataset_id").
		From("subscriptions").
		Where(sq.Eq{"user_id": userID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build get subscriptions by user sql: %w", err)
	}

	rows, err := r.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("get_subscriptions_by_user: %w", err)
	}
	defer rows.Close()

	var datasets []uint64
	for rows.Next() {
		var dsid uint64
		if err := rows.Scan(&dsid); err != nil {
			return nil, fmt.Errorf("scan subscription: %w", err)
		}
		datasets = append(datasets, dsid)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate subscription rows: %w", err)
	}

	return datasets, nil
}
