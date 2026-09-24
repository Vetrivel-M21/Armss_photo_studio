// Command seed creates the first admin user. There is no public registration
// endpoint (single-tenant, spec §2), so this is how the very first account gets made.
package main

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"

	"armss-poster-studio/internal/auth"
	"armss-poster-studio/internal/config"
	"armss-poster-studio/internal/store/mysql"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load(os.LookupEnv)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	email := os.Getenv("SEED_ADMIN_EMAIL")
	password := os.Getenv("SEED_ADMIN_PASSWORD")
	if email == "" || password == "" {
		log.Fatal("SEED_ADMIN_EMAIL and SEED_ADMIN_PASSWORD must be set")
	}

	if err := mysql.Migrate(cfg.MySQLDSN); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	db, err := mysql.Open(cfg.MySQLDSN)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer db.Close()

	usersRepo := mysql.NewUsersRepo(db)

	if _, err := usersRepo.GetByEmail(context.Background(), email); err == nil {
		log.Printf("admin %s already exists, nothing to do", email)
		return
	} else if !errors.Is(err, mysql.ErrNotFound) {
		log.Fatalf("lookup existing admin: %v", err)
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		log.Fatalf("hash password: %v", err)
	}

	if _, err := usersRepo.Create(context.Background(), email, hash, mysql.RoleAdmin); err != nil {
		log.Fatalf("create admin: %v", err)
	}

	log.Printf("created admin user %s", email)
}
