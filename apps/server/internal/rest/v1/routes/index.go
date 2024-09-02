package routes

import (
	"strconv"
	"time"

	fiber "github.com/gofiber/fiber/v2"
)

var uptime = time.Now()

type HealthResponse struct {
	Online bool   `json:"online"`
	Uptime string `json:"uptime"`
}

func (rg *RouteGroup) Index(ctx *fiber.Ctx) error {
	return ctx.JSON(HealthResponse{
		Online: true,
		Uptime: strconv.Itoa(int(uptime.UnixMilli())),
	})
}
