package regularization

import (
	"context"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	hrmuser "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/hrmuser"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_regularization_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/regularization"
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

// List 员工转正单分页列表，支持按单据编号/单据状态/员工工号筛选
// (demo-inventory §3.2). 默认按「我的单据」过滤（创建人=当前用户，spec 用户故事 4）；
// 管理员级角色（持共享审批码 v1:hrm:bill:approve）可看全部，传 mine=1 时管理员也只看自己的单据.
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	user := hrmuser.Current(l.ctx)
	onlyMine := req.Mine == 1 || !hrmuser.CanSeeAllBills(l.ctx, l.svcCtx, user)

	chain := condition.NewChain().
		Equal(hrm_regularization_bill.CreateBy, user.Uuid, condition.WithSkip(!onlyMine)).
		Like(hrm_regularization_bill.BillNo, "%"+req.BillNo+"%", condition.WithSkip(req.BillNo == "")).
		Equal(hrm_regularization_bill.Status, req.Status, condition.WithSkip(req.Status == "")).
		Like(hrm_regularization_bill.EmployeeNo, "%"+req.EmployeeNo+"%", condition.WithSkip(req.EmployeeNo == "")).
		Page(req.Current, req.Size).
		OrderByDesc(hrm_regularization_bill.Id)

	rows, total, err := l.svcCtx.Model.HrmRegularizationBill.PageByCondition(l.ctx, nil, chain.Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.HrmRegularization, 0, len(rows))
	for _, row := range rows {
		records = append(records, regularizationRecord(row))
	}

	return &types.ListResponse{PageResponse: types.PageResponse{
		Current: req.Current,
		Size:    req.Size,
		Total:   total,
	}, Records: records}, nil
}

func regularizationRecord(row *hrm_regularization_bill.HrmRegularizationBill) types.HrmRegularization {
	return types.HrmRegularization{
		Uuid:               row.Uuid,
		BillNo:             row.BillNo,
		Status:             row.Status,
		EmployeeUuid:       row.EmployeeUuid,
		EmployeeNo:         row.EmployeeNo,
		EmployeeName:       row.EmployeeName,
		OrgUuid:            row.OrgUuid,
		CompanyUuid:        row.CompanyUuid,
		Position:           row.Position,
		Post:               row.Post,
		PersonStatus:       row.PersonStatus,
		HireDate:           row.HireDate,
		ExpectedFormalDate: row.ExpectedFormalDate,
		FormalDate:         row.FormalDate,
		Remark:             row.Remark,
		Applied:            row.Applied,
		CreatorUuid:        row.CreateBy,
		CreatorName:        row.CreatorName,
		CreateTime:         row.CreateTime.Format(time.DateTime),
		UpdateTime:         row.UpdateTime.Format(time.DateTime),
	}
}
