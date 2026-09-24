package services

import (
	"context"
	"errors"
	"fmt"

	"bookerang/internal/domain"
	jwtutil "bookerang/internal/jwt"
	"bookerang/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo      *repository.UserRepository
	jwtSecret string
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

func NewUserService(repo *repository.UserRepository, jwtSecret string) *UserService {
	return &UserService{repo: repo, jwtSecret: jwtSecret}
}

func (s *UserService) Login(ctx context.Context, input LoginInput) (string, error) {
	user, exists, err := s.repo.FindByUsername(ctx, input.Username)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", errors.New("user not found")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return "", errors.New("invalid credentials")
	}
	return jwtutil.Generate(s.jwtSecret, user.Username)
}

func (s *UserService) Signup(ctx context.Context, input SignupInput) (string, error) {
	_, exists, err := s.repo.FindByUsername(ctx, input.Username)
	if err != nil {
		return "", err
	}
	if exists {
		return "", errors.New("user already exists")
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
		return "", err
	}
	return jwtutil.Generate(s.jwtSecret, user.Username)
}
