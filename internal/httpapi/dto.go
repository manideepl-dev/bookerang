package httpapi

import "bookerang/internal/domain"

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type signupRequest struct {
	Username  string  `json:"username"`
	Password  string  `json:"password"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type addBookRequest struct {
	Title  string `json:"title"`
	Author string `json:"author"`
}

type tokenResponse struct {
	Token string `json:"token"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type idResponse struct {
	Msg string `json:"msg"`
	ID  string `json:"id"`
}

type copyResponse struct {
	CopyID string `json:"copyId"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

type nearbyBookResponse struct {
	CopyID         string `json:"copyId"`
	Title          string `json:"title"`
	Author         string `json:"author"`
	OwnerUsername  string `json:"ownerUsername"`
	OwnerFirstName string `json:"ownerFirstName"`
	OwnerLastName  string `json:"ownerLastName"`
}

type myBooksResponse struct {
	Books []copyResponse `json:"books"`
}

type nearbyBooksResponse struct {
	Books []nearbyBookResponse `json:"books"`
}

func toCopyResponses(copies []domain.Copy) []copyResponse {
	responses := make([]copyResponse, 0, len(copies))
	for _, copy := range copies {
		responses = append(responses, copyResponse{
			CopyID: copy.CopyID,
			Title:  copy.Title,
			Author: copy.Author,
		})
	}
	return responses
}

func toNearbyBookResponses(books []domain.NearbyBook) []nearbyBookResponse {
	responses := make([]nearbyBookResponse, 0, len(books))
	for _, book := range books {
		responses = append(responses, nearbyBookResponse{
			CopyID:         book.CopyID,
			Title:          book.Title,
			Author:         book.Author,
			OwnerUsername:  book.OwnerUsername,
			OwnerFirstName: book.OwnerFirstName,
			OwnerLastName:  book.OwnerLastName,
		})
	}
	return responses
}
