package repository

import (
	"context"
	"database/sql"
	"fmt"

	"bookerang/internal/domain"
	"github.com/lib/pq"
)

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (domain.User, bool, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT username, password, first_name, last_name
		FROM users
		WHERE username = $1
	`, username)

	var user domain.User
	if err := row.Scan(&user.Username, &user.Password, &user.FirstName, &user.LastName); err != nil {
		if err == sql.ErrNoRows {
			return domain.User{}, false, nil
		}
		return domain.User{}, false, fmt.Errorf("find user by username: %w", err)
	}
	return user, true, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, user domain.User) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO users (first_name, last_name, username, password, location)
		VALUES ($1, $2, $3, $4, ST_SetSRID(ST_MakePoint($5, $6), 4326)::geography)
	`, user.FirstName, user.LastName, user.Username, user.Password, user.Longitude, user.Latitude)
	if err != nil {
		if postgresErr, ok := err.(*pq.Error); ok && postgresErr.Code == "23505" {
			return fmt.Errorf("%w: username already exists", ErrDuplicate)
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}
