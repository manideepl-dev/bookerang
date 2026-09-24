package models

import "time"

// Request payloads

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SignupRequest struct {
	Username  string  `json:"username"`
	Password  string  `json:"password"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type AddBookRequest struct {
	Title  string `json:"title"`
	Author string `json:"author"`
}

// Response payloads

type TokenResponse struct {
	Token string `json:"token"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type IdResponse struct {
	Msg string `json:"msg"`
	ID  string `json:"id"`
}

type AddBookResult struct {
	CopyID string `json:"copyId"`
	Added  bool   `json:"added"`
}

type MyBooksResponse struct {
	Books []CopyDTO `json:"books"`
}

type NearbyBooksResponse struct {
	Books []NearbyBookDTO `json:"books"`
}

type CopyDTO struct {
	CopyID string `json:"copyId"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

type NearbyBookDTO struct {
	CopyID         string `json:"copyId"`
	Title          string `json:"title"`
	Author         string `json:"author"`
	OwnerUsername  string `json:"ownerUsername"`
	OwnerFirstName string `json:"ownerFirstName"`
	OwnerLastName  string `json:"ownerLastName"`
}

// Domain entities

type User struct {
	Username  string  `db:"username"`
	Password  string  `db:"password"`
	FirstName string  `db:"first_name"`
	LastName  string  `db:"last_name"`
	Latitude  float64 `db:"latitude"`
	Longitude float64 `db:"longitude"`
}

type Book struct {
	Title     string    `db:"title"`
	AuthorID  string    `db:"author_id"`
	BookID    string    `db:"book_id"`
	CreatedAt time.Time `db:"created_at"`
}
