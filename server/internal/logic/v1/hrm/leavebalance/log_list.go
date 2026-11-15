package leavebalance

import (
	"context"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_leave_balance_log"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/leavebalance"
)

type LogList struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewLogList(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *LogList {
	return &LogList{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx, r: r,
	}
}

// LogList 假期余额流水查询（demo-inventory 4.4 行操作「流水」）：
// 按余额行 uuid 或 (员工/年度/假期类型) 过滤，时间倒序分页。
// 每条含 变更类型/数量/变更后三值/操作人/时间/关联单据。
func (l *LogList) LogList(req *types.LogListRequest) (resp *types.LogListResponse, err error) {
	chain := condition.NewChain().
		Page(req.Current, req.Size).
		Equal(hrm_leave_balance_log.BalanceUuid, req.BalanceUuid, condition.WithSkip(req.BalanceUuid == "")).
		Equal(hrm_leave_balance_log.Year, req.Year, condition.WithSkip(req.Year <= 0)).
		Equal(hrm_leave_balance_log.EmployeeUuid, req.EmployeeUuid, condition.WithSkip(req.EmployeeUuid == "")).
		Equal(hrm_leave_balance_log.LeaveTypeCode, req.LeaveTypeCode, condition.WithSkip(req.LeaveTypeCode == ""))

	rows, total, err := l.svcCtx.Model.HrmLeaveBalanceLog.PageByCondition(l.ctx, nil,
		chain.OrderByDesc(hrm_leave_balance_log.Id).Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.HrmLeaveBalanceLog, 0, len(rows))
	for _, row := range rows {
		records = append(records, types.HrmLeaveBalanceLog{
			Uuid:          row.Uuid,
			BalanceUuid:   row.BalanceUuid,
			Year:          row.Year,
			EmployeeUuid:  row.EmployeeUuid,
			LeaveTypeCode: row.LeaveTypeCode,
			LeaveTypeName: leaveTypeName(l.ctx, l.svcCtx, row.LeaveTypeCode),
			ChangeType:    row.ChangeType,
			Days:          row.Days,
			GrantedAfter:  row.GrantedAfter,
			UsedAfter:     row.UsedAfter,
			FrozenAfter:   row.FrozenAfter,
			Reason:        row.Reason,
			BillUuid:      row.BillUuid,
			OperatorUuid:  row.OperatorUuid,
			OperatorName:  row.OperatorName,
			CreateTime:    row.CreateTime.Format(time.DateTime),
		})
	}

	return &types.LogListResponse{
		PageResponse: types.PageResponse{Current: req.Current, Size: req.Size, Total: total},
		Records:      records,
	}, nil
}
