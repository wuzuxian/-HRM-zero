package salary_subject

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_subject"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_subject"
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

// Edit 编辑薪酬主体: 关联公司可换（同样须为公司节点），公司名称快照随换随更.
func (l *Edit) Edit(req *types.EditRequest) (resp *types.EditResponse, err error) {
	if err = validateSubjectCommon(req.SubjectName, req.Status); err != nil {
		return nil, err
	}
	if req.OrgUuid == "" {
		return nil, status.ErrorMessage(errcodes.SalarySubjectInvalidCode, "关联公司不能为空")
	}

	row, err := l.svcCtx.Model.HrmSalarySubject.FindOneByUuid(l.ctx, nil, req.Uuid)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.SalarySubjectNotFoundCode, "薪酬主体不存在")
		}
		return nil, err
	}

	// 关联公司必须存在且为公司节点（org_type=1）
	org, err := l.svcCtx.Model.HrmOrg.FindOneByUuid(l.ctx, nil, req.OrgUuid)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.SalarySubjectOrgNotFoundCode, "关联公司不存在")
		}
		return nil, err
	}
	if org.OrgType != orgTypeCompany {
		return nil, status.ErrorMessage(errcodes.SalarySubjectOrgInvalidCode, "关联组织不是公司节点（须为公司）")
	}

	err = l.svcCtx.Model.HrmSalarySubject.UpdateFieldsByCondition(l.ctx, nil, map[string]any{
		string(hrm_salary_subject.SubjectName): req.SubjectName,
		string(hrm_salary_subject.OrgUuid):     req.OrgUuid,
		string(hrm_salary_subject.OrgName):     org.OrgName,
		string(hrm_salary_subject.PayAccount):  req.PayAccount,
		string(hrm_salary_subject.PayBank):     req.PayBank,
		string(hrm_salary_subject.Status):      req.Status,
		string(hrm_salary_subject.Remark):      req.Remark,
	}, condition.Condition{
		Field:    hrm_salary_subject.Uuid,
		Operator: condition.Equal,
		Value:    row.Uuid,
	})
	if err != nil {
		return nil, err
	}

	return &types.EditResponse{}, nil
}
