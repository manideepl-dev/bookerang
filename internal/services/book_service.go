package services

import (
	"context"
	"strings"

	"bookerang/internal/domain"
)

type BookService struct {
	repo BookRepository
}

type BookRepository interface {
	AddBook(ctx context.Context, author string, title string, owner string) (domain.AddBookResult, error)
	MyBooks(ctx context.Context, username string) ([]domain.Copy, error)
	NearbyBooks(ctx context.Context, username string, radius int64) ([]domain.NearbyBook, error)
}

type AddBookInput struct {
	Title  string
	Author string
}

func NewBookService(repo BookRepository) *BookService {
	return &BookService{repo: repo}
}

func (s *BookService) AddBook(ctx context.Context, input AddBookInput, username string) (domain.AddBookResult, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Author = strings.TrimSpace(input.Author)
	if input.Title == "" || len(input.Title) > 255 || input.Author == "" || len(input.Author) > 255 {
		return domain.AddBookResult{}, ErrInvalidInput
	}
	return s.repo.AddBook(ctx, input.Author, input.Title, username)
}

func (s *BookService) MyBooks(ctx context.Context, username string) ([]domain.Copy, error) {
	return s.repo.MyBooks(ctx, username)
}

func (s *BookService) NearbyBooks(ctx context.Context, username string, radius int64) ([]domain.NearbyBook, error) {
	if radius <= 0 || radius > 1000 {
		return nil, ErrInvalidInput
	}
	return s.repo.NearbyBooks(ctx, username, radius)
}
