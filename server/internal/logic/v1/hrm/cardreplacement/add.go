package cardreplacement

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	hrmuser "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/hrmuser"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_card_replacement_bill"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_punch_record"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/cardreplacement"
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

// timeNow 可测试钩子（单据编号日期部分用当天）。
var timeNow = time.Now

// Add 新增补卡单（HR206）：事务内生成单据编号 + 快照员工信息 + 快照当日打卡行的
// 原上班打卡/原下班打卡（hrm_punch_record，空=缺卡），初始状态 未提交。
// 提交/审批/撤销走共享单据底座 v1/hrm/bill（类型码 206）。
func (l *Add) Add(req *types.AddRequest) (resp *types.AddResponse, err error) {
	user := hrmuser.Current(l.ctx)
	if err := validateCardReplacement(req.CardType, req.PunchDate, req.PunchTime); err != nil {
		return nil, err
	}
	// 员工档案校验（补卡单必须挂在真实员工上）
	employee, err := l.svcCtx.Model.HrmEmployee.FindOneByCondition(l.ctx, nil,
		condition.NewChain().Equal(hrm_employee.Uuid, req.EmployeeUuid).Build()...)
	if err != nil {
		if err == hrm_employee.ErrNotFound {
			return nil, employeeNotFound(req.EmployeeUuid)
		}
		return nil, err
	}

	var billNo string
	if txErr := l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		no, err := bill.GenerateBillNo(ctx, session, BillTypeCode, timeNow())
		if err != nil {
			return err
		}
		billNo = no

		origIn, origOut := origClocksOf(ctx, session, l.svcCtx, req.EmployeeUuid, req.PunchDate)

		return l.svcCtx.Model.HrmCardReplacementBill.InsertV2(ctx, session, &hrm_card_replacement_bill.HrmCardReplacementBill{
			Uuid:             uuid.New().String(),
			BillNo:           no,
			Status:           bill.StatusUnsubmitted,
			EmployeeUuid:     employee.Uuid,
			EmployeeNo:       employee.EmployeeNo,
			EmployeeName:     employee.Name,
			OrgUuid:          req.OrgUuid,
			OrgName:          req.OrgName,
			PunchDate:        req.PunchDate,
			CardType:         req.CardType,
			PunchTime:        req.PunchTime,
			Reason:           req.Reason,
			OrigClockInTime:  origIn,
			OrigClockOutTime: origOut,
			CreateBy:         user.Uuid,
			CreatorName:      user.Username,
		})
	}); txErr != nil {
		return nil, txErr
	}

	l.Infof("cardreplacement: 补卡单已创建 billNo=%s employee=%s(%s) date=%s type=%s time=%s",
		billNo, employee.Name, employee.EmployeeNo, req.PunchDate, req.CardType, req.PunchTime)
	return &types.AddResponse{}, nil
}

// origClocksOf 当日打卡行快照（提交时点的 原上班打卡/原下班打卡，缺卡为空）。
func origClocksOf(ctx context.Context, session sqlx.Session, svcCtx *svc.ServiceContext, employeeUuid, date string) (string, string) {
	m := svcCtx.Model.HrmPunchRecord
	if session != nil {
		m = hrm_punch_record.NewHrmPunchRecordModel(sqlx.NewSqlConnFromSession(session))
	}
	rec, err := m.FindOneByCondition(ctx, session,
		condition.NewChain().
			Equal(hrm_punch_record.PunchDate, date).
			Equal(hrm_punch_record.EmployeeUuid, employeeUuid).
			Build()...)
	if err != nil || rec == nil {
		return "", ""
	}
	origIn, origOut := "", ""
	if rec.ClockInTime.Valid {
		origIn = rec.ClockInTime.Time.Format(time.DateTime)
	}
	if rec.ClockOutTime.Valid {
		origOut = rec.ClockOutTime.Time.Format(time.DateTime)
	}
	return origIn, origOut
}
