package candidate

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_recruitment_candidate"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/candidate"
)

type List struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewList(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *List {
	return &List{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// List 候选人分页列表，支持按 姓名/手机号（模糊）、来源渠道、候选人状态 筛选
// （demo-inventory §5：筛选姓名/手机号/来源渠道）。候选人池对 HR 全量可见，无 mine 过滤.
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	chain := condition.NewChain().
		Like(hrm_recruitment_candidate.Name, "%"+req.Name+"%", condition.WithSkip(req.Name == "")).
		Like(hrm_recruitment_candidate.Phone, "%"+req.Phone+"%", condition.WithSkip(req.Phone == "")).
		Equal(hrm_recruitment_candidate.SourceChannel, req.SourceChannel, condition.WithSkip(req.SourceChannel == "")).
		Equal(hrm_recruitment_candidate.Status, req.Status, condition.WithSkip(req.Status == "")).
		Page(req.Current, req.Size).
		OrderByDesc(hrm_recruitment_candidate.Id)

	rows, total, err := l.svcCtx.Model.HrmRecruitmentCandidate.PageByCondition(l.ctx, nil, chain.Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.HrmCandidate, 0, len(rows))
	for _, row := range rows {
		records = append(records, candidateRecord(row))
	}

	return &types.ListResponse{PageResponse: types.PageResponse{
		Current: req.Current,
		Size:    req.Size,
		Total:   total,
	}, Records: records}, nil
}
