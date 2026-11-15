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

type Edit struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewEdit(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Edit {
	return &Edit{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Edit 编辑排班周期（覆盖全部可编辑字段）.
func (l *Edit) Edit(req *types.EditRequest) (resp *types.EditResponse, err error) {
	if req.Uuid == "" {
		return nil, status.ErrorMessage(errcodes.ScheduleCycleNotFoundCode, "排班周期不存在")
	}
	if err = validateCycleBasic(l.svcCtx, cycleFieldsFrom(req.CycleName, req.ShiftUuids, req.GroupUuids, req.Status)); err != nil {
		return nil, err
	}

	row, err := l.svcCtx.Model.HrmScheduleCycle.FindOneByUuid(l.ctx, nil, req.Uuid)
	if err != nil {
		if err == hrm_schedule_cycle.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.ScheduleCycleNotFoundCode, "排班周期不存在")
		}
		return nil, err
	}
	row.CycleName = req.CycleName
	row.ShiftUuids = joinCsv(req.ShiftUuids)
	row.GroupUuids = joinCsv(req.GroupUuids)
	row.Status = req.Status
	row.Remark = req.Remark

	if err = l.svcCtx.Model.HrmScheduleCycle.Update(l.ctx, nil, row); err != nil {
		return nil, err
	}
	return &types.EditResponse{}, nil
}
