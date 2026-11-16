package salary_confirm

import (
	"context"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	hrmuser "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/hrmuser"
	model "github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_confirm_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_confirm"
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

// List 薪资确认审批单分页列表（batchUuid/status 筛选）。默认按「我的单据」过滤
// （创建人=当前用户，spec 用户故事 4）；超级管理员可看全部，传 mine=1 时管理员
// 也只看自己的单据。
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	user := hrmuser.Current(l.ctx)
	onlyMine := req.Mine == 1 || !user.IsSuperAdmin()

	chain := condition.NewChain().
		Equal(model.CreateBy, user.Uuid, condition.WithSkip(!onlyMine)).
		Equal(model.BatchUuid, req.BatchUuid, condition.WithSkip(req.BatchUuid == "")).
		Equal(model.Status, req.Status, condition.WithSkip(req.Status == "")).
		OrderByDesc(model.Id).
		Page(req.Current, req.Size)

	rows, total, err := l.svcCtx.Model.HrmSalaryConfirmBill.PageByCondition(l.ctx, nil, chain.Build()...)
	if err != nil {
		return nil, err
	}
	records := make([]types.HrmSalaryConfirmBill, 0, len(rows))
	for _, row := range rows {
		records = append(records, billRowToRecord(row))
	}
	return &types.ListResponse{PageResponse: types.PageResponse{
		Current: req.Current,
		Size:    req.Size,
		Total:   total,
	}, Records: records}, nil
}

// billRowToRecord 审批单行 → API record.
func billRowToRecord(row *model.HrmSalaryConfirmBill) types.HrmSalaryConfirmBill {
	return types.HrmSalaryConfirmBill{
		Uuid:        row.Uuid,
		BillNo:      row.BillNo,
		BatchUuid:   row.BatchUuid,
		BatchNo:     row.BatchNo,
		BatchName:   row.BatchName,
		SubjectUuid: row.SubjectUuid,
		SubjectName: row.SubjectName,
		BatchYear:   row.BatchYear,
		BatchMonth:  row.BatchMonth,
		Headcount:   row.Headcount,
		NetTotal:    row.NetTotal,
		Status:      row.Status,
		SubmitBy:    row.SubmitBy,
		SubmitTime:  row.SubmitTime,
		ApproveBy:   row.ApproveBy,
		ApproveTime: row.ApproveTime,
		Opinion:     row.Opinion,
		CreatorUuid: row.CreateBy,
		CreatorName: row.CreatorName,
		CreateTime:  row.CreateTime.Format(time.DateTime),
		UpdateTime:  row.UpdateTime.Format(time.DateTime),
	}
}
