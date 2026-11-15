package salary_item

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_item"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_item"
)

type Delete struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewDelete(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Delete {
	return &Delete{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Delete 删除/批量删除薪资项目.
// 守卫: 预置项目（is_preset=1，种子迁移产生）不可删除——如需下线请停用（status=0）。
// 混合批删（含预置项）整体拒绝，避免部分成功造成误解.
func (l *Delete) Delete(req *types.DeleteRequest) (resp *types.DeleteResponse, err error) {
	if len(req.Uuids) == 0 {
		return &types.DeleteResponse{}, nil
	}

	rows, err := l.svcCtx.Model.HrmSalaryItem.FindByCondition(l.ctx, nil, condition.Condition{
		Field:    hrm_salary_item.Uuid,
		Operator: condition.In,
		Value:    req.Uuids,
	})
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row.IsPreset == 1 {
			return nil, status.ErrorMessage(errcodes.SalaryItemPresetGuardCode,
				"预置薪资项目 "+row.ItemName+"（"+row.ItemCode+"）不可删除，如需下线请停用")
		}
	}

	err = l.svcCtx.Model.HrmSalaryItem.DeleteByCondition(l.ctx, nil, condition.Condition{
		Field:    hrm_salary_item.Uuid,
		Operator: condition.In,
		Value:    req.Uuids,
	})
	if err != nil {
		return nil, err
	}

	return &types.DeleteResponse{}, nil
}
