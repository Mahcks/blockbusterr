package radarr

import "github.com/mahcks/blockbusterr/internal/global"

func Setup(gctx global.Context) (Service, error) {
	svc := &radarrService{
		gctx: gctx,
	}

	return svc, nil
}
