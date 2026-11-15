package salary_item

import (
	"context"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_item"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_item"
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

// List 薪资项目分页列表: 默认按计算顺序升序（同序按 id），试算引擎（HRM-37）按此顺序消费；
// 支持按项目名称/项目分类/状态筛选.
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	rows, total, err := l.svcCtx.Model.HrmSalaryItem.PageByCondition(l.ctx, nil,
		condition.NewChain().
			Page(req.Current, req.Size).
			Like(hrm_salary_item.ItemName, "%"+req.ItemName+"%", condition.WithSkip(req.ItemName == "")).
			Equal(hrm_salary_item.ItemClass, req.ItemClass, condition.WithSkip(req.ItemClass == "")).
			Equal(hrm_salary_item.Status, req.Status, condition.WithSkip(req.Status == "")).
			OrderByAsc(hrm_salary_item.CalcOrder).
			OrderByAsc(hrm_salary_item.Id).
			Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.HrmSalaryItem, 0, len(rows))
	for _, row := range rows {
		records = append(records, types.HrmSalaryItem{
			Uuid:       row.Uuid,
			ItemCode:   row.ItemCode,
			ItemName:   row.ItemName,
			ItemType:   row.ItemType,
			ItemClass:  row.ItemClass,
			ValueMode:  row.ValueMode,
			CalcOrder:  row.CalcOrder,
			OrderNum:   row.OrderNum,
			Formula:    row.Formula,
			IsPreset:   row.IsPreset,
			Status:     row.Status,
			Remark:     row.Remark,
			CreateTime: row.CreateTime.Format(time.DateTime),
			UpdateTime: row.UpdateTime.Format(time.DateTime),
		})
	}

	return &types.ListResponse{PageResponse: types.PageResponse{
		Current: req.Current,
		Size:    req.Size,
		Total:   total,
	}, Records: records}, nil
}
