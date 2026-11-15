package tax_special_deduction

import (
	"context"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_tax_special_deduction"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/tax_special_deduction"
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

// List 专项附加扣除分页列表: 按员工姓名（快照列 Like）/扣除年度/扣除项/来源筛选；
// 列表按 员工+年度+月份 排序（逐月维护视角），身份证从员工表带出（demo-inventory 列：姓名、身份证…）。
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	rows, total, err := l.svcCtx.Model.HrmTaxSpecialDeduction.PageByCondition(l.ctx, nil,
		condition.NewChain().
			Page(req.Current, req.Size).
			Like(hrm_tax_special_deduction.EmployeeName, "%"+req.EmployeeName+"%", condition.WithSkip(req.EmployeeName == "")).
			Equal(hrm_tax_special_deduction.DeductYear, req.DeductYear, condition.WithSkip(req.DeductYear == "")).
			Equal(hrm_tax_special_deduction.DeductType, req.DeductType, condition.WithSkip(req.DeductType == "")).
			Equal(hrm_tax_special_deduction.Source, req.Source, condition.WithSkip(req.Source == "")).
			OrderByAsc(hrm_tax_special_deduction.EmployeeUuid).
			OrderByAsc(hrm_tax_special_deduction.DeductYear).
			OrderByAsc(hrm_tax_special_deduction.DeductMonth).
			OrderByAsc(hrm_tax_special_deduction.Id).
			Build()...)
	if err != nil {
		return nil, err
	}

	// 身份证快照列不在本表（明细按月展开，员工可能上千行），从员工表 map 带出
	idCardByUuid := make(map[string]string)
	if len(rows) > 0 {
		employees, err := l.svcCtx.Model.HrmEmployee.FindByCondition(l.ctx, nil)
		if err != nil {
			return nil, err
		}
		for _, emp := range employees {
			idCardByUuid[emp.Uuid] = emp.IdCard
		}
	}

	records := make([]types.HrmTaxSpecialDeductionRow, 0, len(rows))
	for _, row := range rows {
		records = append(records, types.HrmTaxSpecialDeductionRow{
			Uuid:         row.Uuid,
			EmployeeUuid: row.EmployeeUuid,
			EmployeeName: row.EmployeeName,
			IdCard:       idCardByUuid[row.EmployeeUuid],
			DeductYear:   row.DeductYear,
			DeductMonth:  row.DeductMonth,
			DeductType:   row.DeductType,
			Amount:       row.Amount,
			Source:       row.Source,
			Remark:       row.Remark,
			CreateTime:   row.CreateTime.Format(time.DateTime),
			UpdateTime:   row.UpdateTime.Format(time.DateTime),
		})
	}

	return &types.ListResponse{PageResponse: types.PageResponse{
		Current: req.Current,
		Size:    req.Size,
		Total:   total,
	}, Records: records}, nil
}
