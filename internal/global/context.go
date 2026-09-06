package global

import (
	"context"
	"sync"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
	"github.com/mahcks/blockbusterr/internal/services/jobs"
)

type Metadata struct {
	Version string
	Commit  string
}

type Context interface {
	context.Context
	Metadata() Metadata
	Config() *config.Config
	Database() *database.Database
	ExecutionCoordinator() *jobs.ExecutionCoordinator
	ReloadConfig() error
}

type gCtx struct {
	context.Context
	metadata   Metadata
	cfg        *config.Config
	cfgMu      sync.RWMutex
	db         *database.Database
	executions *jobs.ExecutionCoordinator
}

func (g *gCtx) Metadata() Metadata {
	return g.metadata
}

func (g *gCtx) Config() *config.Config {
	g.cfgMu.RLock()
	defer g.cfgMu.RUnlock()
	clone, err := g.cfg.Clone()
	if err != nil {
		panic(err)
	}
	return clone
}

func (g *gCtx) Database() *database.Database {
	return g.db
}

func (g *gCtx) ExecutionCoordinator() *jobs.ExecutionCoordinator { return g.executions }

func (g *gCtx) ReloadConfig() error {
	g.cfgMu.Lock()
	defer g.cfgMu.Unlock()

	// Reload the config from the file
	newCfg, err := config.New(g.metadata.Version)
	if err != nil {
		return err
	}

	g.cfg = newCfg
	g.bindConfigCallbacks(newCfg)
	return nil
}

// UpdateConfig serializes copy-on-write mutations and publishes only saved values.
func (g *gCtx) UpdateConfig(update func(*config.Config) error) error {
	g.cfgMu.Lock()
	defer g.cfgMu.Unlock()

	candidate, err := g.cfg.Clone()
	if err != nil {
		return err
	}
	if err := update(candidate); err != nil {
		return err
	}
	if err := candidate.Save(); err != nil {
		return err
	}
	g.bindConfigCallbacks(candidate)
	g.cfg = candidate
	return nil
}

func (g *gCtx) RestoreConfig(candidate *config.Config) error {
	g.cfgMu.Lock()
	defer g.cfgMu.Unlock()

	if candidate.ConfigFilePath == "" {
		candidate.ConfigFilePath = g.cfg.ConfigFilePath
	}
	if err := candidate.Restore(); err != nil {
		return err
	}
	g.bindConfigCallbacks(candidate)
	g.cfg = candidate
	return nil
}

// UpdateConfig uses the production context's atomic updater and keeps test contexts compatible.
func UpdateConfig(ctx Context, update func(*config.Config) error) error {
	if updater, ok := ctx.(interface {
		UpdateConfig(func(*config.Config) error) error
	}); ok {
		return updater.UpdateConfig(update)
	}
	cfg := ctx.Config()
	candidate, err := cfg.Clone()
	if err != nil {
		return err
	}
	if err := update(candidate); err != nil {
		return err
	}
	if err := candidate.Save(); err != nil {
		return err
	}
	*cfg = *candidate
	return ctx.ReloadConfig()
}

func RestoreConfig(ctx Context, candidate *config.Config) error {
	if restorer, ok := ctx.(interface {
		RestoreConfig(*config.Config) error
	}); ok {
		return restorer.RestoreConfig(candidate)
	}
	current := ctx.Config()
	if candidate.ConfigFilePath == "" {
		candidate.ConfigFilePath = current.ConfigFilePath
	}
	if err := candidate.Restore(); err != nil {
		return err
	}
	*current = *candidate
	return nil
}

func New(
	ctx context.Context,
	cfg *config.Config,
	db *database.Database,
	Version, Commit string,
	executions *jobs.ExecutionCoordinator,
) Context {
	g := &gCtx{
		cfg:     cfg,
		Context: ctx,
		metadata: Metadata{
			Version: Version,
			Commit:  Commit,
		},
		db: db,
	}
	g.executions = executions
	g.bindConfigCallbacks(cfg)
	return g
}

func (g *gCtx) bindConfigCallbacks(cfg *config.Config) {
	cfg.LoadTraktToken = func() (string, string, int64) {
		g.cfgMu.RLock()
		defer g.cfgMu.RUnlock()
		return g.cfg.Trakt.AccessToken, g.cfg.Trakt.RefreshToken, g.cfg.Trakt.TokenExpires
	}

	cfg.UpdateTraktToken = func(access, refresh string, expires int64) error {
		return g.UpdateConfig(func(candidate *config.Config) error {
			candidate.Trakt.AccessToken = access
			candidate.Trakt.RefreshToken = refresh
			candidate.Trakt.TokenExpires = expires
			return nil
		})
	}
}
