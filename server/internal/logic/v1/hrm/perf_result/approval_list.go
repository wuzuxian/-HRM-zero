package perf_result

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	model "github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_result_approval"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_result"
)

type ApprovalList struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewApprovalList(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *ApprovalList {
	return &ApprovalList{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// ApprovalList 批次结果审批列表（HR211 单据；filters: periodUuid/status）。
func (l *ApprovalList) ApprovalList(req *types.ApprovalListRequest) (resp *types.ApprovalListResponse, err error) {
	chain := condition.NewChain().OrderByDesc(model.Id)
	if req.PeriodUuid != "" {
		chain = chain.Equal(model.PeriodUuid, req.PeriodUuid)
	}
	if req.Status != "" {
		chain = chain.Equal(model.Status, req.Status)
	}
	rows, total, err := model.NewHrmPerfResultApprovalModel(l.svcCtx.SqlxConn).PageByCondition(l.ctx, nil,
		chain.Page(req.Current, req.Size).Build()...)
	if err != nil {
		return nil, err
	}
	records := make([]types.HrmPerfResultApproval, 0, len(rows))
	for _, row := range rows {
		records = append(records, approvalRecord(row))
	}
	return &types.ApprovalListResponse{
		PageResponse: types.PageResponse{
			Current: req.Current,
			Size:    req.Size,
			Total:   total,
		},
		Records: records,
	}, nil
}

// approvalRecord maps the model row to the API record.
func approvalRecord(row *model.HrmPerfResultApproval) types.HrmPerfResultApproval {
	return types.HrmPerfResultApproval{
		Uuid:             row.Uuid,
		BillNo:           row.BillNo,
		PeriodUuid:       row.PeriodUuid,
		PeriodNo:         row.PeriodNo,
		PeriodName:       row.PeriodName,
		CompanySnapshot:  row.CompanySnapshot,
		OrgSnapshot:      row.OrgSnapshot,
		ParticipantCount: row.ParticipantCount,
		AvgScore:         row.AvgScore,
		TotalAmount:      row.TotalAmount,
		Status:           row.Status,
		SubmitBy:         row.SubmitBy,
		SubmitTime:       row.SubmitTime,
		ApproveBy:        row.ApproveBy,
		ApproveTime:      row.ApproveTime,
		Opinion:          row.Opinion,
		CreateTime:       row.CreateTime.Format(timeFormat),
		UpdateTime:       row.UpdateTime.Format(timeFormat),
	}
}
