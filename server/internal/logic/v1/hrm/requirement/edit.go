package requirement

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_recruitment_requirement"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/requirement"
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

// Edit 修改招聘需求单：仅 未提交/审批不通过/已撤销 状态可修改.
func (l *Edit) Edit(req *types.EditRequest) (resp *types.EditResponse, err error) {
	row, err := l.svcCtx.Model.HrmRecruitmentRequirement.FindOneByUuid(l.ctx, nil, req.Uuid)
	if err != nil {
		if err == hrm_recruitment_requirement.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.RequirementNotFoundCode, "招聘需求单不存在")
		}
		return nil, err
	}
	if !canEdit(row.Status) {
		return nil, status.ErrorMessage(errcodes.RequirementNotEditableCode,
			"单据 "+row.BillNo+" 当前状态不允许修改（仅 未提交/审批不通过/已撤销 可改）")
	}
	if err = validateRequirement(req.PostName, req.PostCategory, req.Urgency,
		req.ExpectDate, req.Headcount, req.OrgUuid); err != nil {
		return nil, err
	}

	row.PostName = req.PostName
	row.PostCategory = req.PostCategory
	row.Headcount = req.Headcount
	row.Urgency = req.Urgency
	row.ExpectDate = req.ExpectDate
	row.OrgUuid = req.OrgUuid
	row.OrgName = req.OrgName
	row.RequirementDesc = req.RequirementDesc

	if err = l.svcCtx.Model.HrmRecruitmentRequirement.Update(l.ctx, nil, row); err != nil {
		return nil, err
	}

	l.Infof("requirement: 招聘需求单已修改 billNo=%s", row.BillNo)
	return &types.EditResponse{}, nil
}
