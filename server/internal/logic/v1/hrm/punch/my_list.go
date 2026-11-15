package punch

// MyList 员工端「考勤打卡」页（demo-inventory 4.1）：我的打卡记录列表，
// 按当前登录人（管理端代查传 employeeUuid）过滤，筛选 日期区间/签到状态。

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/hrmuser"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_punch_record"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/punch"
	"github.com/jzero-io/jzero/core/status"
)

type MyList struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewMyList(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *MyList {
	return &MyList{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *MyList) MyList(req *types.MyListRequest) (resp *types.MyListResponse, err error) {
	// 当前登录人 → 员工档案（未关联 → 10522）
	user := hrmuser.Current(l.ctx)
	if user.Uuid == "" {
		return nil, status.ErrorMessage(errcodes.PunchEmployeeNotFoundCode, "当前登录人未关联员工档案")
	}
	employee, findErr := l.svcCtx.Model.HrmEmployee.FindOneByCondition(l.ctx, nil,
		condition.NewChain().Equal(hrm_employee.UserUuid, user.Uuid).Build()...)
	if findErr != nil {
		if findErr == hrm_employee.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.PunchEmployeeNotFoundCode, "当前登录人未关联员工档案")
		}
		return nil, findErr
	}
	employeeUuid := employee.Uuid

	chain := condition.NewChain().
		Page(req.Current, req.Size).
		Equal(hrm_punch_record.EmployeeUuid, employeeUuid)

	if req.StartDate != "" {
		chain = chain.GreaterEqualThan(hrm_punch_record.PunchDate, req.StartDate)
	}
	if req.EndDate != "" {
		chain = chain.LessEqualThan(hrm_punch_record.PunchDate, req.EndDate)
	}
	if req.ClockInOutStatus == SettleMissing {
		chain = chain.Equal(hrm_punch_record.ClockInStatus, "")
	} else {
		chain = chain.Equal(hrm_punch_record.ClockInStatus, req.ClockInOutStatus, condition.WithSkip(!isPunchStatus(req.ClockInOutStatus)))
	}

	rows, total, err := l.svcCtx.Model.HrmPunchRecord.PageByCondition(l.ctx, nil,
		chain.OrderByDesc(hrm_punch_record.PunchDate).Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.HrmPunchRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, toRecord(l.ctx, l.svcCtx, row))
	}
	return &types.MyListResponse{
		PageResponse: types.PageResponse{Current: req.Current, Size: req.Size, Total: total},
		Records:      records,
	}, nil
}
