package salary_batch

import (
	"context"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_batch"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_batch"
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

// List 批次分页列表（批次名称模糊/主体/批次类型/年月/试算状态筛选；年月升序=时间线）。
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	chain := condition.NewChain().
		Like(hrm_salary_batch.BatchName, "%"+req.BatchName+"%", condition.WithSkip(req.BatchName == "")).
		Equal(hrm_salary_batch.SubjectUuid, req.SubjectUuid, condition.WithSkip(req.SubjectUuid == "")).
		Equal(hrm_salary_batch.BatchType, req.BatchType, condition.WithSkip(req.BatchType == "")).
		Equal(hrm_salary_batch.BatchYear, req.Year, condition.WithSkip(req.Year == 0)).
		Equal(hrm_salary_batch.BatchMonth, req.Month, condition.WithSkip(req.Month == 0)).
		Equal(hrm_salary_batch.TrialStatus, req.TrialStatus, condition.WithSkip(req.TrialStatus == "")).
		OrderByAsc(hrm_salary_batch.BatchYear).
		OrderByAsc(hrm_salary_batch.BatchMonth).
		OrderByAsc(hrm_salary_batch.Id).
		Page(int(req.Current), int(req.Size))

	rows, total, err := l.svcCtx.Model.HrmSalaryBatch.PageByCondition(l.ctx, nil, chain.Build()...)
	if err != nil {
		return nil, err
	}
	records := make([]types.HrmSalaryBatch, 0, len(rows))
	for _, row := range rows {
		records = append(records, batchRowToRecord(row))
	}
	return &types.ListResponse{PageResponse: types.PageResponse{
		Current: req.Current,
		Size:    req.Size,
		Total:   total,
	}, Records: records}, nil
}

// batchRowToRecord 批次行 → API record（trial_time NULL=未试算空串）.
func batchRowToRecord(row *hrm_salary_batch.HrmSalaryBatch) types.HrmSalaryBatch {
	trialTime := ""
	if row.TrialTime.Valid {
		trialTime = row.TrialTime.Time.Format(time.DateTime)
	}
	warnings := ""
	if row.Warnings.Valid {
		warnings = row.Warnings.String
	}
	return types.HrmSalaryBatch{
		Uuid:           row.Uuid,
		BatchNo:        row.BatchNo,
		BatchName:      row.BatchName,
		BatchType:      row.BatchType,
		SubjectUuid:    row.SubjectUuid,
		SubjectName:    row.SubjectName,
		BatchYear:      row.BatchYear,
		BatchMonth:     row.BatchMonth,
		Headcount:      row.Headcount,
		EarningTotal:   row.EarningTotal,
		DeductionTotal: row.DeductionTotal,
		NetTotal:       row.NetTotal,
		TrialStatus:    row.TrialStatus,
		ConfirmStatus:  row.ConfirmStatus,
		PayStatus:      row.PayStatus,
		PayslipStatus:  row.PayslipStatus,
		TrialTime:      trialTime,
		Version:        row.Version,
		Warnings:       warnings,
		Remark:         row.Remark,
		CreateTime:     row.CreateTime.Format(time.DateTime),
		UpdateTime:     row.UpdateTime.Format(time.DateTime),
	}
}
