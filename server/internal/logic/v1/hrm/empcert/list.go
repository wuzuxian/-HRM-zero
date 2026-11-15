package empcert

import (
	"context"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee_certificate"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/empcert"
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

// List 员工证照分页列表（证照到期预警页，spec 用户故事 16）：
// 支持姓名/工号（快照模糊）、证书类型（模糊）筛选；出参证号统一脱敏（MaskCertNo），
// 距到期天数/预警状态口径与劳动合同一致（≤90 天为 warning）。
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	chain := condition.NewChain().
		Page(req.Current, req.Size).
		Like(hrm_employee_certificate.EmployeeName, "%"+req.EmployeeName+"%", condition.WithSkip(req.EmployeeName == "")).
		Like(hrm_employee_certificate.EmployeeNo, "%"+req.EmployeeNo+"%", condition.WithSkip(req.EmployeeNo == "")).
		Like(hrm_employee_certificate.CertType, "%"+req.CertType+"%", condition.WithSkip(req.CertType == "")).
		OrderByDesc(hrm_employee_certificate.Id)

	rows, total, err := l.svcCtx.Model.HrmEmployeeCertificate.PageByCondition(l.ctx, nil, chain.Build()...)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	records := make([]types.HrmEmployeeCertificate, 0, len(rows))
	for _, row := range rows {
		out := certToType(now, row)
		if req.ExpiryState != "" && out.ExpiryState != req.ExpiryState {
			total--
			continue
		}
		records = append(records, out)
	}

	return &types.ListResponse{PageResponse: types.PageResponse{
		Current: req.Current,
		Size:    req.Size,
		Total:   total,
	}, Records: records}, nil
}

func certToType(now time.Time, row *hrm_employee_certificate.HrmEmployeeCertificate) types.HrmEmployeeCertificate {
	days, _ := DaysUntilDate(now, row.ExpiryDate)
	return types.HrmEmployeeCertificate{
		Uuid:         row.Uuid,
		EmployeeUuid: row.EmployeeUuid,
		EmployeeNo:   row.EmployeeNo,
		EmployeeName: row.EmployeeName,
		OrgUuid:      row.OrgUuid,
		CertType:     row.CertType,
		CertNo:       MaskCertNo(row.CertNo),
		IssueDate:    row.IssueDate,
		ExpiryDate:   row.ExpiryDate,
		DaysToExpiry: days,
		ExpiryState:  ExpiryStateByDays(days),
		Remark:       row.Remark,
		CreateTime:   row.CreateTime.Format("2006-01-02 15:04:05"),
		UpdateTime:   row.UpdateTime.Format("2006-01-02 15:04:05"),
	}
}
