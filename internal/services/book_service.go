package services

import (
	"bookerang/internal/domain"
	"bookerang/internal/repository"
	"context"
)

type BookService struct {
	repo *repository.BookRepository
}

type AddBookInput struct {
	Title  string
	Author string
}

func NewBookService(repo *repository.BookRepository) *BookService {
	return &BookService{repo: repo}
}

func (s *BookService) AddBook(ctx context.Context, input AddBookInput, username string) (domain.AddBookResult, error) {
	return s.repo.AddBook(ctx, input.Author, input.Title, username)
}

func (s *BookService) MyBooks(ctx context.Context, username string) ([]domain.Copy, error) {
	return s.repo.MyBooks(ctx, username)
}

func (s *BookService) NearbyBooks(ctx context.Context, username string, radius int64) ([]domain.NearbyBook, error) {
	return s.repo.NearbyBooks(ctx, username, radius)
}
