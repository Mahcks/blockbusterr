package trakt

import (
	"encoding/json"

	"github.com/charmbracelet/log"
	"github.com/mahcks/blockbusterr/internal/rest/v1/respond"
	"github.com/mahcks/blockbusterr/pkg/errors"
	"github.com/mahcks/blockbusterr/pkg/structures"
)

func (rg *RouteGroup) UpdateTraktSettings(ctx *respond.Ctx) error {
	body := ctx.Body()

	var requestSettings structures.TraktSettings
	err := json.Unmarshal(body, &requestSettings)
	if err != nil {
		log.Errorf("error unmarshalling request body: %v", err)
		return errors.ErrBadRequest()
	}

	err = rg.gctx.Crate().SQL.Queries().UpdateTraktSettings(ctx.Context(), requestSettings.ClientID, requestSettings.ClientSecret)
	if err != nil {
		log.Errorf("error updating trakt settings: %v", err)
		return errors.ErrInternalServerError()
	}

	return nil
}
