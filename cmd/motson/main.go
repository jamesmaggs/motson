// Motson: World Cup 2026 fixtures and scores as a calendar feed and a
// web page. One binary serves HTTP and runs the hourly provider sync
// on an in-process ticker (ADR 0004).
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Jazzatola/motson/internal/config"
	"github.com/Jazzatola/motson/internal/footballdata"
	"github.com/Jazzatola/motson/internal/store"
	"github.com/Jazzatola/motson/internal/syncer"
	"github.com/Jazzatola/motson/internal/web"
)

// tickInterval is how often the ticker checks whether the persisted
// next_run_at has passed; the sync cadence itself is
// fixtures.SyncInterval.
const tickInterval = time.Minute

func main() {
	if err := run(); err != nil {
		slog.Error("motson failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	cfg, err := config.Load(os.LookupEnv)
	if err != nil {
		return err
	}

	st, err := store.NewPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer st.Close()

	source := footballdata.New(cfg.FootballDataURL, cfg.FootballDataToken, cfg.Competition)
	s := syncer.New(source, st)

	runDue := func() {
		if err := s.RunDue(ctx, time.Now().UTC()); err != nil {
			slog.Error("sync failed; serving last synced data", "error", err)
		}
	}
	go func() {
		runDue() // boot sync if due
		for range time.Tick(tickInterval) {
			runDue()
		}
	}()

	handler := web.NewHandler(st, cfg.FeedHost, func() time.Time { return time.Now().UTC() })
	slog.Info("motson listening", "port", cfg.Port)
	return http.ListenAndServe(":"+cfg.Port, handler)
}
