package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/aussenseiter-VsRB/JHIC-BE/config"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/auth"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/user"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/infrastructure/database"
	"github.com/joho/godotenv"
)

func main() {
	exe, _ := os.Executable()
	_ = godotenv.Load(filepath.Join(filepath.Dir(exe), ".env"))
	_ = godotenv.Load()
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	if err := database.RunMigrations(ctx, pool, "cmd/server/migrations"); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	usersRepo := auth.NewUsersRepository(pool)
	sessionsRepo := auth.NewSessionsRepository(pool)
	authSvc := auth.NewService(usersRepo, sessionsRepo)
	authHnd := auth.NewHandler(authSvc)

	userRepo := user.NewRepository(pool)
	userSvc := user.NewService(userRepo)
	userHnd := user.NewHandler(userSvc)

	router := internal.NewRouter(authHnd, userHnd)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("server listening on :%d", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdown); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
}
