package global

import (
	"context"
	"sync"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
)

type Metadata struct {
	Version   string
	Timestamp string
}

type Context interface {
	context.Context
	Metadata() Metadata
	Config() *config.Config
	Database() *database.Database
	ReloadConfig() error
}

type gCtx struct {
	context.Context
	metadata Metadata
	cfg      *config.Config
	cfgMu    sync.RWMutex
	db       *database.Database
}

func (g *gCtx) Metadata() Metadata {
	return g.metadata
}

func (g *gCtx) Config() *config.Config {
	g.cfgMu.RLock()
	defer g.cfgMu.RUnlock()
	return g.cfg
}

func (g *gCtx) Database() *database.Database {
	return g.db
}

func (g *gCtx) ReloadConfig() error {
	g.cfgMu.Lock()
	defer g.cfgMu.Unlock()

	// Reload the config from the file
	newCfg, err := config.New(g.metadata.Version)
	if err != nil {
		return err
	}

	g.cfg = newCfg
	return nil
}

func New(
	ctx context.Context,
	cfg *config.Config,
	db *database.Database,
	Version string,
	Timestamp string,
) Context {
	return &gCtx{
		cfg:     cfg,
		Context: ctx,
		metadata: Metadata{
			Version:   Version,
			Timestamp: Timestamp,
		},
		db: db,
	}
}

func WithCancel(ctx Context) (Context, context.CancelFunc) {
	metadata := ctx.Metadata()
	cfg := ctx.Config()
	db := ctx.Database()

	c, cancel := context.WithCancel(ctx)

	return &gCtx{
		Context:  c,
		cfg:      cfg,
		metadata: metadata,
		db:       db,
	}, cancel
}
