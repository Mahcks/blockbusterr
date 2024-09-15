package jobs

import "github.com/mahcks/blockbusterr/internal/rest/v1/respond"

func (rg *RouteGroup) GetJobStatus(ctx *respond.Ctx) error {
	return ctx.JSON(rg.scheduler.GetJobStatus())
}
