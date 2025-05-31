package postqbuild

import (
	"context"
	"errors"
	"github.com/Masterminds/squirrel"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"ppo/internal/entities"
	"ppo/internal/repositories"
)

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

	// Составляем INSERT ... RETURNING id
	query := psql.
		Insert("users").
		Columns(
			"username",
			"email",
			"password",
			"registration_date",
			"country",
			"is_blocked",
			"role",
		).
		Values(
			u.Username,
			u.Email,
			u.Password,
			u.RegistrationDate,
			u.Country,
			u.IsBlocked,
			u.Role,
		).
		Suffix("RETURNING id")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		// Ошибка построения запроса
		return repositories.ErrUserQueryBuild
	}

	// Выполняем запрос
	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(&u.ID)
	if err != nil {
		// Если дублирование по полю email (код 23505) → возвращаем ErrEmailAlreadyExists
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return repositories.ErrEmailAlreadyExists
		}
		// Любая другая ошибка при создании
		return repositories.ErrUserCreate
	}
	return nil
}

func (r *UserRepo) Update(ctx context.Context, u *entities.User) error {
	// Составляем UPDATE ... WHERE id = $?
	query := psql.
		Update("users").
		Set("username", u.Username).
		Set("email", u.Email).
		Set("password", u.Password).
		Set("country", u.Country).
		Set("is_blocked", u.IsBlocked).
		Set("role", u.Role).
		Where(squirrel.Eq{"id": u.ID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return repositories.ErrUserQueryBuild
	}

	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		// Локально не обрабатываем дублирование email: считаем, что
		// в Update не проверяем уникальность (её контролирует бизнес-логика)
		return repositories.ErrUserUpdate
	}
	if cmd.RowsAffected() == 0 {
		// Ничего не обновилось → пользователь не найден
		return repositories.ErrUserNotFound
	}
	return nil
}

func (r *UserRepo) Delete(ctx context.Context, id uint64) error {
	// Составляем DELETE FROM users WHERE id = $?
	query := psql.Delete("users").Where(squirrel.Eq{"id": id})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return repositories.ErrUserQueryBuild
	}

	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		return repositories.ErrUserDelete
	}
	if cmd.RowsAffected() == 0 {
		return repositories.ErrUserNotFound
	}
	return nil
}

func (r *UserRepo) FindByID(ctx context.Context, id uint64) (*entities.User, error) {
	// Составляем SELECT ... WHERE id = $?
	query := psql.
		Select(
			"id",
			"username",
			"email",
			"password",
			"registration_date",
			"country",
			"is_blocked",
			"role",
		).
		From("users").
		Where(squirrel.Eq{"id": id})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, repositories.ErrUserQueryBuild
	}

	u := &entities.User{}
	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(
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
		return nil, repositories.ErrUserGet
	}
	return u, nil
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	// Составляем SELECT ... WHERE email = $?
	query := psql.
		Select(
			"id",
			"username",
			"email",
			"password",
			"registration_date",
			"country",
			"is_blocked",
			"role",
		).
		From("users").
		Where(squirrel.Eq{"email": email})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, repositories.ErrUserQueryBuild
	}

	u := &entities.User{}
	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(
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
		return nil, repositories.ErrUserGet
	}
	return u, nil
}

func (r *UserRepo) FindAll(ctx context.Context) ([]*entities.User, error) {
	// Составляем SELECT ... ORDER BY registration_date DESC
	query := psql.
		Select(
			"id",
			"username",
			"email",
			"password",
			"registration_date",
			"country",
			"is_blocked",
			"role",
		).
		From("users").
		OrderBy("registration_date DESC")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, repositories.ErrUserQueryBuild
	}

	rows, err := r.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, repositories.ErrUserGetAll
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
			return nil, repositories.ErrUserScan
		}
		list = append(list, u)
	}
	if err := rows.Err(); err != nil {
		return nil, repositories.ErrUserGetAll
	}
	return list, nil
}
