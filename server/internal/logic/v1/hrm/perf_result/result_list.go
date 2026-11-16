package perf_result

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_result"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_result"
)

type ResultList struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewResultList(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *ResultList {
	return &ResultList{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// ResultList 绩效结果列表（demo-inventory 9.3 绩效结果页；filters: periodUuid/
// employeeUuid/signed；mine=仅本人 — JWT 登录账号 → hrm_employee.user_uuid 反查）。
func (l *ResultList) ResultList(req *types.ResultListRequest) (resp *types.ResultListResponse, err error) {
	chain := condition.NewChain().OrderByDesc(hrm_perf_result.Id)
	if req.PeriodUuid != "" {
		chain = chain.Equal(hrm_perf_result.PeriodUuid, req.PeriodUuid)
	}
	if req.EmployeeUuid != "" {
		chain = chain.Equal(hrm_perf_result.EmployeeUuid, req.EmployeeUuid)
	}
	if req.Signed != "" {
		chain = chain.Equal(hrm_perf_result.Signed, req.Signed)
	}
	if req.CompanyUuid != "" {
		chain = chain.Equal(hrm_perf_result.CompanyUuid, req.CompanyUuid)
	}
	if req.OrgUuid != "" {
		chain = chain.Equal(hrm_perf_result.OrgUuid, req.OrgUuid)
	}
	if req.SalarySynced != "" {
		chain = chain.Equal(hrm_perf_result.SalarySynced, req.SalarySynced)
	}
	if req.Mine {
		employeeUuid, rerr := actorEmployeeUuid(l.ctx, l.svcCtx.SqlxConn)
		if rerr != nil {
			return nil, rerr
		}
		chain = chain.Equal(hrm_perf_result.EmployeeUuid, employeeUuid)
	}
	rows, total, err := hrm_perf_result.NewHrmPerfResultModel(l.svcCtx.SqlxConn).PageByCondition(l.ctx, nil,
		chain.Page(req.Current, req.Size).Build()...)
	if err != nil {
		return nil, err
	}
	records := make([]types.HrmPerfResult, 0, len(rows))
	for _, row := range rows {
		records = append(records, resultRecord(row))
	}
	return &types.ResultListResponse{
		PageResponse: types.PageResponse{
			Current: req.Current,
			Size:    req.Size,
			Total:   total,
		},
		Records: records,
	}, nil
}

// resultRecord maps the model row to the API record.
func resultRecord(row *hrm_perf_result.HrmPerfResult) types.HrmPerfResult {
	return types.HrmPerfResult{
		Uuid:           row.Uuid,
		ApprovalUuid:   row.ApprovalUuid,
		AssessmentUuid: row.AssessmentUuid,
		PeriodUuid:     row.PeriodUuid,
		PeriodNo:       row.PeriodNo,
		EmployeeUuid:   row.EmployeeUuid,
		EmployeeName:   row.EmployeeName,
		CompanyUuid:    row.CompanyUuid,
		CompanyName:    row.CompanyName,
		OrgUuid:        row.OrgUuid,
		OrgName:        row.OrgName,
		Score:          row.Score,
		Grade:          row.Grade,
		Coefficient:    row.Coefficient,
		PerfAmount:     row.PerfAmount,
		BonusAmount:    row.BonusAmount,
		ResultStatus:   row.ResultStatus,
		Published:      row.Published,
		Signed:         row.Signed,
		SignTime:       row.SignTime,
		SignBy:         row.SignBy,
		SalarySynced:   row.SalarySynced,
		CreateTime:     row.CreateTime.Format(timeFormat),
		UpdateTime:     row.UpdateTime.Format(timeFormat),
	}
}
