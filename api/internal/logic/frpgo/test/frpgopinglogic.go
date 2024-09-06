package test

import (
	"context"

	"frpgo/api/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type FrpgoPingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFrpgoPingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FrpgoPingLogic {
	return &FrpgoPingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FrpgoPingLogic) FrpgoPing() error {
	// todo: add your logic here and delete this line

	return nil
}
