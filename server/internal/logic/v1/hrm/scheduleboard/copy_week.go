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
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_schedule_entry"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/scheduleboard"
)

type CopyWeek struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewCopyWeek(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *CopyWeek {
	return &CopyWeek{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *CopyWeek) CopyWeek(req *types.CopyWeekRequest) (resp *types.CopyWeekResponse, err error) {
	// 复制周：源周 7 天的排班按星期对齐复制到目标周。
	sourceStart, err := validateDate(req.SourceWeekStart)
	if err != nil {
		return nil, err
	}
	targetStart, err := validateDate(req.TargetWeekStart)
	if err != nil {
		return nil, err
	}
	if sourceStart.Weekday() != time.Monday {
		return nil, status.ErrorMessage(errcodes.ScheduleBoardInvalidCode, weekAlignWarn+": "+req.SourceWeekStart)
	}
	if targetStart.Weekday() != time.Monday {
		return nil, status.ErrorMessage(errcodes.ScheduleBoardInvalidCode, weekAlignWarn+": "+req.TargetWeekStart)
	}
	_, members, err := groupAndMembers(l.ctx, l.svcCtx, req.GroupUuid)
	if err != nil {
		return nil, err
	}
	if sourceStart.Equal(targetStart) {
		return nil, status.ErrorMessage(errcodes.ScheduleBoardInvalidCode, "源周与目标周不能相同")
	}

	// 读源周全部排班
	sourceEnd := sourceStart.AddDate(0, 0, 6).Format(dateLayout)
	sourceRows, err := l.svcCtx.Model.HrmScheduleEntry.FindByCondition(l.ctx, nil,
		condition.NewChain().
			GreaterEqualThan(hrm_schedule_entry.ScheduleDate, req.SourceWeekStart).
			LessEqualThan(hrm_schedule_entry.ScheduleDate, sourceEnd).
			In(hrm_schedule_entry.EmployeeUuid, members).
			Build()...)
	if err != nil {
		return nil, err
	}
	// 按星期偏移索引源班次（0=周一..6=周日）
	type cellKey struct{ emp, date string }
	sourceMap := make(map[cellKey]*hrm_schedule_entry.HrmScheduleEntry, len(sourceRows))
	for _, e := range sourceRows {
		sourceMap[cellKey{e.EmployeeUuid, e.ScheduleDate}] = e
	}
	// 目标周已有排班（overwrite=false 时仅填充空格；已发布日保护）
	targetEnd := targetStart.AddDate(0, 0, 6).Format(dateLayout)
	targetRows, err := l.svcCtx.Model.HrmScheduleEntry.FindByCondition(l.ctx, nil,
		condition.NewChain().
			GreaterEqualThan(hrm_schedule_entry.ScheduleDate, req.TargetWeekStart).
			LessEqualThan(hrm_schedule_entry.ScheduleDate, targetEnd).
			In(hrm_schedule_entry.EmployeeUuid, members).
			Build()...)
	if err != nil {
		return nil, err
	}
	targetMap := make(map[cellKey]*hrm_schedule_entry.HrmScheduleEntry, len(targetRows))
	for _, e := range targetRows {
		targetMap[cellKey{e.EmployeeUuid, e.ScheduleDate}] = e
	}

	overwrite := req.Overwrite == "1"
	var written int64
	if err = l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		for _, empUuid := range members {
			for offset := 0; offset < 7; offset++ {
				srcDate := sourceStart.AddDate(0, 0, offset).Format(dateLayout)
				src := sourceMap[cellKey{empUuid, srcDate}]
				if src == nil || src.ShiftUuid == "" {
					continue // 源周该日无班次（休），不复制
				}
				dstDate := targetStart.AddDate(0, 0, offset).Format(dateLayout)
				dst := targetMap[cellKey{empUuid, dstDate}]
				if dst != nil {
					if dst.Published == publishedYes {
						continue // 发布保护
					}
					if dst.ShiftUuid != "" && !overwrite {
						continue
					}
					if dst.ShiftUuid == src.ShiftUuid {
						continue
					}
				}
				if err := upsertEntry(ctx, l.svcCtx, session, empUuid, dstDate, src.ShiftUuid, req.GroupUuid, false); err != nil {
					return err
				}
				written++
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &types.CopyWeekResponse{WrittenCount: written}, nil
}
