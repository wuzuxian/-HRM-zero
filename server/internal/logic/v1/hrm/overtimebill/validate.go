package overtimebill

import (
	"context"
	"errors"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/overtimerule"
	model "github.com/jzero-io/jzero-admin/server/internal/model/hrm_overtime_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
)

// canEdit 与共享单据状态机一致的「可编辑」状态集（未提交/审批不通过/已撤销）.
func canEdit(statusCode string) bool {
	switch statusCode {
	case StatusUnsubmitted, StatusRejected, StatusRevoked:
		return true
	}
	return false
}

// conditionChainByUuid 单据 uuid 条件.
func conditionChainByUuid(billUuid string) []condition.Condition {
	return condition.NewChain().Equal(model.Uuid, billUuid).Build()
}

// validateOvertime Add/Edit 共用校验：当天类型合法 + 起止时间可计算；返回自动计算的申请时长（小时）.
func validateOvertime(dayType, startTime, endTime string) (float64, error) {
	switch dayType {
	case overtimerule.DayTypeWorkday, overtimerule.DayTypeRestday, overtimerule.DayTypeHoliday:
	default:
		return 0, status.ErrorMessage(errcodes.OvertimeInvalidCode,
			"加班当天类型必须是 工作日/休息日/法定节假日")
	}
	hours, err := CalcOvertimeHours(startTime, endTime, 0)
	if err != nil {
		if errors.Is(err, ErrInvalidTimeFormat) {
			return 0, status.ErrorMessage(errcodes.OvertimeInvalidCode,
				"起止时间格式非法，须为 YYYY-MM-DD HH:MM:SS")
		}
		if errors.Is(err, ErrInvalidTimeRange) {
			return 0, status.ErrorMessage(errcodes.OvertimeInvalidCode, "结束时间必须晚于开始时间")
		}
		return 0, status.ErrorMessage(errcodes.OvertimeInvalidCode, "起止时间非法")
	}
	if hours <= 0 {
		return 0, status.ErrorMessage(errcodes.OvertimeInvalidCode,
			"加班时长不足一个最小计算单位（30 分钟）")
	}
	return hours, nil
}

// findOneByUuid 事务内按 uuid 取单据；不存在映射为 10570.
func findOneByUuid(ctx context.Context, session sqlx.Session, svcCtx *svc.ServiceContext, billUuid string) (*model.HrmOvertimeBill, error) {
	row, err := modelOf(svcCtx, session).FindOneByUuid(ctx, session, billUuid)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.OvertimeNotFoundCode, "加班单不存在")
		}
		return nil, err
	}
	return row, nil
}

// modelOf 事务内构造 model（sqlx.NewSqlConnFromSession 与 leavebalance 同款）.
func modelOf(svcCtx *svc.ServiceContext, session sqlx.Session) model.HrmOvertimeBillModel {
	if session == nil {
		return svcCtx.Model.HrmOvertimeBill
	}
	return model.NewHrmOvertimeBillModel(sqlx.NewSqlConnFromSession(session))
}
