package bill

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/bill"
)

type GetApprovalRecordList struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewGetApprovalRecordList(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *GetApprovalRecordList {
	return &GetApprovalRecordList{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

const recordTimeLayout = "2006-01-02 15:04:05"

func (l *GetApprovalRecordList) GetApprovalRecordList(req *types.GetApprovalRecordListRequest) (resp *types.GetApprovalRecordListResponse, err error) {
	records, err := Records(l.ctx, l.svcCtx.SqlxConn, req.BillType, req.BillUuid)
	if err != nil {
		return nil, err
	}

	resp = &types.GetApprovalRecordListResponse{
		Records: make([]*types.ApprovalRecord, 0, len(records)),
	}
	resp.Total = int64(len(records))
	for _, r := range records {
		resp.Records = append(resp.Records, &types.ApprovalRecord{
			BillNo:       r.BillNo,
			Action:       string(r.Action),
			FromStatus:   r.FromStatus,
			ToStatus:     r.ToStatus,
			OperatorName: r.OperatorName,
			Opinion:      r.Opinion,
			CreateTime:   r.CreateTime.Format(recordTimeLayout),
		})
	}
	return resp, nil
}
