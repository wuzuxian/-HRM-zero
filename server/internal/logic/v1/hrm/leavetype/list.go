package leavetype

import (
	"context"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_leave_type"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/leavetype"
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
		svcCtx: svcCtx, r: r,
	}
}

// List 假期类型分页列表，按排序号升序展示，支持按类型名称/状态筛选.
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	rows, total, err := l.svcCtx.Model.HrmLeaveType.PageByCondition(l.ctx, nil,
		condition.NewChain().
			Page(req.Current, req.Size).
			Like(hrm_leave_type.TypeName, "%"+req.TypeName+"%", condition.WithSkip(req.TypeName == "")).
			Equal(hrm_leave_type.Status, req.Status, condition.WithSkip(req.Status == "")).
			OrderByAsc(hrm_leave_type.OrderNum).
			OrderByAsc(hrm_leave_type.Id).
			Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.HrmLeaveType, 0, len(rows))
	for _, row := range rows {
		records = append(records, types.HrmLeaveType{
			Uuid:               row.Uuid,
			TypeCode:           row.TypeCode,
			TypeName:           row.TypeName,
			TypeValue:          row.TypeValue,
			BalanceEnabled:     row.BalanceEnabled,
			AttachmentRequired: row.AttachmentRequired,
			DefaultDays:        row.DefaultDays,
			OrderNum:           row.OrderNum,
			Status:             row.Status,
			Remark:             row.Remark,
			CreateTime:         row.CreateTime.Format(time.DateTime),
			UpdateTime:         row.UpdateTime.Format(time.DateTime),
		})
	}

	return &types.ListResponse{PageResponse: types.PageResponse{
		Current: req.Current,
		Size:    req.Size,
		Total:   total,
	}, Records: records}, nil
}
