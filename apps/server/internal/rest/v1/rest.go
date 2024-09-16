package v1

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/internal/helpers"
	"github.com/mahcks/blockbusterr/internal/rest/v1/respond"
	"github.com/mahcks/blockbusterr/internal/rest/v1/routes"
	"github.com/mahcks/blockbusterr/internal/rest/v1/routes/jobs"
	"github.com/mahcks/blockbusterr/internal/rest/v1/routes/movies"
	"github.com/mahcks/blockbusterr/internal/rest/v1/routes/radarr"
	"github.com/mahcks/blockbusterr/internal/rest/v1/routes/settings"
	"github.com/mahcks/blockbusterr/internal/rest/v1/routes/shows"
	"github.com/mahcks/blockbusterr/internal/rest/v1/routes/sonarr"
	"github.com/mahcks/blockbusterr/internal/rest/v1/routes/trakt"
	"github.com/mahcks/blockbusterr/internal/scheduler"
	ws "github.com/mahcks/blockbusterr/internal/websocket"
)

func ctx(fn func(*respond.Ctx) error) fiber.Handler {
	return func(c *fiber.Ctx) error {
		newCtx := &respond.Ctx{Ctx: c}
		return fn(newCtx)
	}
}

func New(gctx global.Context, hub *ws.Hub, helpers *helpers.Helpers, scheduler *scheduler.Scheduler, router fiber.Router) {
	indexRoute := routes.NewRouteGroup(gctx, helpers)
	router.Get("/", indexRoute.Index)

	router.Get("/ws", websocket.New(func(c *websocket.Conn) {
		hub.ServeWs(c)
	}))

	settings := settings.NewRouteGroup(gctx, helpers)
	router.Post("/settings", ctx(settings.PostSetting))
	router.Get("/settings", ctx(settings.GetSetting))
	router.Delete("/settings", ctx(settings.DeleteSetting))
	router.Put("/settings", ctx(settings.PutSetting))

	movies := movies.NewRouteGroup(gctx, helpers)
	router.Get("/movie/settings", ctx(movies.GetMovieSettings))

	shows := shows.NewRouteGroup(gctx, helpers)
	router.Get("/show/settings", ctx(shows.GetShowSettings))

	radarr := radarr.NewRouteGroup(gctx, helpers)
	router.Get("/radarr/settings", ctx(radarr.GetRadarrSettings))
	router.Get("/radarr/profiles", ctx(radarr.GetRadarrProfiles))
	router.Get("/radarr/rootfolders", ctx(radarr.GetRadarrRootFolders))

	sonarr := sonarr.NewRouteGroup(gctx, helpers)
	router.Get("/sonarr/settings", ctx(sonarr.GetSonarrSettings))
	router.Get("/sonarr/profiles", ctx(sonarr.GetSonarrProfiles))
	router.Get("/sonarr/rootfolders", ctx(sonarr.GetSonarrRootFolders))

	trakt := trakt.NewRouteGroup(gctx, helpers)
	router.Get("/trakt/settings", ctx(trakt.GetTraktSettings))
	router.Put("/trakt/settings", ctx(trakt.UpdateTraktSettings))

	jobs := jobs.NewRouteGroup(gctx, helpers, scheduler)
	router.Get("/jobs/status", ctx(jobs.GetJobStatus))
}
