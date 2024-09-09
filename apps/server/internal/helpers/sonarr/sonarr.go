package sonarr

import "github.com/mahcks/blockbusterr/internal/global"

func Setup(gctx global.Context) (Service, error) {
	svc := &sonarrService{
		gctx: gctx,
	}

	return svc, nil
}
