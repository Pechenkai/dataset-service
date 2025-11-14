package postqbuild

import (
	"context"
	"errors"
	"github.com/Masterminds/squirrel"
	"go.uber.org/zap"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type UserRepo struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

type userRow struct {
	ID               uint64
	Username         string
	Email            string
	Password         string
	RegistrationDate time.Time
	Country          string
	IsBlocked        bool
	Role             string
}

func userRowFromEntity(u *entities.User) userRow {
	if u == nil {
		return userRow{}
	}
	return userRow{
		ID:               u.ID,
		Username:         u.Username,
		Email:            u.Email,
		Password:         u.Password,
		RegistrationDate: u.RegistrationDate,
		Country:          u.Country,
		IsBlocked:        u.IsBlocked,
		Role:             u.Role,
	}
}

func (row userRow) toEntity() *entities.User {
	return &entities.User{
		ID:               row.ID,
		Username:         row.Username,
		Email:            row.Email,
		Password:         row.Password,
		RegistrationDate: row.RegistrationDate,
		Country:          row.Country,
		IsBlocked:        row.IsBlocked,
		Role:             row.Role,
	}
}

func NewUserRepo(pool *pgxpool.Pool, logger *zap.Logger) *UserRepo {
	logger.Debug("NewUserRepo initialized")
	return &UserRepo{db: pool, logger: logger}
}

func (r *UserRepo) Create(ctx context.Context, u *entities.User) error {
	row := userRowFromEntity(u)
	if row.RegistrationDate.IsZero() {
		row.RegistrationDate = time.Now().UTC()
		u.RegistrationDate = row.RegistrationDate
	}

	r.logger.Debug("Create User called",
		zap.String("username", row.Username),
		zap.String("email", row.Email),
		zap.String("country", row.Country),
		zap.Bool("is_blocked", row.IsBlocked),
		zap.String("role", row.Role),
	)

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
			row.Username,
			row.Email,
			row.Password,
			row.RegistrationDate,
			row.Country,
			row.IsBlocked,
			row.Role,
		).
		Suffix("RETURNING id")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build create user SQL",
			zap.Error(err),
		)
		return repositories.ErrUserQueryBuild
	}

	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(&row.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			r.logger.Warn("duplicate email on create user",
				zap.String("email", u.Email),
				zap.String("constraint", pgErr.ConstraintName),
			)
			return repositories.ErrEmailAlreadyExists
		}
		r.logger.Error("failed to execute create user query",
			zap.Error(err),
			zap.String("email", u.Email),
		)
		return repositories.ErrUserCreate
	}

	u.ID = row.ID
	r.logger.Info("user created successfully",
		zap.Uint64("id", row.ID),
		zap.String("email", row.Email),
	)
	return nil
}

func (r *UserRepo) Update(ctx context.Context, u *entities.User) error {
	row := userRowFromEntity(u)
	r.logger.Debug("Update User called",
		zap.Uint64("id", row.ID),
		zap.String("username", row.Username),
		zap.String("email", row.Email),
		zap.String("country", row.Country),
		zap.Bool("is_blocked", row.IsBlocked),
		zap.String("role", row.Role),
	)

	query := psql.
		Update("users").
		Set("username", row.Username).
		Set("email", row.Email).
		Set("password", row.Password).
		Set("country", row.Country).
		Set("is_blocked", row.IsBlocked).
		Set("role", row.Role).
		Where(squirrel.Eq{"id": row.ID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build update user SQL",
			zap.Error(err),
			zap.Uint64("id", row.ID),
		)
		return repositories.ErrUserQueryBuild
	}

	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		r.logger.Error("failed to execute update user query",
			zap.Error(err),
			zap.Uint64("id", row.ID),
		)
		return repositories.ErrUserUpdate
	}
	if cmd.RowsAffected() == 0 {
		r.logger.Warn("no user found to update",
			zap.Uint64("id", row.ID),
		)
		return repositories.ErrUserNotFound
	}

	r.logger.Info("user updated successfully",
		zap.Uint64("id", row.ID),
		zap.String("email", row.Email),
	)
	return nil
}

