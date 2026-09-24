package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"bookerang/internal/services"
)

type UserHandler struct {
	service *services.UserService
}

type BookHandler struct {
	service *services.BookService
}

func RegisterRoutes(router *http.ServeMux, userSvc *services.UserService, bookSvc *services.BookService, jwtSecret string) {
	userHandler := &UserHandler{service: userSvc}
	bookHandler := &BookHandler{service: bookSvc}

	router.HandleFunc("/user/login", userHandler.login)
	router.Handle("/user/me", withAuth(http.HandlerFunc(userHandler.profile), jwtSecret))
	router.HandleFunc("/user/signup", userHandler.signup)

	router.Handle("/books", withAuth(http.HandlerFunc(bookHandler.myBooks), jwtSecret))
	router.Handle("/books/add", withAuth(http.HandlerFunc(bookHandler.addBook), jwtSecret))
	router.Handle("/books/nearby", withAuth(http.HandlerFunc(bookHandler.nearbyBooks), jwtSecret))
}

func (h *UserHandler) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var req loginRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	token, err := h.service.Login(r.Context(), services.LoginInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			writeJSON(w, http.StatusRequestTimeout, map[string]string{"error": "request timed out"})
			return
		}
		if errors.Is(err, services.ErrInvalidInput) {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid login request"})
		} else if errors.Is(err, services.ErrInvalidCredentials) {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "wrong credentials"})
		} else {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		}
		return
	}

	writeJSON(w, http.StatusOK, tokenResponse{Token: token})
}

func (h *UserHandler) signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var req signupRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	token, err := h.service.Signup(r.Context(), services.SignupInput{
		Username:  req.Username,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
	})
	if err != nil {
		if errors.Is(err, services.ErrInvalidInput) {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid signup request"})
		} else if errors.Is(err, services.ErrUserExists) {
			writeJSON(w, http.StatusConflict, errorResponse{Error: "User already exists, recheck"})
		} else {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		}
		return
	}

	writeJSON(w, http.StatusOK, tokenResponse{Token: token})
}

func (h *UserHandler) profile(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, profileResponse{Username: usernameFromContext(r.Context())})
}

func (h *BookHandler) addBook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	var req addBookRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	username := usernameFromContext(r.Context())
	result, err := h.service.AddBook(r.Context(), services.AddBookInput{
		Title:  req.Title,
		Author: req.Author,
	}, username)
	if err != nil {
		if errors.Is(err, services.ErrInvalidInput) {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid book request"})
		} else {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		}
		return
	}
	message := "Book added to your collection"
	status := http.StatusCreated
	if !result.Added {
		message = "You already listed this book"
		status = http.StatusConflict
	}
	writeJSON(w, status, idResponse{Msg: message, ID: result.CopyID})
}

func (h *BookHandler) myBooks(w http.ResponseWriter, r *http.Request) {
	username := usernameFromContext(r.Context())
	books, err := h.service.MyBooks(r.Context(), username)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}
	writeJSON(w, http.StatusOK, myBooksResponse{Books: toCopyResponses(books)})
}

func (h *BookHandler) nearbyBooks(w http.ResponseWriter, r *http.Request) {
	username := usernameFromContext(r.Context())
	radiusParam := r.URL.Query().Get("radius")
	if radiusParam == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "radius query parameter is required"})
		return
	}
	radius, err := strconv.ParseInt(radiusParam, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "radius must be a number"})
		return
	}
	books, err := h.service.NearbyBooks(r.Context(), username, radius)
	if err != nil {
		if errors.Is(err, services.ErrInvalidInput) {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "radius must be between 1 and 1000 kilometers"})
		} else {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		}
		return
	}
	writeJSON(w, http.StatusOK, nearbyBooksResponse{Books: toNearbyBookResponses(books)})
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, destination interface{}) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var extra interface{}
	if err := decoder.Decode(&extra); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	if extra != nil {
		return errors.New("request body must contain one JSON object")
	}
	return nil
}
