package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/internal/rest/v1/routes"
)

func New(gctx global.Context, router fiber.Router) {
	rg := routes.NewRouteGroup(gctx)

	// Health check route
	router.Get("/", rg.Index)

	// Register integration API routes
	routes.RegisterTraktRoutes(rg, router)
	routes.RegisterRadarrRoutes(rg, router)
	routes.RegisterSonarrRoutes(rg, router)
	rg.RegisterJellyseerrRoutes(router)

	// Register jobs API routes
	routes.AddJobsRoutes(router, gctx)

	// Register dynamic jobs API routes
	routes.AddDynamicJobsRoutes(router, gctx)

	// Register activity API routes
	routes.RegisterActivityRoutes(router, gctx)
}
