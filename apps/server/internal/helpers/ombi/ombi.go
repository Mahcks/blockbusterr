package ombi

import (
	"github.com/mahcks/blockbusterr/internal/global"
)

func Setup(gctx global.Context) (Service, error) {
	svc := &ombiService{
		gctx: gctx,
	}

	return svc, nil
}
