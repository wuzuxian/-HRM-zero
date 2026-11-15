package punch

// Edit 管理端修改打卡记录（demo-inventory 4.4 行操作「修改」）：更新上/下班卡
// 时间/状态/地点/加班，保存后即时日结刷新 settle_* 快照。

import (
	"context"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_punch_record"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/punch"
)

type Edit struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewEdit(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Edit {
	return &Edit{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *Edit) Edit(req *types.EditRequest) (resp *types.EditResponse, err error) {
	if req.Uuid == "" {
		return nil, status.ErrorMessage(errcodes.PunchNotFoundCode, "打卡记录不能为空")
	}
	if err := validatePunchTimes(req.ClockInTime, req.ClockOutTime); err != nil {
		return nil, err
	}

	row, err := l.svcCtx.Model.HrmPunchRecord.FindOneByCondition(l.ctx, nil,
		condition.NewChain().Equal(hrm_punch_record.Uuid, req.Uuid).Build()...)
	if err != nil {
		if err == hrm_punch_record.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.PunchNotFoundCode, "打卡记录不存在")
		}
		return nil, err
	}

	// 请求字段可选: 未传（时间与状态均为空）= 该侧不改动；传时间未传状态 → 状态跟随 normal；
	// 空时间 + 显式空串状态语义不支持（可选字段解析后无法区分，清卡请删除整行重录）。
	fields := map[string]any{}
	if req.ClockInTime != "" || req.ClockInStatus != "" {
		if req.ClockInTime == "" {
			return nil, status.ErrorMessage(errcodes.PunchInvalidCode, "签到状态与签到时间必须同时填写")
		}
		tm, parseErr := time.ParseInLocation(time.DateTime, req.ClockInTime, time.Local)
		if parseErr != nil {
			return nil, status.ErrorMessage(errcodes.PunchInvalidCode, "签到时间必须为 yyyy-MM-dd HH:mm:ss")
		}
		fields["clock_in_time"] = tm
		inStatus := req.ClockInStatus
		if inStatus == "" {
			inStatus = StatusNormal
		}
		fields["clock_in_status"] = inStatus
		fields["clock_in_location"] = req.ClockInLocation
	}
	if req.ClockOutTime != "" || req.ClockOutStatus != "" {
		if req.ClockOutTime == "" {
			return nil, status.ErrorMessage(errcodes.PunchInvalidCode, "签退状态与签退时间必须同时填写")
		}
		tm, parseErr := time.ParseInLocation(time.DateTime, req.ClockOutTime, time.Local)
		if parseErr != nil {
			return nil, status.ErrorMessage(errcodes.PunchInvalidCode, "签退时间必须为 yyyy-MM-dd HH:mm:ss")
		}
		fields["clock_out_time"] = tm
		outStatus := req.ClockOutStatus
		if outStatus == "" {
			outStatus = StatusNormal
		}
		fields["clock_out_status"] = outStatus
		fields["clock_out_location"] = req.ClockOutLocation
	}
	if req.Overtime != "" {
		fields["overtime"] = normalizeOvertime(req.Overtime)
	}
	if len(fields) == 0 {
		return nil, status.ErrorMessage(errcodes.PunchInvalidCode, "没有需要修改的字段")
	}

	if err := l.svcCtx.Model.HrmPunchRecord.UpdateFieldsByCondition(l.ctx, nil, fields,
		condition.NewChain().Equal(hrm_punch_record.Uuid, row.Uuid).Build()...); err != nil {
		return nil, err
	}

	// 修改后即时日结
	if _, err := SettleDay(l.ctx, l.svcCtx, row.EmployeeUuid, row.PunchDate); err != nil {
		return nil, err
	}
	return &types.EditResponse{}, nil
}