func (r *UserRepo) Delete(ctx context.Context, id uint64) error {
	r.logger.Debug("Delete User called",
		zap.Uint64("id", id),
	)

	query := psql.Delete("users").Where(squirrel.Eq{"id": id})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build delete user SQL",
			zap.Error(err),
			zap.Uint64("id", id),
		)
		return repositories.ErrUserQueryBuild
	}

	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		r.logger.Error("failed to execute delete user query",
			zap.Error(err),
			zap.Uint64("id", id),
		)
		return repositories.ErrUserDelete
	}
	if cmd.RowsAffected() == 0 {
		r.logger.Warn("no user found to delete",
			zap.Uint64("id", id),
		)
		return repositories.ErrUserNotFound
	}

	r.logger.Info("user deleted successfully",
		zap.Uint64("id", id),
	)
	return nil
}

func (r *UserRepo) FindByID(ctx context.Context, id uint64) (*entities.User, error) {
	r.logger.Debug("FindByID User called",
		zap.Uint64("id", id),
	)

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
		r.logger.Error("failed to build find user by ID SQL",
			zap.Error(err),
			zap.Uint64("id", id),
		)
		return nil, repositories.ErrUserQueryBuild
	}

	row := &userRow{}
	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(
		&row.ID,
		&row.Username,
		&row.Email,
		&row.Password,
		&row.RegistrationDate,
		&row.Country,
		&row.IsBlocked,
		&row.Role,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Warn("user not found by ID",
				zap.Uint64("id", id),
			)
			return nil, repositories.ErrUserNotFound
		}
		r.logger.Error("failed to execute find user by ID query",
			zap.Error(err),
			zap.Uint64("id", id),
		)
		return nil, repositories.ErrUserGet
	}

	entity := row.toEntity()
	r.logger.Info("user fetched successfully",
		zap.Uint64("id", entity.ID),
		zap.String("email", entity.Email),
		zap.String("username", entity.Username),
	)
	return entity, nil
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	r.logger.Debug("FindByEmail User called",
		zap.String("email", email),
	)

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
		r.logger.Error("failed to build find user by email SQL",
			zap.Error(err),
			zap.String("email", email),
		)
		return nil, repositories.ErrUserQueryBuild
	}

	row := &userRow{}
	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(
		&row.ID,
		&row.Username,
		&row.Email,
		&row.Password,
		&row.RegistrationDate,
		&row.Country,
		&row.IsBlocked,
		&row.Role,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Warn("user not found by email",
				zap.String("email", email),
			)
			return nil, repositories.ErrUserNotFound
		}
		r.logger.Error("failed to execute find user by email query",
			zap.Error(err),
			zap.String("email", email),
		)
		return nil, repositories.ErrUserGet
	}

	entity := row.toEntity()
	r.logger.Info("user fetched successfully by email",
		zap.Uint64("id", entity.ID),
		zap.String("email", entity.Email),
		zap.String("username", entity.Username),
	)
	return entity, nil
}

func (r *UserRepo) FindAll(ctx context.Context) ([]*entities.User, error) {
	r.logger.Debug("FindAll Users called")

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
		r.logger.Error("failed to build find all users SQL",
			zap.Error(err),
		)
		return nil, repositories.ErrUserQueryBuild
	}

	rows, err := r.db.Query(ctx, sqlStr, args...)
	if err != nil {
		r.logger.Error("failed to execute find all users query",
			zap.Error(err),
		)
		return nil, repositories.ErrUserGetAll
	}
	defer rows.Close()

	var list []*entities.User
	for rows.Next() {
		row := userRow{}
		if err := rows.Scan(
			&row.ID,
			&row.Username,
			&row.Email,
			&row.Password,
			&row.RegistrationDate,
			&row.Country,
			&row.IsBlocked,
			&row.Role,
		); err != nil {
			r.logger.Error("failed to scan user row",
				zap.Error(err),
			)
			return nil, repositories.ErrUserScan
		}
		list = append(list, row.toEntity())
	}
	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating over user rows",
			zap.Error(err),
		)
		return nil, repositories.ErrUserGetAll
	}

	r.logger.Info("all users fetched successfully",
		zap.Int("count", len(list)),
	)
	return list, nil
}
