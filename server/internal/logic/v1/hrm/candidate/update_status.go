package candidate

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/candidate"
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

// UpdateStatus 候选人状态流转（新线索→已投递→面试中→已发Offer→录用通过）：
// 只能推进到下一阶段（不可跳级/不可回退）；不通过（rejected）可从任意非终态进入；
// 录用通过（hired）为终态，不可再变更.
func (l *UpdateStatus) UpdateStatus(req *types.UpdateStatusRequest) (resp *types.UpdateStatusResponse, err error) {
	if !validCandidateStatus(req.Status) {
		return nil, status.ErrorMessage(errcodes.CandidateInvalidCode, "候选人状态非法")
	}

	row, err := ensureCandidate(&l.Logger, l.ctx, l.svcCtx, req.Uuid)
	if err != nil {
		return nil, err
	}

	if !validStatusTransition(row.Status, req.Status) {
		return nil, status.ErrorMessage(errcodes.CandidateStatusTransitionInvalidCode,
			"候选人状态不允许从当前阶段流转到目标阶段（不可跳级/不可回退）")
	}

	row.Status = req.Status
	if err = l.svcCtx.Model.HrmRecruitmentCandidate.Update(l.ctx, nil, row); err != nil {
		return nil, err
	}

	l.Infof("candidate: 候选人状态流转 uuid=%s -> %s", req.Uuid, req.Status)
	return &types.UpdateStatusResponse{}, nil
}
