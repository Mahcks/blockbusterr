package global

import (
	"context"
	"time"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/services"
)

type Context interface {
	context.Context
	Config() *config.Config
	Crate() *services.Crate
}

type gCtx struct {
	context.Context
	cfg   *config.Config
	crate *services.Crate
}

func (g *gCtx) Config() *config.Config {
	return g.cfg
}

func (g *gCtx) Crate() *services.Crate {
	return g.crate
}

func New(ctx context.Context, cfg *config.Config) Context {
	return &gCtx{
		Context: ctx,
		cfg:     cfg,
		crate:   &services.Crate{},
	}
}

func WithCancel(ctx Context) (Context, context.CancelFunc) {
	cfg := ctx.Config()
	crate := ctx.Crate()

	c, cancel := context.WithCancel(ctx)

	return &gCtx{
		Context: c,
		cfg:     cfg,
		crate:   crate,
	}, cancel
}

func WithDeadline(ctx Context, deadline time.Time) (Context, context.CancelFunc) {
	cfg := ctx.Config()
	crate := ctx.Crate()

	c, cancel := context.WithDeadline(ctx, deadline)

	return &gCtx{
		Context: c,
		cfg:     cfg,
		crate:   crate,
	}, cancel
}

func WithValue(ctx Context, key interface{}, value interface{}) Context {
	cfg := ctx.Config()
	crate := ctx.Crate()

	return &gCtx{
		Context: context.WithValue(ctx, key, value),
		cfg:     cfg,
		crate:   crate,
	}
}

func WithTimeout(ctx Context, timeout time.Duration) (Context, context.CancelFunc) {
	cfg := ctx.Config()
	crate := ctx.Crate()

	c, cancel := context.WithTimeout(ctx, timeout)

	return &gCtx{
		Context: c,
		cfg:     cfg,
		crate:   crate,
	}, cancel
}
