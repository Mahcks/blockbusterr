package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/internal/helpers"
	"github.com/mahcks/blockbusterr/internal/rest/v1/respond"
	"github.com/mahcks/blockbusterr/internal/rest/v1/routes"
	"github.com/mahcks/blockbusterr/internal/rest/v1/routes/settings"
)

func ctx(fn func(*respond.Ctx) error) fiber.Handler {
	return func(c *fiber.Ctx) error {
		newCtx := &respond.Ctx{Ctx: c}
		return fn(newCtx)
	}
}

func New(gctx global.Context, helpers *helpers.Helpers, router fiber.Router) {
	indexRoute := routes.NewRouteGroup(gctx, helpers)
	router.Get("/", indexRoute.Index)

	settings := settings.NewRouteGroup(gctx, helpers)
	router.Post("/settings", ctx(settings.PostSetting))
	router.Get("/settings", ctx(settings.GetSetting))
	router.Delete("/settings", ctx(settings.DeleteSetting))
	router.Put("/settings", ctx(settings.PutSetting))
}
