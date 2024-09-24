package movies

import (
	"database/sql"
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/blockbusterr/internal/rest/v1/respond"
	"github.com/mahcks/blockbusterr/pkg/errors"
	"github.com/mahcks/blockbusterr/pkg/utils"
)

type MovieSettingPaylod struct {
	Anticipated              *int    `json:"anticipated"`
	CronJobAnticipated       *string `json:"cron_job_anticipated"`
	BoxOffice                *int    `json:"box_office"`
	CronJobBoxOffice         *string `json:"cron_job_box_office"`
	Popular                  *int    `json:"popular"`
	CronJobPopular           *string `json:"cron_job_popular"`
	Trending                 *int    `json:"trending"`
	CronJobTrending          *string `json:"cron_job_trending"`
	MaxRuntime               *int    `json:"max_runtime"`
	MinRuntime               *int    `json:"min_runtime"`
	MinYear                  *int    `json:"min_year"`
	MaxYear                  *int    `json:"max_year"`
	AllowedCountries         *string `json:"allowed_countries"`
	AllowedLanguages         *string `json:"allowed_languages"`
	BlacklistedGenres        *string `json:"blacklisted_genres"`
	BlacklistedTitleKeywords *string `json:"blacklisted_title_keywords"`
	BlacklistedTMDBIDs       *string `json:"blacklisted_tmdb_ids"`
}

func (rg *RouteGroup) UpdateMovieSettings(ctx *respond.Ctx) error {
	var payload MovieSettingPaylod
	if err := json.Unmarshal(ctx.Body(), &payload); err != nil {
		return errors.ErrBadRequest().SetDetail("Invalid JSON payload")
	}

	utils.PrettyPrintStruct(payload)
	err := rg.gctx.Crate().SQL.Queries().UpdateMovieSettings(
		ctx.Context(),
		utils.PointerToNullInt32(payload.Anticipated),
		utils.PointerToNullInt32(payload.BoxOffice),
		utils.PointerToNullInt32(payload.Popular),
		utils.PointerToNullInt32(payload.Trending),
		utils.PointerToNullInt32(payload.MaxRuntime),
		utils.PointerToNullInt32(payload.MinRuntime),
		utils.PointerToNullInt32(payload.MinYear),
		utils.PointerToNullInt32(payload.MaxYear),
		sql.NullString{String: "", Valid: false},
		utils.PointerToNullString(payload.CronJobAnticipated),
		utils.PointerToNullString(payload.CronJobBoxOffice),
		utils.PointerToNullString(payload.CronJobPopular),
		utils.PointerToNullString(payload.CronJobTrending),
	)
	if err != nil {
		return errors.ErrInternalServerError().SetDetail("Failed to update movie settings")
	}

	return ctx.JSON(fiber.Map{"success": true})
}
