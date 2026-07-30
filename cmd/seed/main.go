package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/infrastructure/database"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/pkg/id"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

type seedUser struct {
	email    string
	password string
	name     string
	role     string
}

type seedBerita struct {
	authorIdx int
	title     string
	content   string
}

func main() {
	exe, _ := os.Executable()
	_ = godotenv.Load(filepath.Join(filepath.Dir(exe), ".env"))
	_ = godotenv.Load()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := database.Connect(ctx, dsn)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	users := []seedUser{
		{email: "admin@jhic.com", password: "admin123", name: "Admin User", role: "admin"},
		{email: "jurnal@jhic.com", password: "jurnal123", name: "Jurnalis User", role: "jurnal"},
		{email: "user@jhic.com", password: "user123", name: "Regular User", role: "user"},
	}

	var userIDs []string
	for _, u := range users {
		uid := id.New()
		hash, err := bcrypt.GenerateFromPassword([]byte(u.password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("hash password for %s: %v", u.email, err)
		}
		now := time.Now().UTC()
		_, err = pool.Exec(ctx,
			`INSERT INTO users (id, email, password_hash, name, role, avatar_url, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, '', $6, $7)
			 ON CONFLICT (email) DO UPDATE SET name = $4, role = $5, password_hash = $3, updated_at = $7`,
			uid, u.email, string(hash), u.name, u.role, now, now,
		)
		if err != nil {
			log.Fatalf("insert user %s: %v", u.email, err)
		}
		userIDs = append(userIDs, uid)
		fmt.Printf("user: %s / %s (%s)\n", u.email, u.password, uid)
	}

	for _, uid := range userIDs {
		token := id.New()
		_, err := pool.Exec(ctx,
			`INSERT INTO sessions (token, user_id, created_at, expires_at)
			 VALUES ($1, $2, $3, $4)`,
			token, uid, time.Now().UTC(), time.Now().Add(72*time.Hour).UTC(),
		)
		if err != nil {
			log.Fatalf("insert session for %s: %v", uid, err)
		}
		fmt.Printf("session token for %s: %s\n", uid, token)
	}

	beritas := []seedBerita{
		{authorIdx: 0, title: "Selamat Datang di JHIC", content: "Ini adalah berita pertama di portal JHIC."},
		{authorIdx: 0, title: "Pembukaan Pendaftaran Anggota Baru", content: "Pendaftaran anggota baru JHIC telah dibuka."},
		{authorIdx: 1, title: "Liputan Kegiatan Jurnalistik", content: "Kegiatan jurnalistik hari ini berjalan dengan lancar."},
		{authorIdx: 1, title: "Wawancara Eksklusif dengan Tokoh Masyarakat", content: "Berikut adalah wawancara eksklusif kami."},
		{authorIdx: 0, title: "Pengumuman Libur Nasional", content: "JHIC libur pada hari libur nasional."},
	}

	for _, b := range beritas {
		bid := id.New()
		now := time.Now().UTC()
		_, err := pool.Exec(ctx,
			`INSERT INTO berita (id, author_id, title, content, image_url, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, '', $5, $6)
			 ON CONFLICT (id) DO NOTHING`,
			bid, userIDs[b.authorIdx], b.title, b.content, now, now,
		)
		if err != nil {
			log.Fatalf("insert berita: %v", err)
		}
		fmt.Printf("berita: %s by %s\n", b.title, users[b.authorIdx].email)
	}

	fmt.Println("\nseed complete")
}

func init() {
	log.SetFlags(0)
}
