package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/karzhen/restaurant-lk/internal/domain"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (
			id, email, password_hash, first_name, last_name,
			middle_name, phone, age, role, is_active, created_at, updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.MiddleName,
		user.Phone,
		user.Age,
		user.Role,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	)
	return err
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name,
		       middle_name, phone, age, role, is_active, created_at, updated_at
		FROM users
		WHERE lower(email) = lower($1)
	`

	user, err := scanUser(r.db.QueryRow(ctx, query, email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name,
		       middle_name, phone, age, role, is_active, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user, err := scanUser(r.db.QueryRow(ctx, query, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) UpdateProfile(ctx context.Context, userID uuid.UUID, input domain.UpdateUserProfile) (*domain.User, error) {
	setParts := make([]string, 0, 6)
	args := make([]any, 0, 7)
	index := 1

	if input.FirstName != nil {
		setParts = append(setParts, fmt.Sprintf("first_name = $%d", index))
		args = append(args, strings.TrimSpace(*input.FirstName))
		index++
	}
	if input.LastName != nil {
		setParts = append(setParts, fmt.Sprintf("last_name = $%d", index))
		args = append(args, strings.TrimSpace(*input.LastName))
		index++
	}
	if input.MiddleName != nil {
		setParts = append(setParts, fmt.Sprintf("middle_name = $%d", index))
		trimmed := strings.TrimSpace(*input.MiddleName)
		if trimmed == "" {
			args = append(args, nil)
		} else {
			args = append(args, trimmed)
		}
		index++
	}
	if input.Phone != nil {
		setParts = append(setParts, fmt.Sprintf("phone = $%d", index))
		trimmed := strings.TrimSpace(*input.Phone)
		if trimmed == "" {
			args = append(args, nil)
		} else {
			args = append(args, trimmed)
		}
		index++
	}
	if input.Age != nil {
		setParts = append(setParts, fmt.Sprintf("age = $%d", index))
		args = append(args, *input.Age)
		index++
	}

	if len(setParts) == 0 {
		return r.GetByID(ctx, userID)
	}

	setParts = append(setParts, "updated_at = now()")
	args = append(args, userID)

	query := fmt.Sprintf(`
		UPDATE users
		SET %s
		WHERE id = $%d
		RETURNING id, email, password_hash, first_name, last_name,
		          middle_name, phone, age, role, is_active, created_at, updated_at
	`, strings.Join(setParts, ", "), index)

	updated, err := scanUser(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return updated, nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = now() WHERE id = $2`
	result, err := r.db.Exec(ctx, query, passwordHash, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func scanUser(row pgx.Row) (*domain.User, error) {
	var (
		user       domain.User
		middleName sql.NullString
		phone      sql.NullString
		age        sql.NullInt32
	)

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&middleName,
		&phone,
		&age,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if middleName.Valid {
		value := middleName.String
		user.MiddleName = &value
	}
	if phone.Valid {
		value := phone.String
		user.Phone = &value
	}
	if age.Valid {
		value := int(age.Int32)
		user.Age = &value
	}

	return &user, nil
}
