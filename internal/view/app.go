package view

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/slok/grafterm/internal/service/log"
	viewsync "github.com/slok/grafterm/internal/view/sync"
	"github.com/slok/grafterm/internal/view/template"
)

// AppConfig are the options to run the app.
// this configuration  has values at global app level.
type AppConfig struct {
	RefreshInterval   time.Duration
	TimeRangeStart    time.Time // Fixed optional time.
	TimeRangeEnd      time.Time // Fixed optional time.
	RelativeTimeRange time.Duration
}

func (a *AppConfig) defaults() {
	const (
		defRelativeTimeRange = 1 * time.Hour
		defRefreshInterval   = 10 * time.Second
	)

	if a.RefreshInterval == 0 {
		a.RefreshInterval = defRefreshInterval
	}
	if a.RelativeTimeRange == 0 {
		a.RelativeTimeRange = defRelativeTimeRange
	}
}

// App represents the application that will render the metrics dashboard.
type App struct {
	syncer viewsync.Syncer
	cfg    AppConfig
	logger log.Logger

	running bool
	mu      sync.Mutex
}

// NewApp Is the main application
func NewApp(cfg AppConfig, syncer viewsync.Syncer, logger log.Logger) *App {
	cfg.defaults()

	return &App{
		cfg:    cfg,
		syncer: syncer,
		logger: logger,
	}
}

// Run will start running the application.
func (a *App) Run(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.running {
		return errors.New("already running")
	}
	a.running = true

	// TODO(slok): Think if we should set running to false, for now we
	// don't want to reuse the app.
	return a.run(ctx)
}

func (a *App) run(ctx context.Context) error {
	// Start the sync loop. This operation blocks.
	a.sync()

	tk := time.NewTicker(a.cfg.RefreshInterval)
	defer tk.Stop()
	for {
		// Check if we already done.
		select {
		case <-ctx.Done():
			return nil
		case <-tk.C:
		}

		a.sync()
	}
}

func (a *App) sync() {
	defer func() {
		if r := recover(); r != nil {
			a.logger.Errorf("app sync panic recovered: %v", r)
		}
	}()

	// Create context with timeout for this sync operation
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	if a.syncer == nil {
		a.logger.Errorf("syncer is nil, cannot sync")
		return
	}

	r := a.syncRequest()
	if r == nil {
		a.logger.Errorf("sync request is nil, cannot sync")
		return
	}

	err := a.syncer.Sync(ctx, r)
	if err != nil {
		switch ctx.Err() {
		case context.DeadlineExceeded:
			a.logger.Errorf("app sync timeout after 8s: %s", err)
		case context.Canceled:
			a.logger.Errorf("app sync canceled: %s", err)
		default:
			a.logger.Errorf("app level error, syncer failed sync: %s", err)
		}
	}
}

func (a *App) syncRequest() *viewsync.Request {
	r := &viewsync.Request{
		TimeRangeStart: a.cfg.TimeRangeStart,
		TimeRangeEnd:   a.cfg.TimeRangeEnd,
	}

	// If we don't have fixed time, make the time ranges work in relative mode
	// based on now timestamp.
	if r.TimeRangeEnd.IsZero() {
		r.TimeRangeEnd = time.Now().UTC()
	}
	if r.TimeRangeStart.IsZero() {
		r.TimeRangeStart = r.TimeRangeEnd.Add(-1 * a.cfg.RelativeTimeRange)
	}

	// Create the template data for each sync.
	r.TemplateData = a.syncData(r)

	return r
}

func (a *App) syncData(r *viewsync.Request) template.Data {
	data := map[string]interface{}{
		"__start": fmt.Sprintf("%v", r.TimeRangeStart),
		"__end":   fmt.Sprintf("%v", r.TimeRangeEnd),
	}
	return data
}
