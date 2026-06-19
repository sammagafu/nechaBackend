package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/nechaafrica/backend/internal/config"
	"github.com/nechaafrica/backend/internal/database"
)

func main() {
	email := flag.String("email", "", "Admin email (or set SUPERUSER_EMAIL / ADMIN_EMAIL)")
	password := flag.String("password", "", "Admin password (or set SUPERUSER_PASSWORD / ADMIN_PASSWORD)")
	name := flag.String("name", "", "Display name (default: Necha Admin)")
	force := flag.Bool("force", false, "Update password and promote to admin if the email already exists")
	flag.Parse()

	resolvedEmail := firstNonEmpty(
		*email,
		os.Getenv("SUPERUSER_EMAIL"),
		os.Getenv("ADMIN_EMAIL"),
	)
	resolvedPassword := firstNonEmpty(
		*password,
		os.Getenv("SUPERUSER_PASSWORD"),
		os.Getenv("ADMIN_PASSWORD"),
	)
	resolvedName := firstNonEmpty(*name, os.Getenv("SUPERUSER_NAME"), "Necha Admin")

	if resolvedEmail == "" {
		log.Fatal("email is required: pass --email or set SUPERUSER_EMAIL")
	}
	if resolvedPassword == "" {
		log.Fatal("password is required: pass --password or set SUPERUSER_PASSWORD")
	}

	cfg := config.Load()
	db, err := database.Connect(cfg.Database.DSN())
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}

	created, err := database.UpsertSuperUser(db, resolvedEmail, resolvedPassword, resolvedName, *force)
	if err != nil {
		log.Fatalf("failed to upsert super user: %v", err)
	}

	if created {
		fmt.Printf("Created admin user %s\n", strings.ToLower(resolvedEmail))
		return
	}
	if *force {
		fmt.Printf("Updated admin user %s\n", strings.ToLower(resolvedEmail))
		return
	}
	fmt.Printf("Admin user %s already exists (use --force to update password and role)\n", strings.ToLower(resolvedEmail))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
