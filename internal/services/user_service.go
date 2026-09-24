package services

import (
	"context"
	"errors"
	"fmt"

	jwtutil "bookerang/internal/jwt"
	"bookerang/internal/models"
	"bookerang/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo      *repository.UserRepository
	jwtSecret string
}

func NewUserService(repo *repository.UserRepository, jwtSecret string) *UserService {
	return &UserService{repo: repo, jwtSecret: jwtSecret}
}

func (s *UserService) Login(ctx context.Context, req models.LoginRequest) (string, error) {
	user, exists, err := s.repo.FindByUsername(ctx, req.Username)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", errors.New("user not found")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", errors.New("invalid credentials")
	}
	return jwtutil.Generate(s.jwtSecret, user.Username)
}

func (s *UserService) Signup(ctx context.Context, req models.SignupRequest) (string, error) {
	_, exists, err := s.repo.FindByUsername(ctx, req.Username)
	if err != nil {
		return "", err
	}
	if exists {
		return "", errors.New("user already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	user := models.User{
		Username:  req.Username,
		Password:  string(hash),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
	}
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return "", err
	}
	return jwtutil.Generate(s.jwtSecret, user.Username)
}
