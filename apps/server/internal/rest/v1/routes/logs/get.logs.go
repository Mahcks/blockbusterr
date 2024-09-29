package logs

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/mahcks/blockbusterr/internal/rest/v1/respond"
	"github.com/mahcks/blockbusterr/pkg/errors"
)

func (rg *RouteGroup) GetLogs(ctx *respond.Ctx) error {
	// Default values
	take := 30
	skip := 0
	filter := ""
	search := ""

	// Get query parameters
	if param := ctx.Query("take"); param != "" {
		if parsedTake, err := strconv.Atoi(param); err == nil {
			take = parsedTake
		}
	}

	if param := ctx.Query("skip"); param != "" {
		if parsedSkip, err := strconv.Atoi(param); err == nil {
			skip = parsedSkip
		}
	}

	if param := ctx.Query("filter"); param != "" {
		fmt.Println("SEARCH PARAM", param)
		if param == "all" {
			param = ""
		} else {
			filter = param
		}
	}

	if param := ctx.Query("search"); param != "" {
		search = param
	}

	// Call the database function to get the logs
	logs, err := rg.gctx.Crate().SQL.Queries().GetLogs(ctx.Context(), take, skip, filter, search)
	if err != nil {
		slog.Error("Error getting logs", "error", err)
		return errors.ErrInternalServerError().SetDetail("Failed to get logs")
	}

	// For example, returning a response
	return ctx.JSON(logs)
}
