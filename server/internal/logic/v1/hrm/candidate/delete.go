package candidate

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_recruitment_candidate"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/candidate"
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

// Delete 删除/批量删除候选人（demo-inventory §5 工具栏 批量删除）.
// 候选人是普通实体（非单据底座），删除不限状态；HRM-25 录用审批如需锁定
// 已被录用审批单引用的候选人，由其侧按 candidateUuid 校验.
func (l *Delete) Delete(req *types.DeleteRequest) (resp *types.DeleteResponse, err error) {
	if len(req.Uuids) == 0 {
		return &types.DeleteResponse{}, nil
	}

	err = l.svcCtx.Model.HrmRecruitmentCandidate.DeleteByCondition(l.ctx, nil, condition.Condition{
		Field:    hrm_recruitment_candidate.Uuid,
		Operator: condition.In,
		Value:    req.Uuids,
	})
	if err != nil {
		return nil, err
	}

	return &types.DeleteResponse{}, nil
}

// ensureCandidate loads a candidate by uuid, mapping ErrNotFound to 10440.
func ensureCandidate(l *logx.Logger, ctx context.Context, svcCtx *svc.ServiceContext, candidateUuid string) (*hrm_recruitment_candidate.HrmRecruitmentCandidate, error) {
	row, err := svcCtx.Model.HrmRecruitmentCandidate.FindOneByUuid(ctx, nil, candidateUuid)
	if err != nil {
		if err == hrm_recruitment_candidate.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.CandidateNotFoundCode, "候选人不存在")
		}
		return nil, err
	}
	return row, nil
}
