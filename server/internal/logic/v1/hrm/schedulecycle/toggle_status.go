package schedulecycle

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_schedule_cycle"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/schedulecycle"
)

type ToggleStatus struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewToggleStatus(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *ToggleStatus {
	return &ToggleStatus{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// ToggleStatus 启用/停用排班周期. 停用后 HRM-15 排班生成不再使用该周期.
func (l *ToggleStatus) ToggleStatus(req *types.ToggleStatusRequest) (resp *types.ToggleStatusResponse, err error) {
	if req.Status != statusEnabled && req.Status != statusDisabled {
		return nil, status.ErrorMessage(errcodes.ScheduleCycleInvalidCode, "排班周期状态必须是 启用 或 停用")
	}
	row, err := l.svcCtx.Model.HrmScheduleCycle.FindOneByUuid(l.ctx, nil, req.Uuid)
	if err != nil {
		if err == hrm_schedule_cycle.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.ScheduleCycleNotFoundCode, "排班周期不存在")
		}
		return nil, err
	}
	row.Status = req.Status
	if err = l.svcCtx.Model.HrmScheduleCycle.Update(l.ctx, nil, row); err != nil {
		return nil, err
	}
	return &types.ToggleStatusResponse{}, nil
}
