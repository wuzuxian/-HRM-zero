package scheduleboard

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero/core/stores/condition"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_schedule_entry"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/scheduleboard"
)

type Publish struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewPublish(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Publish {
	return &Publish{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *Publish) Publish(req *types.PublishRequest) (resp *types.PublishResponse, err error) {
	// 发布本月 / 撤销发布：当月该考勤组已有排班行整批标记。
	first, next, err := parseMonth(req.Month)
	if err != nil {
		return nil, err
	}
	_, members, err := groupAndMembers(l.ctx, l.svcCtx, req.GroupUuid)
	if err != nil {
		return nil, err
	}
	days := monthDays(first, next)

	target := publishedYes
	if req.Unpublish == "1" {
		target = publishedNo
	}

	var count int64
	if err = l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		rows, err := l.svcCtx.Model.HrmScheduleEntry.FindByCondition(ctx, session,
			condition.NewChain().
				GreaterEqualThan(hrm_schedule_entry.ScheduleDate, days[0]).
				LessEqualThan(hrm_schedule_entry.ScheduleDate, days[len(days)-1]).
				In(hrm_schedule_entry.EmployeeUuid, members).
				Build()...)
		if err != nil {
			return err
		}
		for _, row := range rows {
			if row.Published == target {
				continue
			}
			if err := l.svcCtx.Model.HrmScheduleEntry.UpdateFieldsByCondition(ctx, session,
				map[string]any{"published": target},
				condition.NewChain().Equal(hrm_schedule_entry.Uuid, row.Uuid).Build()...); err != nil {
				return err
			}
			count++
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &types.PublishResponse{PublishedCount: count}, nil
}
