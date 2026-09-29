package main

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/sakid00/enmasse-be/internal/config"
	"github.com/sakid00/enmasse-be/internal/domain"
	"github.com/sakid00/enmasse-be/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}
	path := "testdata/artist_import.csv"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	if err := run(context.Background(), cfg.DatabaseURL, path); err != nil {
		slog.Error("seed", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, dsn, path string) error {
	db, err := store.Connect(ctx, dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("open csv: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	header, err := r.Read()
	if err != nil {
		return fmt.Errorf("csv header: %w", err)
	}
	idx := index(header)

	inserted := 0
	skipped := 0
	for {
		row, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("csv row: %w", err)
		}
		email := domain.CanonicalEmail(col(row, idx, "email"))
		if !domain.LooksLikeEmail(email) {
			skipped++
			continue
		}
		if _, err := db.Queries.GetUserByEmail(ctx, email); err == nil {
			skipped++
			continue
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		handle := domain.CanonicalHandle(col(row, idx, "handle"))
		if !domain.ValidHandle(handle) {
			handle = uniqueHandle(ctx, db, domain.HandleFromEmail(email))
		} else if _, err := db.Queries.GetUserByHandle(ctx, handle); err == nil {
			handle = uniqueHandle(ctx, db, handle)
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		tx, err := db.Pool.Begin(ctx)
		if err != nil {
			return err
		}
		q := db.Queries.WithTx(tx)
		user, err := q.CreateUser(ctx, store.CreateUserParams{
			Email:              email,
			PasswordHash:       nil,
			Role:               "artist",
			Handle:             handle,
			MustChangePassword: false,
		})
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("user %s: %w", email, err)
		}
		_, err = q.CreateArtist(ctx, store.CreateArtistParams{
			UserID:               user.ID,
			FirstName:            col(row, idx, "first_name"),
			LastName:             col(row, idx, "last_name"),
			ArtistName:           col(row, idx, "artist_name"),
			PhotoUrl:             strPtr(col(row, idx, "profile_photo_url")),
			Address:              nil,
			WebsiteUrl:           strPtr(col(row, idx, "website_url")),
			InstagramHandle:      strPtr(col(row, idx, "instagram_handle")),
			TwitterHandle:        strPtr(col(row, idx, "twitter_handle")),
			Whatsapp:             strPtr(col(row, idx, "whatsapp")),
			ImportedArtistID:     parseUUID(col(row, idx, "id")),
			ImportedPortalUserID: parseUUID(col(row, idx, "user_id")),
			Bio:                  strPtr(col(row, idx, "bio")),
			Nationality:          strPtr(col(row, idx, "nationality")),
		})
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("artist %s: %w", email, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
		inserted++
	}
	slog.Info("seed complete", "inserted", inserted, "skipped", skipped)
	return nil
}

func uniqueHandle(ctx context.Context, db *store.DB, base string) string {
	if !domain.ValidHandle(base) {
		base = "artist"
	}
	for i := 0; i < 1000; i++ {
		h := base
		if i > 0 {
			suffix := fmt.Sprintf("%d", i)
			if len(base)+len(suffix) > 32 {
				h = base[:32-len(suffix)] + suffix
			} else {
				h = base + suffix
			}
		}
		if _, err := db.Queries.GetUserByHandle(ctx, h); errors.Is(err, pgx.ErrNoRows) {
			return h
		}
	}
	return fmt.Sprintf("a%s", strings.ReplaceAll(uuid.NewString(), "-", "")[:16])
}

func index(header []string) map[string]int {
	m := make(map[string]int, len(header))
	for i, h := range header {
		m[strings.TrimSpace(h)] = i
	}
	return m
}

func col(row []string, idx map[string]int, name string) string {
	i, ok := idx[name]
	if !ok || i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func parseUUID(s string) *uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return nil
	}
	return &id
}
