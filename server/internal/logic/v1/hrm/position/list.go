package position

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_recruitment_position"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_recruitment_requirement"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/position"
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

// List 招聘职位分页列表，支持按 岗位名称（模糊）/来源需求单/岗位类别/状态 筛选
// （demo-inventory §5：岗位名称/来源需求单/岗位类别）。已录用人数联动由
// IncreaseHiredCount（HRM-27 契约）维护。
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	chain := condition.NewChain().
		Like(hrm_recruitment_position.PostName, "%"+req.PostName+"%", condition.WithSkip(req.PostName == "")).
		Equal(hrm_recruitment_position.RequirementUuid, req.RequirementUuid, condition.WithSkip(req.RequirementUuid == "")).
		Equal(hrm_recruitment_position.PostCategory, req.PostCategory, condition.WithSkip(req.PostCategory == "")).
		Equal(hrm_recruitment_position.Status, req.Status, condition.WithSkip(req.Status == "")).
		Page(req.Current, req.Size).
		OrderByDesc(hrm_recruitment_position.Id)

	rows, total, err := l.svcCtx.Model.HrmRecruitmentPosition.PageByCondition(l.ctx, nil, chain.Build()...)
	if err != nil {
		return nil, err
	}

	// 来源需求单单据编号快照（列表展示用，批查避免 N+1）
	billNos := l.requirementBillNos(rows)

	records := make([]types.HrmPosition, 0, len(rows))
	for _, row := range rows {
		records = append(records, positionRecord(row, billNos[row.RequirementUuid]))
	}

	return &types.ListResponse{PageResponse: types.PageResponse{
		Current: req.Current,
		Size:    req.Size,
		Total:   total,
	}, Records: records}, nil
}

// requirementBillNos batch-loads bill_no for the referenced 需求单 (best-effort:
// a deleted requirement simply leaves the snapshot empty).
func (l *List) requirementBillNos(rows []*hrm_recruitment_position.HrmRecruitmentPosition) map[string]string {
	result := make(map[string]string, len(rows))
	uuids := make([]string, 0, len(rows))
	seen := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		if row.RequirementUuid == "" {
			continue
		}
		if _, ok := seen[row.RequirementUuid]; ok {
			continue
		}
		seen[row.RequirementUuid] = struct{}{}
		uuids = append(uuids, row.RequirementUuid)
	}
	if len(uuids) == 0 {
		return result
	}
	reqs, err := l.svcCtx.Model.HrmRecruitmentRequirement.FindByCondition(l.ctx, nil,
		condition.NewChain().In(hrm_recruitment_requirement.Uuid, uuids).Build()...)
	if err != nil {
		l.Errorf("position: load requirement bill_no snapshot: %v", err)
		return result
	}
	for _, req := range reqs {
		result[req.Uuid] = req.BillNo
	}
	return result
}
