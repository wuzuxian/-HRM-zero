package schedulecycle

import (
	"context"

	"github.com/jzero-io/jzero/core/status"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_attendance_group"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_shift"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
)

// cycleFields Add/Edit 共用的校验输入（结构一致, 仅 Status 语义略有差别）.
type cycleFields struct {
	CycleName  string
	ShiftUuids []string
	GroupUuids []string
	Status     string
}

func cycleFieldsFrom(cycleName string, shiftUuids, groupUuids []string, cycleStatus string) *cycleFields {
	return &cycleFields{
		CycleName:  cycleName,
		ShiftUuids: shiftUuids,
		GroupUuids: groupUuids,
		Status:     cycleStatus,
	}
}

// validateCycleBasic 排班周期字段校验（Add/Edit 共用）.
func validateCycleBasic(svcCtx *svc.ServiceContext, f *cycleFields) error {
	if f.CycleName == "" {
		return status.ErrorMessage(errcodes.ScheduleCycleInvalidCode, "周期名称不能为空")
	}
	if err := validateShiftSequence(svcCtx, f.ShiftUuids); err != nil {
		return err
	}
	if err := validateGroupBindings(svcCtx, f.GroupUuids); err != nil {
		return err
	}
	if f.Status != statusEnabled && f.Status != statusDisabled {
		return status.ErrorMessage(errcodes.ScheduleCycleInvalidCode, "排班周期状态必须是 启用 或 停用")
	}
	return nil
}

// validateShiftSequence 班次序列校验: 非空、序列中的每个班次 id 必须存在且启用（GLOSSARY「排班周期」）.
func validateShiftSequence(svcCtx *svc.ServiceContext, shiftUuids []string) error {
	if len(shiftUuids) == 0 {
		return status.ErrorMessage(errcodes.ScheduleCycleInvalidCode, "班次序列不能为空")
	}
	for _, su := range shiftUuids {
		if su == "" {
			return status.ErrorMessage(errcodes.ScheduleCycleShiftInvalidCode, "班次序列中存在空班次引用")
		}
		shift, err := svcCtx.Model.HrmShift.FindOneByUuid(context.Background(), nil, su)
		if err != nil {
			if err == hrm_shift.ErrNotFound {
				return status.ErrorMessage(errcodes.ScheduleCycleShiftInvalidCode, "班次序列中的班次不存在")
			}
			return err
		}
		if shift.Status != statusEnabled {
			return status.ErrorMessage(errcodes.ScheduleCycleShiftInvalidCode, "班次序列中的班次已停用: "+shift.ShiftName)
		}
	}
	return nil
}

// validateGroupBindings 适用考勤组必须真实存在.
func validateGroupBindings(svcCtx *svc.ServiceContext, groupUuids []string) error {
	for _, g := range groupUuids {
		if g == "" {
			return status.ErrorMessage(errcodes.ScheduleCycleGroupInvalidCode, "适用考勤组中存在空引用")
		}
		_, err := svcCtx.Model.HrmAttendanceGroup.FindOneByUuid(context.Background(), nil, g)
		if err != nil {
			if err == hrm_attendance_group.ErrNotFound {
				return status.ErrorMessage(errcodes.ScheduleCycleGroupInvalidCode, "适用考勤组不存在")
			}
			return err
		}
	}
	return nil
}
