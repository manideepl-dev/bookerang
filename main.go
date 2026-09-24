package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"

	"bookerang/internal/config"
	"bookerang/internal/httpapi"
	"bookerang/internal/repository"
	"bookerang/internal/services"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	dbContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(dbContext); err != nil {
		log.Fatalf("connect to database: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	bookRepo := repository.NewBookRepository(db)

	userSvc := services.NewUserService(userRepo, cfg.JWTSecret)
	bookSvc := services.NewBookService(bookRepo)

	router := http.NewServeMux()
	httpapi.RegisterRoutes(router, userSvc, bookSvc, cfg.JWTSecret)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           httpapi.WithCORS(router, cfg.CORSOrigin),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("Starting Bookerang on :%s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
