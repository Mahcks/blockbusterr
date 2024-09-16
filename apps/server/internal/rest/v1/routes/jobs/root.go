package jobs

import (
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/internal/helpers"
	"github.com/mahcks/blockbusterr/internal/scheduler"
)

type RouteGroup struct {
	gctx      global.Context
	helpers   *helpers.Helpers
	scheduler *scheduler.Scheduler
}

func NewRouteGroup(gctx global.Context, helpers *helpers.Helpers, scheduler *scheduler.Scheduler) *RouteGroup {
	return &RouteGroup{
		gctx:      gctx,
		helpers:   helpers,
		scheduler: scheduler,
	}
}
