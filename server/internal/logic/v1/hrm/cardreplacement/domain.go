package cardreplacement

// 补卡单领域核心（HRM-20）：常量 + 校验 + 单据类型注册 + 审批通过写回。
// 端点 logic 见同目录 add.go / edit.go / delete.go / list.go。

import (
	"context"
	"time"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_card_replacement_bill"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_punch_record"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/punch"
)

// BillTypeCode is the 补卡单 类型码 (HR206, GLOSSARY「单据编号」/ demo-inventory §4.2).
const BillTypeCode = "206"

// 补卡类型（GLOSSARY「补卡」）：上班卡 / 下班卡；列值与打卡端点 punchAt 的类型一致.
const (
	CardTypeClockIn  = "clockIn"
	CardTypeClockOut = "clockOut"
)

// DateLayout / TimeLayout 统一时间字符串格式.
const (
	DateLayout = "2006-01-02"
	TimeLayout = "2006-01-02 15:04:05"
)

// validateCardReplacement Add/Edit 共用校验：补卡类型 ∈ {clockIn, clockOut}、
// 补卡日期可解析、补卡时间可解析且属于补卡日期当天.
func validateCardReplacement(cardType, punchDate, punchTime string) error {
	if cardType != CardTypeClockIn && cardType != CardTypeClockOut {
		return status.ErrorMessage(errcodes.CardReplacementInvalidCode,
			"补卡类型必须为 clockIn（上班卡）或 clockOut（下班卡）")
	}
	d, err := time.ParseInLocation(DateLayout, punchDate, time.Local)
	if err != nil {
		return status.ErrorMessage(errcodes.CardReplacementInvalidCode,
			"补卡日期必须为 yyyy-MM-dd 格式: "+punchDate)
	}
	t, err := time.ParseInLocation(TimeLayout, punchTime, time.Local)
	if err != nil {
		return status.ErrorMessage(errcodes.CardReplacementInvalidCode,
			"补卡时间必须为 yyyy-MM-dd HH:mm:ss 格式: "+punchTime)
	}
	if t.Format(DateLayout) != d.Format(DateLayout) {
		return status.ErrorMessage(errcodes.CardReplacementDateMismatchCode,
			"补卡时间必须落在补卡日期当天: "+punchTime+" ∉ "+punchDate)
	}
	return nil
}

// registerBillHandler registers the 补卡单 document type against the shared
// approval base (HRM-06). OnApproved（审批通过事务内）:
//  1. 补卡时间写入当日 hrm_punch_record 对应列（上班卡→clock_in_*，下班卡→clock_out_*；
//     无行则建行，source=card_replacement）——同一事务；
//  2. re-run punch.SettleDay 重算该员工该日日结（HRM-16 消费契约）。
//
// SettleDay 需要完整 svcCtx（规则解析走 attendancegroup.ResolveEffectiveRule），
// 而审批事务回调只带 (ctx, session, billUuid)，因此复用 leave 包的 svcCtx-key
// 约定：handler 入口通过 WithSvcCtx 注入，OnApproved 从 ctx 取回。
func registerBillHandler() {
	bill.RegisterType(BillTypeCode, bill.Handler{
		Load: func(ctx context.Context, session sqlx.Session, billUuid string) (string, string, error) {
			row, err := modelOf(session).FindOneByUuid(ctx, session, billUuid)
			if err != nil {
				return "", "", err
			}
			return row.BillNo, row.Status, nil
		},
		SetStatus: func(ctx context.Context, session sqlx.Session, billUuid, statusCode string) error {
			return modelOf(session).UpdateFieldsByCondition(ctx, session,
				map[string]any{"status": statusCode},
				condition.NewChain().Equal(hrm_card_replacement_bill.Uuid, billUuid).Build()...)
		},
		OnApproved: onApprovedApplyPunch,
	})
}

// modelOf 事务内构造 model（session==nil 用聚合 Model；与 leavebalance.modelOf 同款）.
func modelOf(session sqlx.Session) hrm_card_replacement_bill.HrmCardReplacementBillModel {
	if session == nil {
		return svcCtxOf(context.Background()).Model.HrmCardReplacementBill
	}
	return hrm_card_replacement_bill.NewHrmCardReplacementBillModel(sqlx.NewSqlConnFromSession(session))
}

