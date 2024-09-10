package discord

import "github.com/mahcks/blockbusterr/internal/global"

func Setup(gctx global.Context) (Service, error) {
	svc := &discordService{
		gctx: gctx,
	}

	return svc, nil
}
