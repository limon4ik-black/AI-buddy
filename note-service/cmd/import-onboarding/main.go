package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/limon4ik-black/AI-buddy/note-service/onboarding"
	"log"
	"os"
)

func main() {
	ctx := context.Background()
	src, err := pgxpool.New(ctx, os.Getenv("LEGACY_DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer src.Close()
	dst, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer dst.Close()
	if err = onboarding.ImportLegacy(ctx, src, dst); err != nil {
		log.Fatal(err)
	}
	log.Print("Onboarding snapshot imported; existing Note data preserved")
}
