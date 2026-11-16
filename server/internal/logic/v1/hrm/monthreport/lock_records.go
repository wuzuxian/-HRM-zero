package monthreport

// LockRecords 锁定记录（demo-inventory 4.4 按钮）：谁何时锁/解锁/重算流水。

import (
	"context"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_attendance_month_lock"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/monthreport"
)

type LockRecords struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewLockRecords(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *LockRecords {
	return &LockRecords{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// actionName 动作中文名。
func actionName(action string) string {
	switch action {
	case ActionLock:
		return "锁定月报"
	case ActionUnlock:
		return "解锁月报"
	case ActionRecalc:
		return "重新计算"
	}
	return action
}

func (l *LockRecords) LockRecords(req *types.LockRecordsRequest) (resp *types.LockRecordsResponse, err error) {
	chain := condition.NewChain()
	if req.YearMonth != "" {
		chain = chain.Equal(hrm_attendance_month_lock.YearMonth, req.YearMonth)
	}
	rows, err := l.svcCtx.Model.HrmAttendanceMonthLock.FindByCondition(l.ctx, nil,
		chain.OrderByDesc(hrm_attendance_month_lock.Id).Build()...)
	if err != nil {
		return nil, err
	}
	records := make([]types.LockRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, types.LockRecord{
			Uuid:         row.Uuid,
			YearMonth:    row.YearMonth,
			Action:       row.Action,
			ActionName:   actionName(row.Action),
			OperatorName: row.OperatorName,
			Remark:       row.Remark,
			CreateTime:   row.CreateTime.Format(time.DateTime),
		})
	}
	return &types.LockRecordsResponse{Records: records}, nil
}
