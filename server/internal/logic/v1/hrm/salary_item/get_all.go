package salary_item

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_item"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_item"
)

type GetAll struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewGetAll(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *GetAll {
	return &GetAll{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// GetAll 全量薪资项目（不分页, calc_order ASC, id ASC —— 薪酬方案配置弹窗全量勾选用）。
func (l *GetAll) GetAll(req *types.GetAllResponse) (resp *types.GetAllResponse, err error) {
	rows, err := l.svcCtx.Model.HrmSalaryItem.FindByCondition(l.ctx, nil,
		condition.NewChain().
			OrderByAsc(hrm_salary_item.CalcOrder).
			OrderByAsc(hrm_salary_item.Id).
			Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.HrmSalaryItem, 0, len(rows))
	for _, row := range rows {
		records = append(records, itemRowToRecord(row))
	}

	return &types.GetAllResponse{Records: records}, nil
}
