package routes

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

var uptime = time.Now()

type HealthResponse struct {
	Version string `json:"version"`
	Uptime  string `json:"uptime"`
}

func (rg *RouteGroup) Index(ctx *fiber.Ctx) error {
	return ctx.JSON(HealthResponse{
		Version: rg.gctx.Config().Version,
		Uptime:  strconv.Itoa(int(uptime.UnixMilli())),
	})
}
