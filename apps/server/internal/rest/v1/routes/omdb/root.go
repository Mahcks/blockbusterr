package omdb

import (
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/internal/helpers"
)

type RouteGroup struct {
	gctx    global.Context
	helpers *helpers.Helpers
}

func NewRouteGroup(gctx global.Context, helpers *helpers.Helpers) *RouteGroup {
	return &RouteGroup{
		gctx:    gctx,
		helpers: helpers,
	}
}
