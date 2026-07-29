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
	pipelineDomain "github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/pipeline"
	userDomain "github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/user"
	workspaceDomain "github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/workspace"
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

	userRepo := userDomain.NewRepository(pool)
	userSvc := userDomain.NewService(userRepo)
	userHnd := userDomain.NewHandler(userSvc)

	workspaceRepo := workspaceDomain.NewRepository(pool)
	workspaceSvc := workspaceDomain.NewService(workspaceRepo)
	workspaceHnd := workspaceDomain.NewHandler(workspaceSvc)

	pipelineRepo := pipelineDomain.NewRepository(pool)
	pipelineSvc := pipelineDomain.NewService(pipelineRepo)
	pipelineHnd := pipelineDomain.NewHandler(pipelineSvc)

	router := internal.NewRouter(userHnd, workspaceHnd, pipelineHnd)

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
