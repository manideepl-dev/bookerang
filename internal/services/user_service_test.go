package services

import (
	"context"
	"errors"
	"testing"

	"bookerang/internal/domain"
	"bookerang/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type fakeUserRepository struct {
	user        domain.User
	exists      bool
	createErr   error
	createdUser domain.User
}

func (f *fakeUserRepository) FindByUsername(context.Context, string) (domain.User, bool, error) {
	return f.user, f.exists, nil
}

func (f *fakeUserRepository) CreateUser(_ context.Context, user domain.User) error {
	f.createdUser = user
	return f.createErr
}

func TestSignupRejectsInvalidInput(t *testing.T) {
	repo := &fakeUserRepository{}
	service := NewUserService(repo, "secret")

	_, err := service.Signup(context.Background(), SignupInput{
		Username: "alice",
		Password: "short",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("error = %v, want ErrInvalidInput", err)
	}
	if repo.createdUser.Username != "" {
		t.Fatal("repository should not be called for invalid input")
	}
}

func TestSignupHashesPasswordAndCreatesUser(t *testing.T) {
	repo := &fakeUserRepository{}
	service := NewUserService(repo, "secret")

	_, err := service.Signup(context.Background(), SignupInput{
		Username:  "alice",
		Password:  "long-enough-password",
		FirstName: "Alice",
		LastName:  "Reader",
		Latitude:  40,
		Longitude: -73,
	})
	if err != nil {
		t.Fatalf("signup: %v", err)
	}
	if repo.createdUser.Password == "long-enough-password" {
		t.Fatal("password was stored in plaintext")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(repo.createdUser.Password), []byte("long-enough-password")); err != nil {
		t.Fatalf("stored password is not a valid hash: %v", err)
	}
}

func TestSignupMapsDuplicateUser(t *testing.T) {
	repo := &fakeUserRepository{createErr: repository.ErrDuplicate}
	service := NewUserService(repo, "secret")

	_, err := service.Signup(context.Background(), SignupInput{
		Username:  "alice",
		Password:  "long-enough-password",
		FirstName: "Alice",
		LastName:  "Reader",
		Latitude:  40,
		Longitude: -73,
	})
	if !errors.Is(err, ErrUserExists) {
		t.Fatalf("error = %v, want ErrUserExists", err)
	}
}
