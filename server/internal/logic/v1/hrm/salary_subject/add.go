package salary_subject

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_subject"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_subject"
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
		svcCtx: svcCtx,
		r:      r,
	}
}

// Add 新增薪酬主体（GLOSSARY「薪酬主体」）: 发薪的法人实体，关联一个公司；
// 关联公司名称以快照方式落库，列表/详情不回查组织表。
func (l *Add) Add(req *types.AddRequest) (resp *types.AddResponse, err error) {
	if req.Status == "" {
		req.Status = "1"
	}
	if err = validateSubjectCommon(req.SubjectName, req.Status); err != nil {
		return nil, err
	}
	if req.OrgUuid == "" {
		return nil, status.ErrorMessage(errcodes.SalarySubjectInvalidCode, "关联公司不能为空")
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

	_, err = l.svcCtx.Model.HrmSalarySubject.Insert(l.ctx, nil, &hrm_salary_subject.HrmSalarySubject{
		Uuid:        uuid.New().String(),
		SubjectName: req.SubjectName,
		OrgUuid:     req.OrgUuid,
		OrgName:     org.OrgName,
		PayAccount:  req.PayAccount,
		PayBank:     req.PayBank,
		Status:      req.Status,
		Remark:      req.Remark,
	})
	if err != nil {
		return nil, err
	}

	return &types.AddResponse{}, nil
}
