package repository

import (
	"context"
	"database/sql"
	"fmt"

	"bookerang/internal/domain"
)

type BookRepository struct {
	DB *sql.DB
}

func NewBookRepository(db *sql.DB) *BookRepository {
	return &BookRepository{DB: db}
}

func (r *BookRepository) AddBook(ctx context.Context, authorTitle string, title string, owner string) (domain.AddBookResult, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return domain.AddBookResult{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	var authorID string
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO authors (name)
		VALUES ($1)
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING author_id
	`, authorTitle).Scan(&authorID); err != nil {
		return domain.AddBookResult{}, fmt.Errorf("upsert author: %w", err)
	}

	var bookID string
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO books (title, author_id)
		VALUES ($1, $2)
		ON CONFLICT (title, author_id) DO UPDATE SET title = EXCLUDED.title
		RETURNING book_id
	`, title, authorID).Scan(&bookID); err != nil {
		return domain.AddBookResult{}, fmt.Errorf("upsert book: %w", err)
	}

	var copyID string
	var added bool
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO copies (book_id, owner_id)
		VALUES ($1, $2)
		ON CONFLICT (book_id, owner_id) DO UPDATE SET owner_id = EXCLUDED.owner_id
		RETURNING copy_id, xmax = 0 AS added
	`, bookID, owner).Scan(&copyID, &added); err != nil {
		return domain.AddBookResult{}, fmt.Errorf("upsert copy: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return domain.AddBookResult{}, fmt.Errorf("commit transaction: %w", err)
	}

	return domain.AddBookResult{CopyID: copyID, Added: added}, nil
}

func (r *BookRepository) MyBooks(ctx context.Context, username string) ([]domain.Copy, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT c.copy_id, b.title, a.name
		FROM copies c
		JOIN books b ON c.book_id = b.book_id
		JOIN authors a ON b.author_id = a.author_id
		WHERE c.owner_id = $1
	`, username)
	if err != nil {
		return nil, fmt.Errorf("list my books: %w", err)
	}
	defer rows.Close()

	books := make([]domain.Copy, 0)
	for rows.Next() {
		var item domain.Copy
		if err := rows.Scan(&item.CopyID, &item.Title, &item.Author); err != nil {
			return nil, fmt.Errorf("scan my books: %w", err)
		}
		books = append(books, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate my books: %w", err)
	}
	return books, nil
}

func (r *BookRepository) NearbyBooks(ctx context.Context, username string, radius int64) ([]domain.NearbyBook, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT c.copy_id,
		       b.title,
		       a.name,
		       owner.username AS owner_username,
		       owner.first_name AS owner_first_name,
		       owner.last_name AS owner_last_name
		FROM users requester
		JOIN users owner ON owner.username <> requester.username
		JOIN copies c ON c.owner_id = owner.username
		JOIN books b ON c.book_id = b.book_id
		JOIN authors a ON b.author_id = a.author_id
		WHERE requester.username = $1
		  AND requester.location IS NOT NULL
		  AND owner.location IS NOT NULL
		  AND ST_DWithin(owner.location, requester.location, $2 * 1000)
		ORDER BY ST_Distance(owner.location, requester.location), b.title
	`, username, radius)
	if err != nil {
		return nil, fmt.Errorf("list nearby books: %w", err)
	}
	defer rows.Close()

	books := make([]domain.NearbyBook, 0)
	for rows.Next() {
		var item domain.NearbyBook
		if err := rows.Scan(&item.CopyID, &item.Title, &item.Author, &item.OwnerUsername, &item.OwnerFirstName, &item.OwnerLastName); err != nil {
			return nil, fmt.Errorf("scan nearby books: %w", err)
		}
		books = append(books, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate nearby books: %w", err)
	}
	return books, nil
}
