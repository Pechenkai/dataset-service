package postqbuild

import (
	"context"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"ppo/internal/entities"
	"ppo/internal/repositories"
)

var ErrEmailAlreadyExists = errors.New("email already exists")

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: pool}
}

func (r *UserRepo) Create(ctx context.Context, u *entities.User) error {
	if u.RegistrationDate.IsZero() {
		u.RegistrationDate = time.Now().UTC()
	}
	query := psql.
		Insert("users").
		Columns("username", "email", "password", "registration_date", "country", "is_blocked", "role").
		Values(u.Username, u.Email, u.Password, u.RegistrationDate, u.Country, u.IsBlocked, u.Role).
		Suffix("RETURNING id")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build insert user sql: %w", err)
	}

	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(&u.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrEmailAlreadyExists
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *UserRepo) Update(ctx context.Context, u *entities.User) error {
	query := psql.
		Update("users").
		Set("username", u.Username).
		Set("email", u.Email).
		Set("password", u.Password).
		Set("country", u.Country).
		Set("is_blocked", u.IsBlocked).
		Set("role", u.Role).
		Where(sq.Eq{"id": u.ID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build update user sql: %w", err)
	}

	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return repositories.ErrUserNotFound
	}
	return nil
}

func (r *UserRepo) Delete(ctx context.Context, id uint64) error {
	query := psql.Delete("users").Where(sq.Eq{"id": id})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build delete user sql: %w", err)
	}

	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return repositories.ErrUserNotFound
	}
	return nil
}

func (r *UserRepo) FindByID(ctx context.Context, id uint64) (*entities.User, error) {
	query := psql.
		Select("id", "username", "email", "password", "registration_date", "country", "is_blocked", "role").
		From("users").
		Where(sq.Eq{"id": id})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build find user by id sql: %w", err)
	}

	u := &entities.User{}
	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(
		&u.ID, &u.Username, &u.Email, &u.Password, &u.RegistrationDate, &u.Country, &u.IsBlocked, &u.Role,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repositories.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return u, nil
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	query := psql.
		Select("id", "username", "email", "password", "registration_date", "country", "is_blocked", "role").
		From("users").
		Where(sq.Eq{"email": email})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build find user by email sql: %w", err)
	}

	u := &entities.User{}
	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(
		&u.ID, &u.Username, &u.Email, &u.Password, &u.RegistrationDate, &u.Country, &u.IsBlocked, &u.Role,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repositories.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return u, nil
}

func (r *UserRepo) FindAll(ctx context.Context) ([]*entities.User, error) {
	query := psql.
		Select("id", "username", "email", "password", "registration_date", "country", "is_blocked", "role").
		From("users").
		OrderBy("registration_date DESC")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build find all users sql: %w", err)
	}

	rows, err := r.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("query all users: %w", err)
	}
	defer rows.Close()

	var list []*entities.User
	for rows.Next() {
		u := &entities.User{}
		if err := rows.Scan(
			&u.ID, &u.Username, &u.Email, &u.Password, &u.RegistrationDate, &u.Country, &u.IsBlocked, &u.Role,
		); err != nil {
			return nil, fmt.Errorf("scan user row: %w", err)
		}
		list = append(list, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user rows: %w", err)
	}
	return list, nil
}
