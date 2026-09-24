package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

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

	log.Printf("Starting Bookerang on :%s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err)
	}
}
