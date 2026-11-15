package empcert

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/empcert"
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

func (l *Edit) Edit(req *types.EditRequest) (resp *types.EditResponse, err error) {
	if req.Uuid == "" {
		return nil, status.ErrorMessage(errcodes.EmployeeCertificateNotFoundCode, "uuid 不能为空")
	}
	row, err := findCertByUuid(l.ctx, l.svcCtx, req.Uuid)
	if err != nil {
		return nil, err
	}

	// 前端回显的是脱敏证号；若提交值 == 脱敏值 则视为未修改，避免把星号落库。
	if _, err = validateCertInput(l.ctx, l.svcCtx, row.EmployeeUuid, req.CertType, req.CertNo); err != nil {
		return nil, err
	}
	if req.IssueDate != "" && !validDate(req.IssueDate) {
		return nil, status.ErrorMessage(errcodes.EmployeeCertificateInvalidCode, "发证日期格式须为 YYYY-MM-DD")
	}
	if req.ExpiryDate != "" && !validDate(req.ExpiryDate) {
		return nil, status.ErrorMessage(errcodes.EmployeeCertificateInvalidCode, "到期日格式须为 YYYY-MM-DD")
	}

	if req.CertNo != MaskCertNo(row.CertNo) {
		row.CertNo = req.CertNo
	}
	row.CertType = req.CertType
	row.IssueDate = req.IssueDate
	row.ExpiryDate = req.ExpiryDate
	row.Remark = req.Remark

	if err = l.svcCtx.Model.HrmEmployeeCertificate.Update(l.ctx, nil, row); err != nil {
		return nil, err
	}

	return &types.EditResponse{}, nil
}
