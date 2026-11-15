package position

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_recruitment_position"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/position"
)

type UpdateStatus struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewUpdateStatus(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *UpdateStatus {
	return &UpdateStatus{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// UpdateStatus 职位状态流转（手工）：open→closed 关闭职位 / open→full 手工置满
// （演示口径）；closed 为终态不可再流转；已满(full)/已关闭(closed) 之间禁止互转。
// 已满自动联动由 IncreaseHiredCount（HRM-27 契约）负责。
func (l *UpdateStatus) UpdateStatus(req *types.UpdateStatusRequest) (resp *types.UpdateStatusResponse, err error) {
	if !validPositionStatus(req.Status) {
		return nil, status.ErrorMessage(errcodes.PositionInvalidCode, "职位状态非法")
	}

	row, err := l.svcCtx.Model.HrmRecruitmentPosition.FindOneByCondition(l.ctx, nil,
		condition.NewChain().Equal(hrm_recruitment_position.Uuid, req.Uuid).Build()...)
	if err != nil {
		return nil, status.ErrorMessage(errcodes.PositionNotFoundCode, "招聘职位不存在")
	}

	if !validStatusTransition(row.Status, req.Status) {
		return nil, status.ErrorMessage(errcodes.PositionStatusTransitionInvalidCode,
			"职位状态不允许从当前状态流转到目标状态（已关闭为终态）")
	}

	row.Status = req.Status
	if err = l.svcCtx.Model.HrmRecruitmentPosition.Update(l.ctx, nil, row); err != nil {
		return nil, err
	}

	l.Infof("position: 职位状态流转 uuid=%s -> %s", req.Uuid, req.Status)
	return &types.UpdateStatusResponse{}, nil
}

// validStatusTransition manual transitions: only open may move (to full/closed);
// full→closed is allowed (提前关闭已满职位); closed is terminal.
func validStatusTransition(from, to string) bool {
	if from == to {
		return false
	}
	switch from {
	case PositionStatusOpen:
		return to == PositionStatusFull || to == PositionStatusClosed
	case PositionStatusFull:
		return to == PositionStatusClosed
	default: // closed terminal
		return false
	}
}
