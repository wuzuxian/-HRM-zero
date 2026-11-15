package punch

// Delete 管理端删除打卡记录（demo-inventory 4.4 行操作「删除」+ 批量删除）。

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_punch_record"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/punch"
)

type Delete struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewDelete(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Delete {
	return &Delete{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *Delete) Delete(req *types.DeleteRequest) (resp *types.DeleteResponse, err error) {
	if len(req.Uuids) == 0 {
		return nil, status.ErrorMessage(errcodes.PunchInvalidCode, "请选择要删除的打卡记录")
	}
	if err := l.svcCtx.Model.HrmPunchRecord.DeleteByCondition(l.ctx, nil,
		condition.NewChain().In(hrm_punch_record.Uuid, req.Uuids).Build()...); err != nil {
		return nil, err
	}
	return &types.DeleteResponse{}, nil
}
