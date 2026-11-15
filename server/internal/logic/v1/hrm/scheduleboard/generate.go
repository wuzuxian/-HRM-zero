package scheduleboard

import (
	"context"
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_schedule_cycle"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_schedule_entry"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/scheduleboard"
)

type Generate struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewGenerate(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Generate {
	return &Generate{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *Generate) Generate(req *types.GenerateRequest) (resp *types.GenerateResponse, err error) {
	// 周期生成：选排班周期 → ExpandCycle 从锚点日展开落当月（HRM-14 消费契约）。
	// 相位锚定锚点日（缺省=当月 1 日）；跨月保持绝对相位需传同一历史锚点日。
	first, next, err := parseMonth(req.Month)
	if err != nil {
		return nil, err
	}
	_, members, err := groupAndMembers(l.ctx, l.svcCtx, req.GroupUuid)
	if err != nil {
		return nil, err
	}

	cycle, err := l.svcCtx.Model.HrmScheduleCycle.FindOneByUuid(l.ctx, nil, req.CycleUuid)
	if err != nil {
		if err == hrm_schedule_cycle.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.ScheduleBoardCycleNotFoundCode, "排班周期不存在")
		}
		return nil, err
	}
	if cycle.Status != statusEnabled {
		return nil, status.ErrorMessage(errcodes.ScheduleBoardCycleNotFoundCode, "排班周期已停用: "+cycle.CycleName)
	}
	seq := splitCsv(cycle.ShiftUuids)
	if len(seq) == 0 {
		return nil, status.ErrorMessage(errcodes.ScheduleBoardCycleNotFoundCode, "排班周期班次序列为空")
	}
	if err = checkShifts(l.ctx, l.svcCtx, seq); err != nil {
		return nil, err
	}

	anchor := first
	if req.AnchorDate != "" {
		anchor, err = validateDate(req.AnchorDate)
		if err != nil {
			return nil, err
		}
	}

	// ExpandCycle 签名是 []int64（HRM-14 工单指定）—— uuid 序列场景改为本地逐日相位展开，
	// 相位语义与 ExpandCycle 完全一致：锚点日 = 序列第 0 项，逐日 +1 取模，月界不重置相位。
	days := monthDays(first, next)
	anchorKey := anchor.Format(dateLayout)
	startOffset := 0
	for i, day := range days {
		if day == anchorKey {
			startOffset = i
			break
		}
	}
	// 锚点日早于月初时（历史锚点），相位差 = 天数差 % len(seq)
	if anchor.Before(first) {
		startOffset = int(first.Sub(time.Date(anchor.Year(), anchor.Month(), anchor.Day(), 0, 0, 0, 0, time.Local)).Hours()/24) % len(seq)
	}

	overwrite := req.Overwrite == "1"
	// 已发布行永不覆盖（发布保护）
	memberSet := make(map[string]bool, len(members))
	for _, u := range members {
		memberSet[u] = true
	}

	existing, err := l.svcCtx.Model.HrmScheduleEntry.FindByCondition(l.ctx, nil,
		condition.NewChain().
			GreaterEqualThan(hrm_schedule_entry.ScheduleDate, days[0]).
			LessEqualThan(hrm_schedule_entry.ScheduleDate, days[len(days)-1]).
			In(hrm_schedule_entry.EmployeeUuid, members).
			Build()...)
	if err != nil {
		return nil, err
	}
	type cellKey struct{ emp, date string }
	existingMap := make(map[cellKey]*hrm_schedule_entry.HrmScheduleEntry, len(existing))
	for _, e := range existing {
		existingMap[cellKey{e.EmployeeUuid, e.ScheduleDate}] = e
	}

	var written int64
	if err = l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		for _, empUuid := range members {
			if !memberSet[empUuid] {
				continue
			}
			for i, day := range days {
				shiftUuid := seq[(i+startOffset)%len(seq)]
				e := existingMap[cellKey{empUuid, day}]
				if e != nil {
					if e.Published == publishedYes {
						continue // 发布保护：已发布日跳过
					}
					if e.ShiftUuid != "" && !overwrite {
						continue // 仅填充空单元格
					}
					if e.ShiftUuid == shiftUuid {
						continue // 同值幂等
					}
				}
				if err := upsertEntry(ctx, l.svcCtx, session, empUuid, day, shiftUuid, req.GroupUuid, false); err != nil {
					return err
				}
				written++
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &types.GenerateResponse{WrittenCount: written}, nil
}