// onApprovedApplyPunch 审批通过事务内的业务回写：补卡写打卡行 + 重算日结。
// 幂等：由单据状态机保证（审批通过是终态，ApproveBill 重复调用被 ErrInvalidTransition 拒绝）.
func onApprovedApplyPunch(ctx context.Context, session sqlx.Session, billUuid string) error {
	row, err := modelOf(session).FindOneByUuid(ctx, session, billUuid)
	if err != nil {
		return err
	}

	employee, err := svcCtxOf(ctx).Model.HrmEmployee.FindOneByCondition(ctx, session,
		condition.NewChain().Equal(hrm_employee.Uuid, row.EmployeeUuid).Build()...)
	if err != nil {
		if err == hrm_employee.ErrNotFound {
			return status.ErrorMessage(errcodes.CardReplacementEmployeeNotFoundCode,
				"员工不存在: "+row.EmployeeUuid)
		}
		return err
	}

	punchTime, err := time.ParseInLocation(TimeLayout, row.PunchTime, time.Local)
	if err != nil {
		return status.ErrorMessage(errcodes.CardReplacementInvalidCode,
			"补卡时间非法: "+row.PunchTime)
	}

	if err := applyPunchInTx(ctx, session, svcCtxOf(ctx), employee, row.CardType, row.PunchDate, punchTime); err != nil {
		return err
	}

	// re-run SettleDay（HRM-16 唯一日结入口）：修正后的打卡行重算日结快照
	// （缺卡→正常/迟到…）。传入 事务绑定 的 svcCtx（TxSvcCtx），SettleDay 的
	// 读写全部走审批事务 session —— 与打卡写入同事务原子提交（sqlite 下另开
	// 连接会 SQLITE_BUSY 死锁）。
	_, err = punch.SettleDay(ctx, TxBoundSvcCtx(ctx, session, svcCtxOf(ctx)), row.EmployeeUuid, row.PunchDate)
	return err
}

// applyPunchInTx 把补卡时间写入当日打卡行对应列（无行则建行）。
// 与 punch.punchAt 语义一致但走事务 session + 不做重复卡拒绝（补卡单语义 = 明确修正）.
func applyPunchInTx(ctx context.Context, session sqlx.Session, svcCtx *svc.ServiceContext,
	employee *hrm_employee.HrmEmployee, cardType, punchDate string, punchTime time.Time) error {
	m := svcCtx.Model.HrmPunchRecord
	if session != nil {
		m = hrm_punch_record.NewHrmPunchRecordModel(sqlx.NewSqlConnFromSession(session))
	}

	rec, err := findPunchRowInTx(ctx, session, m, svcCtx, employee.Uuid, punchDate)
	if err != nil {
		return err
	}
	if rec == nil {
		rec = &hrm_punch_record.HrmPunchRecord{
			Uuid:         uuidNew(),
			PunchDate:    punchDate,
			EmployeeUuid: employee.Uuid,
			UserUuid:     employee.UserUuid,
			Username:     usernameOfTx(ctx, session, svcCtx, employee.UserUuid),
			EmployeeNo:   employee.EmployeeNo,
			EmployeeName: employee.Name,
			Source:       punch.SourceCardReplacement,
		}
		if _, err := m.Insert(ctx, session, rec); err != nil {
			return err
		}
	}

	fields := map[string]any{
		// 行来源标记为补卡（无论行是新建还是已有行被修正，都指向最近一次写入）
		"source": punch.SourceCardReplacement,
	}
	switch cardType {
	case CardTypeClockIn:
		fields["clock_in_time"] = punchTime
		fields["clock_in_status"] = punch.StatusNormal
		fields["clock_in_location"] = ""
	case CardTypeClockOut:
		fields["clock_out_time"] = punchTime
		fields["clock_out_status"] = punch.StatusNormal
		fields["clock_out_location"] = ""
	}
	return m.UpdateFieldsByCondition(ctx, session, fields,
		condition.NewChain().Equal(hrm_punch_record.Uuid, rec.Uuid).Build()...)
}

// findPunchRowInTx 当日打卡行（一人一天一行），无行返回 nil.
func findPunchRowInTx(ctx context.Context, session sqlx.Session, m hrm_punch_record.HrmPunchRecordModel,
	svcCtx *svc.ServiceContext, employeeUuid, date string) (*hrm_punch_record.HrmPunchRecord, error) {
	row, err := m.FindOneByCondition(ctx, session,
		condition.NewChain().
			Equal(hrm_punch_record.PunchDate, date).
			Equal(hrm_punch_record.EmployeeUuid, employeeUuid).
			Build()...)
	if err != nil {
		if err == hrm_punch_record.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return row, nil
}

// usernameOfTx 登录账号快照（user_uuid 无对应账号时返回空）.
func usernameOfTx(ctx context.Context, session sqlx.Session, svcCtx *svc.ServiceContext, userUuid string) string {
	if userUuid == "" {
		return ""
	}
	row, err := svcCtx.Model.ManageUser.FindOneByUuid(ctx, session, userUuid)
	if err != nil || row == nil {
		return ""
	}
	return row.Username
}

