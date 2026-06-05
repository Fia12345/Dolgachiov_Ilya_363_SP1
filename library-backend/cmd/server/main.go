package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"library-backend/internal/config"
	"library-backend/internal/handlers"
	"library-backend/internal/middleware"
	"library-backend/internal/repository"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	repo, err := repository.NewSQLiteRepository(cfg.DBPath)
	if err != nil {
		slog.Error("Не удалось инициализировать базу данных", "error", err)
		os.Exit(1)
	}
	defer repo.Close()
	h := handlers.NewHandlers(repo, cfg.JWT.Secret)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /books", h.GetBooks)
	mux.HandleFunc("GET /books/{id}", h.GetBook)
	protected := middleware.AuthMiddleware(h, cfg.JWT.Secret)

	mux.Handle("POST /books", protected(http.HandlerFunc(h.CreateBook)))
	mux.Handle("PUT /books/{id}", protected(http.HandlerFunc(h.UpdateBook)))
	mux.Handle("POST /users", protected(http.HandlerFunc(h.RegisterUser)))
	mux.Handle("GET /users/{id}/books", protected(http.HandlerFunc(h.GetUserBooks)))
	mux.Handle("POST /issues", protected(http.HandlerFunc(h.IssueBook)))
	mux.Handle("POST /returns", protected(http.HandlerFunc(h.ReturnBook)))
	mux.HandleFunc("POST /login", h.Login)

	slog.Info("Сервер запускается", "port", cfg.ServerPort)

	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: mux,
	}
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Ошибка запуска сервера", "error", err)
		}
	}()

	<-done
	slog.Info("Ошибка завершения работы сервера...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Shutdown failed", "error", err)
	}
}
