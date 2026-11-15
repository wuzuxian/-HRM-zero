package candidate

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/candidate"
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

// Edit 编辑候选人档案字段（姓名/联系方式/来源渠道/学历/当前公司/岗位/期望薪资/备注）.
// 状态不在此接口修改（走 updateCandidateStatus 流转）；候选人可编辑，不限状态.
func (l *Edit) Edit(req *types.EditRequest) (resp *types.EditResponse, err error) {
	if err = validateCandidate(req.Name, req.SourceChannel, req.Education); err != nil {
		return nil, err
	}

	row, err := ensureCandidate(&l.Logger, l.ctx, l.svcCtx, req.Uuid)
	if err != nil {
		return nil, err
	}

	row.Name = req.Name
	row.Phone = req.Phone
	row.Email = req.Email
	if req.SourceChannel != "" {
		row.SourceChannel = req.SourceChannel
	}
	row.Education = req.Education
	row.CurrentCompany = req.CurrentCompany
	row.CurrentPosition = req.CurrentPosition
	row.ExpectSalary = req.ExpectSalary
	row.Remark = req.Remark

	if err = l.svcCtx.Model.HrmRecruitmentCandidate.Update(l.ctx, nil, row); err != nil {
		return nil, err
	}

	return &types.EditResponse{}, nil
}
