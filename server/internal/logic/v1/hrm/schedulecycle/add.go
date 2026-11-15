package schedulecycle

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_schedule_cycle"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/schedulecycle"
)

type Add struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewAdd(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Add {
	return &Add{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx, r: r,
	}
}

// Add 新增排班周期（GLOSSARY「排班周期」）: 有序班次序列/适用考勤组/启停/备注.
// 周期天数 = 班次序列长度（派生值, 不落库）.
func (l *Add) Add(req *types.AddRequest) (resp *types.AddResponse, err error) {
	if req.Status == "" {
		req.Status = statusEnabled
	}
	if err = validateCycleBasic(l.svcCtx, cycleFieldsFrom(req.CycleName, req.ShiftUuids, req.GroupUuids, req.Status)); err != nil {
		return nil, err
	}

	if _, err = l.svcCtx.Model.HrmScheduleCycle.Insert(l.ctx, nil, &hrm_schedule_cycle.HrmScheduleCycle{
		Uuid:       uuid.New().String(),
		CycleName:  req.CycleName,
		ShiftUuids: joinCsv(req.ShiftUuids),
		GroupUuids: joinCsv(req.GroupUuids),
		Status:     req.Status,
		Remark:     req.Remark,
	}); err != nil {
		return nil, err
	}

	return &types.AddResponse{}, nil
}
