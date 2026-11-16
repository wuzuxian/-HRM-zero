package salary_batch

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_batch_item"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_batch"
)

type GetBatchResult struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewGetBatchResult(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *GetBatchResult {
	return &GetBatchResult{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// GetBatchResult 查看结果：批次行 + 全员明细行（按计算顺序与工号稳定排序，前端按员工分组展示）。
func (l *GetBatchResult) GetBatchResult(req *types.GetBatchResultRequest) (resp *types.GetBatchResultResponse, err error) {
	batch, err := findBatchByUuid(l.ctx, l.svcCtx, req.BatchUuid)
	if err != nil {
		return nil, err
	}
	if batch.TrialStatus != "1" {
		return nil, statusNoResult("批次尚未试算，请先执行试算")
	}
	items, err := l.svcCtx.Model.HrmSalaryBatchItem.FindByCondition(l.ctx, nil,
		condition.NewChain().
			Equal(hrm_salary_batch_item.BatchUuid, req.BatchUuid).
			OrderByAsc(hrm_salary_batch_item.EmployeeNo).
			OrderByAsc(hrm_salary_batch_item.CalcOrder).
			Build()...)
	if err != nil {
		return nil, err
	}
	records := make([]types.HrmSalaryBatchItem, 0, len(items))
	for _, row := range items {
		records = append(records, batchItemRowToRecord(row))
	}
	return &types.GetBatchResultResponse{
		Batch: batchRowToRecord(batch),
		Items: records,
	}, nil
}

// batchItemRowToRecord 明细行 → API record.
func batchItemRowToRecord(row *hrm_salary_batch_item.HrmSalaryBatchItem) types.HrmSalaryBatchItem {
	return types.HrmSalaryBatchItem{
		Uuid:         row.Uuid,
		BatchUuid:    row.BatchUuid,
		EmployeeUuid: row.EmployeeUuid,
		EmployeeNo:   row.EmployeeNo,
		EmployeeName: row.EmployeeName,
		OrgName:      row.OrgName,
		ItemUuid:     row.ItemUuid,
		ItemCode:     row.ItemCode,
		ItemName:     row.ItemName,
		ItemClass:    row.ItemClass,
		ItemType:     row.ItemType,
		Amount:       row.Amount,
		Source:       row.Source,
		SourceDetail: row.SourceDetail,
		CalcOrder:    row.CalcOrder,
	}
}
