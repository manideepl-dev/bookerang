package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"bookerang/internal/domain"
	jwtutil "bookerang/internal/jwt"
	"bookerang/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo      UserRepository
	jwtSecret string
}

type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (domain.User, bool, error)
	CreateUser(ctx context.Context, user domain.User) error
}

type LoginInput struct {
	Username string
	Password string
}

type SignupInput struct {
	Username  string
	Password  string
	FirstName string
	LastName  string
	Latitude  float64
	Longitude float64
}

func NewUserService(repo UserRepository, jwtSecret string) *UserService {
	return &UserService{repo: repo, jwtSecret: jwtSecret}
}

func (s *UserService) Login(ctx context.Context, input LoginInput) (string, error) {
	input.Username = strings.TrimSpace(input.Username)
	if input.Username == "" || input.Password == "" {
		return "", ErrInvalidInput
	}

	user, exists, err := s.repo.FindByUsername(ctx, input.Username)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return "", ErrInvalidCredentials
	}
	return jwtutil.Generate(s.jwtSecret, user.Username)
}

func (s *UserService) Signup(ctx context.Context, input SignupInput) (string, error) {
	input.Username = strings.TrimSpace(input.Username)
	input.FirstName = strings.TrimSpace(input.FirstName)
	input.LastName = strings.TrimSpace(input.LastName)
	if input.Username == "" || len(input.Username) > 255 ||
		input.Password == "" || len(input.Password) < 8 ||
		input.FirstName == "" || len(input.FirstName) > 255 ||
		input.LastName == "" || len(input.LastName) > 255 ||
		math.IsNaN(input.Latitude) || math.IsInf(input.Latitude, 0) || input.Latitude < -90 || input.Latitude > 90 ||
		math.IsNaN(input.Longitude) || math.IsInf(input.Longitude, 0) || input.Longitude < -180 || input.Longitude > 180 {
		return "", ErrInvalidInput
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	user := domain.User{
		Username:  input.Username,
		Password:  string(hash),
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Latitude:  input.Latitude,
		Longitude: input.Longitude,
	}
	if err := s.repo.CreateUser(ctx, user); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return "", ErrUserExists
		}
		return "", err
	}
	return jwtutil.Generate(s.jwtSecret, user.Username)
}
