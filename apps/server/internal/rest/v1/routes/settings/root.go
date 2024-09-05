package settings

import (
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/internal/helpers"
)

type SettingPayload struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  string `json:"type"` // Optional field, can default to "text"
}

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
