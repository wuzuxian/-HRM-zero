package resignation

import (
	"context"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	hrmuser "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/hrmuser"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_resignation_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/resignation"
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

// List 员工离职单分页列表，支持按单据编号/单据状态/员工工号筛选.
// 默认按「我的单据」过滤（创建人=当前用户，spec 用户故事 4）；
// 超级管理员可看全部，传 mine=1 时管理员也只看自己的单据.
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	user := hrmuser.Current(l.ctx)
	onlyMine := req.Mine == 1 || !user.IsSuperAdmin()

	chain := condition.NewChain().
		Equal(hrm_resignation_bill.CreateBy, user.Uuid, condition.WithSkip(!onlyMine)).
		Like(hrm_resignation_bill.BillNo, "%"+req.BillNo+"%", condition.WithSkip(req.BillNo == "")).
		Equal(hrm_resignation_bill.Status, req.Status, condition.WithSkip(req.Status == "")).
		Like(hrm_resignation_bill.EmployeeNo, "%"+req.EmployeeNo+"%", condition.WithSkip(req.EmployeeNo == "")).
		Page(req.Current, req.Size).
		OrderByDesc(hrm_resignation_bill.Id)

	rows, total, err := l.svcCtx.Model.HrmResignationBill.PageByCondition(l.ctx, nil, chain.Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.HrmResignation, 0, len(rows))
	for _, row := range rows {
		records = append(records, resignationRecord(row))
	}

	return &types.ListResponse{PageResponse: types.PageResponse{
		Current: req.Current,
		Size:    req.Size,
		Total:   total,
	}, Records: records}, nil
}

func resignationRecord(row *hrm_resignation_bill.HrmResignationBill) types.HrmResignation {
	return types.HrmResignation{
		Uuid:             row.Uuid,
		BillNo:           row.BillNo,
		Status:           row.Status,
		EmployeeUuid:     row.EmployeeUuid,
		EmployeeNo:       row.EmployeeNo,
		EmployeeName:     row.EmployeeName,
		OrgUuid:          row.OrgUuid,
		CompanyUuid:      row.CompanyUuid,
		Position:         row.Position,
		Post:             row.Post,
		ResignType:       row.ResignType,
		ResignReason:     row.ResignReason,
		ApplyDate:        row.ApplyDate,
		ResignDate:       row.ResignDate,
		LastWorkDate:     row.LastWorkDate,
		HandoverPerson:   row.HandoverPerson,
		SalarySettleDate: row.SalarySettleDate,
		Applied:          row.Applied,
		CreatorUuid:      row.CreateBy,
		CreatorName:      row.CreatorName,
		CreateTime:       row.CreateTime.Format(time.DateTime),
		UpdateTime:       row.UpdateTime.Format(time.DateTime),
	}
}
