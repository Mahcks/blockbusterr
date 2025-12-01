package routes

import "github.com/mahcks/blockbusterr/internal/global"

type RouteGroup struct {
	gctx global.Context
}

func NewRouteGroup(gctx global.Context) *RouteGroup {
	return &RouteGroup{
		gctx: gctx,
	}
}
