package repository

import (
	"context"
	"database/sql"
	"fmt"

	"bookerang/internal/models"
)

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (models.User, bool, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT username, password, first_name, last_name
		FROM users
		WHERE username = $1
	`, username)

	var user models.User
	if err := row.Scan(&user.Username, &user.Password, &user.FirstName, &user.LastName); err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, false, nil
		}
		return models.User{}, false, fmt.Errorf("find user by username: %w", err)
	}
	return user, true, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, user models.User) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO users (first_name, last_name, username, password, location)
		VALUES ($1, $2, $3, $4, ST_SetSRID(ST_MakePoint($5, $6), 4326)::geography)
	`, user.FirstName, user.LastName, user.Username, user.Password, user.Longitude, user.Latitude)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}
