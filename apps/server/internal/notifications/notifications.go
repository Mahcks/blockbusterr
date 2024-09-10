package notifications

import (
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/internal/notifications/discord"
)

type Providers struct {
	Discord discord.Service
}

// Setup the notifications struct, setting up Discord service
func Setup(gctx global.Context) (*Providers, error) {
	discordService, err := discord.Setup(gctx)
	if err != nil {
		return nil, err
	}

	return &Providers{
		Discord: discordService,
	}, nil
}
