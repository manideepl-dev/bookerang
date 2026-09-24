package services

import (
	"bookerang/internal/models"
	"bookerang/internal/repository"
	"context"
)

type BookService struct {
	repo *repository.BookRepository
}

func NewBookService(repo *repository.BookRepository) *BookService {
	return &BookService{repo: repo}
}

func (s *BookService) AddBook(ctx context.Context, req models.AddBookRequest, username string) (models.AddBookResult, error) {
	return s.repo.AddBook(ctx, req.Author, req.Title, username)
}

func (s *BookService) MyBooks(ctx context.Context, username string) ([]models.CopyDTO, error) {
	return s.repo.MyBooks(ctx, username)
}

func (s *BookService) NearbyBooks(ctx context.Context, username string, radius int64) ([]models.NearbyBookDTO, error) {
	return s.repo.NearbyBooks(ctx, username, radius)
}
