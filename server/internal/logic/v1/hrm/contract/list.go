package contract

import (
	"context"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee_contract"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/contract"
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

// List 劳动合同分页列表（到期预警页 + 待续签页共享端点，spec 用户故事 14/15）：
// 支持姓名/工号（快照模糊）、到期范围（end_date 区间）筛选；
// 出参含 距到期天数 daysToExpiry 与预警状态 expiryState（口径见 base.go）。
// 预警状态筛选（expiryState）在内存中过滤——口径是纯函数，SQL 层无法直接表达。
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	chain := condition.NewChain().
		Page(req.Current, req.Size).
		Like(hrm_employee_contract.EmployeeName, "%"+req.EmployeeName+"%", condition.WithSkip(req.EmployeeName == "")).
		Like(hrm_employee_contract.EmployeeNo, "%"+req.EmployeeNo+"%", condition.WithSkip(req.EmployeeNo == ""))

	// 到期范围筛选（合同到期日 end_date ∈ [startDate, endDate]）
	if req.StartDate != "" && validDate(req.StartDate) {
		chain = chain.GreaterEqualThan(hrm_employee_contract.EndDate, req.StartDate)
	}
	if req.EndDate != "" && validDate(req.EndDate) {
		chain = chain.LessEqualThan(hrm_employee_contract.EndDate, req.EndDate)
	}
	chain = chain.OrderByDesc(hrm_employee_contract.Id)

	rows, total, err := l.svcCtx.Model.HrmEmployeeContract.PageByCondition(l.ctx, nil, chain.Build()...)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	records := make([]types.HrmEmployeeContract, 0, len(rows))
	for _, row := range rows {
		out := contractToType(now, row)
		// 预警状态口径筛选（内存过滤；过期行数与分页一致性问题可接受，预警页数据量小）
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

func contractToType(now time.Time, row *hrm_employee_contract.HrmEmployeeContract) types.HrmEmployeeContract {
	days, _ := DaysUntilDate(now, row.EndDate)
	return types.HrmEmployeeContract{
		Uuid:         row.Uuid,
		EmployeeUuid: row.EmployeeUuid,
		EmployeeNo:   row.EmployeeNo,
		EmployeeName: row.EmployeeName,
		OrgUuid:      row.OrgUuid,
		ContractNo:   row.ContractNo,
		StartDate:    row.StartDate,
		EndDate:      row.EndDate,
		RenewCount:   row.RenewCount,
		Status:       row.Status,
		DaysToExpiry: days,
		ExpiryState:  ContractExpiryState(row.Status, days),
		Remark:       row.Remark,
		CreateTime:   row.CreateTime.Format("2006-01-02 15:04:05"),
		UpdateTime:   row.UpdateTime.Format("2006-01-02 15:04:05"),
	}
}
