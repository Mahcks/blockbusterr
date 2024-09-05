package global

import (
	"context"
	"time"

	"github.com/mahcks/blockbusterr/internal/services"
)

type Context interface {
	context.Context
	Crate() *services.Crate
}

type gCtx struct {
	context.Context
	crate *services.Crate
}

func (g *gCtx) Crate() *services.Crate {
	return g.crate
}

func New(ctx context.Context) Context {
	return &gCtx{
		Context: ctx,
		crate:   &services.Crate{},
	}
}

func WithCancel(ctx Context) (Context, context.CancelFunc) {
	crate := ctx.Crate()

	c, cancel := context.WithCancel(ctx)

	return &gCtx{
		Context: c,
		crate:   crate,
	}, cancel
}

func WithDeadline(ctx Context, deadline time.Time) (Context, context.CancelFunc) {
	crate := ctx.Crate()

	c, cancel := context.WithDeadline(ctx, deadline)

	return &gCtx{
		Context: c,
		crate:   crate,
	}, cancel
}

func WithValue(ctx Context, key interface{}, value interface{}) Context {
	crate := ctx.Crate()

	return &gCtx{
		Context: context.WithValue(ctx, key, value),
		crate:   crate,
	}
}

func WithTimeout(ctx Context, timeout time.Duration) (Context, context.CancelFunc) {
	crate := ctx.Crate()

	c, cancel := context.WithTimeout(ctx, timeout)

	return &gCtx{
		Context: c,
		crate:   crate,
	}, cancel
}
