package punch

// StatisticsDetail 考勤统计行操作「查看明细」（HRM-22）：某员工在日期区间内的
// 逐日 settle 详情弹窗 —— 复用打卡记录数据（toRecord），日期区间缺省当月。

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_punch_record"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/punch"
)

type StatisticsDetail struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewStatisticsDetail(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *StatisticsDetail {
	return &StatisticsDetail{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *StatisticsDetail) StatisticsDetail(req *types.StatisticsDetailRequest) (resp *types.StatisticsDetailResponse, err error) {
	if req.EmployeeUuid == "" {
		return nil, status.ErrorMessage(errcodes.PunchStatisticsInvalidCode, "员工必填")
	}
	startDate, endDate, rangeErr := resolveStatsRange(req.StartDate, req.EndDate)
	if rangeErr != nil {
		return nil, rangeErr
	}

	rows, err := l.svcCtx.Model.HrmPunchRecord.FindByCondition(l.ctx, nil,
		condition.NewChain().
			Equal(hrm_punch_record.EmployeeUuid, req.EmployeeUuid).
			GreaterEqualThan(hrm_punch_record.PunchDate, startDate).
			LessEqualThan(hrm_punch_record.PunchDate, endDate).
			OrderByAsc(hrm_punch_record.PunchDate).
			Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.HrmPunchRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, toRecord(l.ctx, l.svcCtx, row))
	}
	return &types.StatisticsDetailResponse{Records: records}, nil
}
