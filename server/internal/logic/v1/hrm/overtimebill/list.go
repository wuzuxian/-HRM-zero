package overtimebill

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	hrmuser "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/hrmuser"
	model "github.com/jzero-io/jzero-admin/server/internal/model/hrm_overtime_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/overtimebill"
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

// List 加班申请单分页列表，支持按 单据编号/单据状态/当天类型 筛选.
// 默认按「我的单据」过滤（创建人=当前用户，spec 用户故事 4）；
// 超级管理员可看全部，传 mine=1 时管理员也只看自己的单据.
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	user := hrmuser.Current(l.ctx)
	onlyMine := req.Mine == 1 || !user.IsSuperAdmin()

	chain := condition.NewChain().
		Equal(model.CreateBy, user.Uuid, condition.WithSkip(!onlyMine)).
		Like(model.BillNo, "%"+req.BillNo+"%", condition.WithSkip(req.BillNo == "")).
		Equal(model.Status, req.Status, condition.WithSkip(req.Status == "")).
		Equal(model.DayType, req.DayType, condition.WithSkip(req.DayType == "")).
		Page(req.Current, req.Size).
		OrderByDesc(model.Id)

	rows, total, err := l.svcCtx.Model.HrmOvertimeBill.PageByCondition(l.ctx, nil, chain.Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.HrmOvertimeBill, 0, len(rows))
	for _, row := range rows {
		records = append(records, overtimeRecord(row))
	}

	return &types.ListResponse{PageResponse: types.PageResponse{
		Current: req.Current,
		Size:    req.Size,
		Total:   total,
	}, Records: records}, nil
}

func overtimeRecord(row *model.HrmOvertimeBill) types.HrmOvertimeBill {
	return types.HrmOvertimeBill{
		Uuid:           row.Uuid,
		BillNo:         row.BillNo,
		Status:         row.Status,
		EmployeeUuid:   row.EmployeeUuid,
		EmployeeNo:     row.EmployeeNo,
		EmployeeName:   row.EmployeeName,
		OrgUuid:        row.OrgUuid,
		OrgName:        row.OrgName,
		OvertimeDate:   row.OvertimeDate,
		DayType:        row.DayType,
		StartTime:      row.StartTime,
		EndTime:        row.EndTime,
		ApplyHours:     row.ApplyHours,
		Reason:         row.Reason,
		Source:         row.Source,
		RuleSource:     row.RuleSource,
		CompMultiplier: row.CompMultiplier,
		ConfirmHours:   row.ConfirmHours,
		CompType:       row.CompType,
		AccrueHours:    row.AccrueHours,
		Settled:        row.Settled,
		CreatorUuid:    row.CreateBy,
		CreatorName:    row.CreatorName,
		CreateTime:     row.CreateTime.Format("2006-01-02 15:04:05"),
		UpdateTime:     row.UpdateTime.Format("2006-01-02 15:04:05"),
	}
}
