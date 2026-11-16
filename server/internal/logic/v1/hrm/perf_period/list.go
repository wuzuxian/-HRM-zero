package perf_period

import (
	"context"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_period"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_scheme"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_period"
)

type List struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewList(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *List {
	return &List{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// List 工作台列表页：状态列 + 参与人数快照 + 绩效方案名称快照（批量 IN 查询，无 N+1）。
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	rows, total, err := l.svcCtx.Model.HrmPerfPeriod.PageByCondition(l.ctx, nil, condition.NewChain().
		Page(req.Current, req.Size).
		OrderByDesc(hrm_perf_period.CreateTime).
		Like(hrm_perf_period.Name, "%"+req.Name+"%", condition.WithSkip(req.Name == "")).
		Equal(hrm_perf_period.Status, req.Status, condition.WithSkip(req.Status == "")).
		Build()...)
	if err != nil {
		return nil, err
	}

	// 绩效方案名称快照（一次批量查询）
	schemeUuids := make([]string, 0, len(rows))
	for _, row := range rows {
		schemeUuids = append(schemeUuids, row.SchemeUuid)
	}
	schemeNames := make(map[string]string, len(schemeUuids))
	if len(schemeUuids) > 0 {
		schemes, err := l.svcCtx.Model.HrmPerfScheme.FindByCondition(l.ctx, nil,
			condition.NewChain().In(hrm_perf_scheme.Uuid, schemeUuids).Build()...)
		if err != nil {
			return nil, err
		}
		for _, s := range schemes {
			schemeNames[s.Uuid] = s.Name
		}
	}

	// 参与人数快照（一次 GROUP BY 聚合查询）
	counts := make(map[string]int64, len(rows))
	if len(rows) > 0 {
		type countRow struct {
			PeriodUuid string `db:"period_uuid"`
			Cnt        int64  `db:"cnt"`
		}
		var countRows []countRow
		query := "SELECT period_uuid, COUNT(*) AS cnt FROM hrm_perf_period_participant WHERE period_uuid IN (" +
			placeholders(len(rows)) + ") GROUP BY period_uuid"
		args := make([]any, 0, len(rows))
		for _, row := range rows {
			args = append(args, row.Uuid)
		}
		if err := l.svcCtx.SqlxConn.QueryRowsCtx(l.ctx, &countRows, query, args...); err != nil {
			return nil, err
		}
		for _, c := range countRows {
			counts[c.PeriodUuid] = c.Cnt
		}
	}

	records := make([]types.HrmPerfPeriod, 0, len(rows))
	for _, row := range rows {
		records = append(records, periodRecord(row, schemeNames[row.SchemeUuid], counts[row.Uuid]))
	}
	return &types.ListResponse{
		PageResponse: types.PageResponse{
			Current: req.Current,
			Size:    req.Size,
			Total:   total,
		},
		Records: records,
	}, nil
}

// placeholders renders "?,?,?" (dialect-agnostic — both sqlite and mysql accept ?).
func placeholders(n int) string {
	out := ""
	for i := 0; i < n; i++ {
		if i > 0 {
			out += ","
		}
		out += "?"
	}
	return out
}

// periodRecord renders one 绩效批次 row for the API (schemeName = 绩效方案名称快照).
func periodRecord(row *hrm_perf_period.HrmPerfPeriod, schemeName string, participantCount int64) types.HrmPerfPeriod {
	return types.HrmPerfPeriod{
		Uuid:              row.Uuid,
		PeriodNo:          row.PeriodNo,
		Name:              row.Name,
		PeriodType:        row.PeriodType,
		SchemeUuid:        row.SchemeUuid,
		SchemeName:        schemeName,
		Scope:             row.Scope,
		PeriodStart:       row.PeriodStart,
		PeriodEnd:         row.PeriodEnd,
		ResultPublishDate: row.ResultPublishDate,
		ParticipantCount:  participantCount,
		Status:            row.Status,
		Remark:            row.Remark,
		CreateTime:        row.CreateTime.Format(time.DateTime),
		UpdateTime:        row.UpdateTime.Format(time.DateTime),
	}
}
