package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/aussenseiter-VsRB/JHIC-BE/config"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/infrastructure/database"
	"github.com/joho/godotenv"
)

func main() {
	if len(os.Args) != 3 || os.Args[1] != "rollback" {
		log.Fatalf("usage: go run ./cmd/migrate rollback <version>")
	}
	version, err := strconv.Atoi(os.Args[2])
	if err != nil || version < 1 {
		log.Fatalf("invalid migration version: %q", os.Args[2])
	}
	_ = godotenv.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	p, err := database.Connect(ctx, config.Load().DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()
	if err := database.RollbackMigration(ctx, p, "cmd/server/migrations", version); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("rolled back migration %03d\n", version)
}
