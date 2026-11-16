package requirement

import (
	"context"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	hrmuser "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/hrmuser"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_recruitment_requirement"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/requirement"
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

// List 招聘需求单分页列表，支持按单据编号/单据状态/岗位名称/岗位类别/紧急度/所属部门筛选.
// 默认按「我的单据」过滤（创建人=当前用户，spec 用户故事 4）；
// 管理员级角色（持共享审批码 v1:hrm:bill:approve）可看全部，传 mine=1 时管理员也只看自己的单据.
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	user := hrmuser.Current(l.ctx)
	onlyMine := req.Mine == 1 || !hrmuser.CanSeeAllBills(l.ctx, l.svcCtx, user)

	chain := condition.NewChain().
		Equal(hrm_recruitment_requirement.CreateBy, user.Uuid, condition.WithSkip(!onlyMine)).
		Like(hrm_recruitment_requirement.BillNo, "%"+req.BillNo+"%", condition.WithSkip(req.BillNo == "")).
		Equal(hrm_recruitment_requirement.Status, req.Status, condition.WithSkip(req.Status == "")).
		Like(hrm_recruitment_requirement.PostName, "%"+req.PostName+"%", condition.WithSkip(req.PostName == "")).
		Equal(hrm_recruitment_requirement.PostCategory, req.PostCategory, condition.WithSkip(req.PostCategory == "")).
		Equal(hrm_recruitment_requirement.Urgency, req.Urgency, condition.WithSkip(req.Urgency == "")).
		Equal(hrm_recruitment_requirement.OrgUuid, req.OrgUuid, condition.WithSkip(req.OrgUuid == "")).
		Page(req.Current, req.Size).
		OrderByDesc(hrm_recruitment_requirement.Id)

	rows, total, err := l.svcCtx.Model.HrmRecruitmentRequirement.PageByCondition(l.ctx, nil, chain.Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.HrmRequirement, 0, len(rows))
	for _, row := range rows {
		records = append(records, requirementRecord(row))
	}

	return &types.ListResponse{PageResponse: types.PageResponse{
		Current: req.Current,
		Size:    req.Size,
		Total:   total,
	}, Records: records}, nil
}

func requirementRecord(row *hrm_recruitment_requirement.HrmRecruitmentRequirement) types.HrmRequirement {
	return types.HrmRequirement{
		Uuid:            row.Uuid,
		BillNo:          row.BillNo,
		Status:          row.Status,
		PostName:        row.PostName,
		PostCategory:    row.PostCategory,
		Headcount:       row.Headcount,
		Urgency:         row.Urgency,
		ExpectDate:      row.ExpectDate,
		OrgUuid:         row.OrgUuid,
		OrgName:         row.OrgName,
		RequirementDesc: row.RequirementDesc,
		CreatorUuid:     row.CreateBy,
		CreatorName:     row.CreatorName,
		CreateTime:      row.CreateTime.Format(time.DateTime),
		UpdateTime:      row.UpdateTime.Format(time.DateTime),
	}
}
