package entry

import (
	"context"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	hrmuser "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/hrmuser"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_entry_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/entry"
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

// List 员工入职单分页列表，支持按单据编号/单据状态/姓名筛选 (demo-inventory 3.1).
// 默认按「我的单据」过滤（创建人=当前用户，spec 用户故事 4）；
// 超级管理员可看全部，传 mine=1 时管理员也只看自己的单据.
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	user := hrmuser.Current(l.ctx)
	onlyMine := req.Mine == 1 || !user.IsSuperAdmin()

	chain := condition.NewChain().
		Equal(hrm_entry_bill.CreateBy, user.Uuid, condition.WithSkip(!onlyMine)).
		Like(hrm_entry_bill.BillNo, "%"+req.BillNo+"%", condition.WithSkip(req.BillNo == "")).
		Equal(hrm_entry_bill.Status, req.Status, condition.WithSkip(req.Status == "")).
		Like(hrm_entry_bill.Name, "%"+req.Name+"%", condition.WithSkip(req.Name == "")).
		Page(req.Current, req.Size).
		OrderByDesc(hrm_entry_bill.Id)

	rows, total, err := l.svcCtx.Model.HrmEntryBill.PageByCondition(l.ctx, nil, chain.Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.HrmEntryBill, 0, len(rows))
	for _, row := range rows {
		records = append(records, entryRecord(row))
	}

	return &types.ListResponse{PageResponse: types.PageResponse{
		Current: req.Current,
		Size:    req.Size,
		Total:   total,
	}, Records: records}, nil
}

func entryRecord(row *hrm_entry_bill.HrmEntryBill) types.HrmEntryBill {
	return types.HrmEntryBill{
		Uuid:             row.Uuid,
		BillNo:           row.BillNo,
		Status:           row.Status,
		Name:             row.Name,
		Phone:            row.Phone,
		Email:            row.Email,
		Gender:           row.Gender,
		Birthday:         row.Birthday,
		Nation:           row.Nation,
		PoliticalStatus:  row.PoliticalStatus,
		MaritalStatus:    row.MaritalStatus,
		CultureLevel:     row.CultureLevel,
		NativePlace:      row.NativePlace,
		Address:          row.Address,
		HouseholdAddress: row.HouseholdAddress,
		EmergencyContact: row.EmergencyContact,
		EmergencyPhone:   row.EmergencyPhone,
		BankName:         row.BankName,
		BankAccount:      row.BankAccount,
		Avatar:           row.Avatar,
		Position:         row.Position,
		Post:             row.Post,
		ProbationMonths:  row.ProbationMonths,
		OrgUuid:          row.OrgUuid,
		CompanyUuid:      row.CompanyUuid,
		HireDate:         row.HireDate,
		ExpectFormalDate: row.ExpectFormalDate,
		AttachmentDesc:   row.AttachmentDesc,
		Remark:           row.Remark,
		EmployeeUuid:     row.EmployeeUuid,
		CreatorUuid:      row.CreateBy,
		CreatorName:      row.CreatorName,
		CreateTime:       row.CreateTime.Format(time.DateTime),
		UpdateTime:       row.UpdateTime.Format(time.DateTime),
	}
}
