package entry

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	hrmuser "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/hrmuser"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_entry_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/entry"
)

type Add struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewAdd(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Add {
	return &Add{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Add 新增入职单（HR201）：事务内生成单据编号，初始状态 未提交.
// 单据编号/状态列与共享审批底座（HRM-06）对齐，提交/审批走共享端点；
// 审批通过后 OnApproved 事务内自动建员工档案（demo-inventory 3.1）:
// 工号自动生成、预计转正=入职日期+试用期月、三子表落档案（幂等）.
func (l *Add) Add(req *types.AddRequest) (resp *types.AddResponse, err error) {
	if err = validateEntry(req.Name, req.Gender, req.MaritalStatus, req.HireDate,
		req.ExpectFormalDate, req.ProbationMonths); err != nil {
		return nil, err
	}

	user := hrmuser.Current(l.ctx)

	var billNo string
	if txErr := l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		no, err := bill.GenerateBillNo(ctx, session, BillTypeCode, time.Now())
		if err != nil {
			return err
		}
		billUuid := uuid.New().String()
		billNo = no
		err = l.svcCtx.Model.HrmEntryBill.InsertV2(ctx, session, &hrm_entry_bill.HrmEntryBill{
			Uuid:             billUuid,
			BillNo:           no,
			Status:           bill.StatusUnsubmitted,
			Name:             req.Name,
			Phone:            req.Phone,
			Email:            req.Email,
			Gender:           req.Gender,
			Birthday:         req.Birthday,
			Nation:           req.Nation,
			PoliticalStatus:  req.PoliticalStatus,
			MaritalStatus:    req.MaritalStatus,
			CultureLevel:     req.CultureLevel,
			NativePlace:      req.NativePlace,
			Address:          req.Address,
			HouseholdAddress: req.HouseholdAddress,
			EmergencyContact: req.EmergencyContact,
			EmergencyPhone:   req.EmergencyPhone,
			BankName:         req.BankName,
			BankAccount:      req.BankAccount,
			Avatar:           req.Avatar,
			Position:         req.Position,
			Post:             req.Post,
			ProbationMonths:  req.ProbationMonths,
			OrgUuid:          req.OrgUuid,
			CompanyUuid:      req.CompanyUuid,
			HireDate:         req.HireDate,
			ExpectFormalDate: expectFormalOf(req.HireDate, req.ExpectFormalDate, req.ProbationMonths),
			AttachmentDesc:   req.AttachmentDesc,
			Remark:           req.Remark,
			EmployeeUuid:     "",
			CreateBy:         user.Uuid,
			CreatorName:      user.Username,
		})
		if err != nil {
			return err
		}
		return insertChildren(ctx, session, l.svcCtx, billUuid, req.Works, req.Educations, req.Families)
	}); txErr != nil {
		return nil, txErr
	}

	l.Infof("entry: 入职单已创建 billNo=%s name=%s hireDate=%s probation=%d月",
		billNo, req.Name, req.HireDate, req.ProbationMonths)
	return &types.AddResponse{}, nil
}
