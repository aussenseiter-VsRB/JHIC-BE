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
	authpg "github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/auth/pg"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/berita"
	beritapg "github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/berita/pg"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/user"
	userpg "github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/user/pg"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/infrastructure/database"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/infrastructure/middleware"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/infrastructure/storage"
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

	usersRepo := authpg.NewUsersRepository(pool)
	sessionsRepo := authpg.NewSessionsRepository(pool)
	authSvc := auth.NewService(usersRepo, sessionsRepo)
	authHnd := auth.NewHandler(authSvc)

	userRepo := userpg.NewRepository(pool)
	userSvc := user.NewService(userRepo)
	userHnd := user.NewHandler(userSvc)

	b2Cfg := storage.B2Config{
		Endpoint: cfg.B2Endpoint,
		Region:   cfg.B2Region,
		KeyID:    cfg.B2KeyID,
		AppKey:   cfg.B2AppKey,
		Bucket:   cfg.B2Bucket,
	}
	b2Client, err := storage.NewB2Client(ctx, b2Cfg)
	if err != nil {
		log.Fatalf("b2 storage: %v", err)
	}

	beritaRepo := beritapg.NewRepository(pool)
	beritaSvc := berita.NewService(beritaRepo)
	beritaHnd := berita.NewHandler(beritaSvc, b2Client)

	tokenValidator := middleware.TokenValidator(auth.NewTokenValidator(sessionsRepo))
	authMw := middleware.Auth(tokenValidator)
	roleCheck := userSvc.ByID
	roleChecker := func(ctx context.Context, userID string) (string, error) {
		u, err := roleCheck(ctx, userID)
		if err != nil {
			return "", err
		}
		if u == nil {
			return "", fmt.Errorf("user not found")
		}
		return u.Role, nil
	}
	roleMw := middleware.RequireRole("jurnal")(roleChecker)

	router := internal.NewRouter(authHnd, userHnd, beritaHnd, authMw, roleMw)

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
