package postqbuild

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"ppo/internal/repositories"

	"ppo/internal/entities"
)

var ErrEmailAlreadyExists = errors.New("email already exists")

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: pool}
}

func (r *UserRepo) Create(ctx context.Context, u *entities.User) error {
	const sql = `
	INSERT INTO users
	  (username, email, password, registration_date, country, is_blocked, role)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id
	`
	if u.RegistrationDate.IsZero() {
		u.RegistrationDate = time.Now().UTC()
	}
	err := r.db.QueryRow(ctx, sql,
		u.Username,
		u.Email,
		u.Password,
		u.RegistrationDate,
		u.Country,
		u.IsBlocked,
		u.Role,
	).Scan(&u.ID)
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
	const sql = `
	UPDATE users
	SET username = $1,
	    email = $2,
	    password = $3,
	    country = $4,
	    is_blocked = $5,
	    role = $6
	WHERE id = $7
	`
	cmd, err := r.db.Exec(ctx, sql,
		u.Username,
		u.Email,
		u.Password,
		u.Country,
		u.IsBlocked,
		u.Role,
		u.ID,
	)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return repositories.ErrUserNotFound
	}
	return nil
}

func (r *UserRepo) Delete(ctx context.Context, id uint64) error {
	const sql = `DELETE FROM users WHERE id = $1`
	exec, err := r.db.Exec(ctx, sql, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if exec.RowsAffected() == 0 {
		return repositories.ErrUserNotFound
	}
	return nil
}

func (r *UserRepo) FindByID(ctx context.Context, id uint64) (*entities.User, error) {
	const sql = `
	SELECT id, username, email, password, registration_date, country, is_blocked, role
	FROM users
	WHERE id = $1
	`
	u := &entities.User{}
	err := r.db.QueryRow(ctx, sql, id).Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.Password,
		&u.RegistrationDate,
		&u.Country,
		&u.IsBlocked,
		&u.Role,
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
	const sql = `
	SELECT id, username, email, password, registration_date, country, is_blocked, role
	FROM users
	WHERE email = $1
	`
	u := &entities.User{}
	err := r.db.QueryRow(ctx, sql, email).Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.Password,
		&u.RegistrationDate,
		&u.Country,
		&u.IsBlocked,
		&u.Role,
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
	const sql = `
	SELECT id, username, email, password, registration_date, country, is_blocked, role
	FROM users
	ORDER BY registration_date DESC
	`
	rows, err := r.db.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("query all users: %w", err)
	}
	defer rows.Close()

	var list []*entities.User
	for rows.Next() {
		u := &entities.User{}
		if err := rows.Scan(
			&u.ID,
			&u.Username,
			&u.Email,
			&u.Password,
			&u.RegistrationDate,
			&u.Country,
			&u.IsBlocked,
			&u.Role,
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
