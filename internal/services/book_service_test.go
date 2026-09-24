package services

import (
	"context"
	"errors"
	"testing"

	"bookerang/internal/domain"
)

type fakeBookRepository struct {
	addBookCalled bool
}

func (f *fakeBookRepository) AddBook(context.Context, string, string, string) (domain.AddBookResult, error) {
	f.addBookCalled = true
	return domain.AddBookResult{CopyID: "copy-1", Added: true}, nil
}

func (f *fakeBookRepository) MyBooks(context.Context, string) ([]domain.Copy, error) {
	return nil, nil
}

func (f *fakeBookRepository) NearbyBooks(context.Context, string, int64) ([]domain.NearbyBook, error) {
	return nil, nil
}

func TestAddBookRejectsEmptyFields(t *testing.T) {
	repo := &fakeBookRepository{}
	service := NewBookService(repo)

	_, err := service.AddBook(context.Background(), AddBookInput{Title: "", Author: "Author"}, "alice")
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("error = %v, want ErrInvalidInput", err)
	}
	if repo.addBookCalled {
		t.Fatal("repository should not be called for invalid input")
	}
}

func TestNearbyBooksRejectsOutOfRangeRadius(t *testing.T) {
	service := NewBookService(&fakeBookRepository{})

	_, err := service.NearbyBooks(context.Background(), "alice", 1001)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("error = %v, want ErrInvalidInput", err)
	}
}
