package store_test

import (
	"context"
	"os"
	"testing"

	"github.com/Jazzatola/motson/internal/fixtures"
	"github.com/Jazzatola/motson/internal/store"
	"github.com/Jazzatola/motson/internal/store/storetest"
)

// The same contract suite as the memory fake, against real Postgres
// (ADR 0007). Runs whenever TEST_DATABASE_URL is set; CI always sets it.
func TestPostgresStoreContract(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping Postgres contract suite")
	}

	ctx := context.Background()
	pg, err := store.NewPostgres(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pg.Close)

	storetest.Run(t, func(t *testing.T) fixtures.Store {
		if err := pg.Reset(ctx); err != nil {
			t.Fatal(err)
		}
		return pg
	})
}
