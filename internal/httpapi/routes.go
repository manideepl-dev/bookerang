package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	jwtutil "bookerang/internal/jwt"
	"bookerang/internal/models"
	"bookerang/internal/services"
)

type UserHandler struct {
	service   *services.UserService
	jwtSecret string
}

type BookHandler struct {
	service *services.BookService
}

func RegisterRoutes(router *http.ServeMux, userSvc *services.UserService, bookSvc *services.BookService, jwtSecret string) {
	userHandler := &UserHandler{service: userSvc, jwtSecret: jwtSecret}
	bookHandler := &BookHandler{service: bookSvc}

	router.HandleFunc("/user/login", userHandler.login)
	router.HandleFunc("/user/me", userHandler.profile)
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

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	token, err := h.service.Login(context.Background(), req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			writeJSON(w, http.StatusRequestTimeout, map[string]string{"error": "request timed out"})
			return
		}
		switch {
		case strings.Contains(err.Error(), "user not found"):
			writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "Wrong credentials, recheck"})
		case strings.Contains(err.Error(), "invalid credentials"):
			writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "Wrong credentials, recheck"})
		default:
			writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		}
		return
	}

	writeJSON(w, http.StatusOK, models.TokenResponse{Token: token})
}

func (h *UserHandler) signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var req models.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	token, err := h.service.Signup(context.Background(), req)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "user already exists"):
			writeJSON(w, http.StatusConflict, models.ErrorResponse{Error: "User already exists, recheck"})
		default:
			writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		}
		return
	}

	writeJSON(w, http.StatusOK, models.TokenResponse{Token: token})
}

func (h *UserHandler) profile(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if _, err := jwtutil.Validate(token, h.jwtSecret); err == nil {
			writeJSON(w, http.StatusOK, models.IdResponse{Msg: "casD", ID: "dqd"})
			return
		}
	}
	writeJSON(w, http.StatusOK, models.IdResponse{Msg: "casD", ID: "dqd"})
}

func (h *BookHandler) addBook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	var req models.AddBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	username := usernameFromContext(r.Context())
	result, err := h.service.AddBook(r.Context(), req, username)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}
	message := "Book added to your collection"
	status := http.StatusCreated
	if !result.Added {
		message = "You already listed this book"
		status = http.StatusConflict
	}
	writeJSON(w, status, models.IdResponse{Msg: message, ID: result.CopyID})
}

func (h *BookHandler) myBooks(w http.ResponseWriter, r *http.Request) {
	username := usernameFromContext(r.Context())
	books, err := h.service.MyBooks(r.Context(), username)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, models.MyBooksResponse{Books: books})
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
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, models.NearbyBooksResponse{Books: books})
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
