package test

import (
	"net/http"

	"frp-go/api/internal/logic/frpgo/test"
	"frp-go/api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func FrpgoPingHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := test.NewFrpgoPingLogic(r.Context(), svcCtx)
		err := l.FrpgoPing()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.Ok(w)
		}
	}
}
