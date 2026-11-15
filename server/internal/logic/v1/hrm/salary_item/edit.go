package salary_item

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_item"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_item"
)

type Edit struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewEdit(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Edit {
	return &Edit{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Edit 编辑薪资项目: 编码唯一（排除自身）；预置项目允许编辑计算/显示顺序与启停，但不可删除（delete 守卫）.
func (l *Edit) Edit(req *types.EditRequest) (resp *types.EditResponse, err error) {
	if err = validateItemCommon(req.ItemCode, req.ItemName, req.ItemType, req.ItemClass,
		req.ValueMode, req.Formula, req.Status); err != nil {
		return nil, err
	}

	row, err := l.svcCtx.Model.HrmSalaryItem.FindOneByUuid(l.ctx, nil, req.Uuid)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.SalaryItemNotFoundCode, "薪资项目不存在")
		}
		return nil, err
	}

	// 项目编码唯一（排除自身）
	duplicated, err := l.svcCtx.Model.HrmSalaryItem.FindOneByItemCode(l.ctx, nil, req.ItemCode)
	if err != nil && err != sqlx.ErrNotFound {
		return nil, err
	}
	if duplicated != nil && duplicated.Uuid != row.Uuid {
		return nil, status.ErrorMessage(errcodes.SalaryItemCodeDuplicateCode, "项目编码 "+req.ItemCode+" 已存在")
	}

	err = l.svcCtx.Model.HrmSalaryItem.UpdateFieldsByCondition(l.ctx, nil, map[string]any{
		string(hrm_salary_item.ItemCode):  req.ItemCode,
		string(hrm_salary_item.ItemName):  req.ItemName,
		string(hrm_salary_item.ItemType):  req.ItemType,
		string(hrm_salary_item.ItemClass): req.ItemClass,
		string(hrm_salary_item.ValueMode): req.ValueMode,
		string(hrm_salary_item.CalcOrder): req.CalcOrder,
		string(hrm_salary_item.OrderNum):  req.OrderNum,
		string(hrm_salary_item.Formula):   normalizeFormula(req.ItemType, req.Formula),
		string(hrm_salary_item.Status):    req.Status,
		string(hrm_salary_item.Remark):    req.Remark,
	}, condition.Condition{
		Field:    hrm_salary_item.Uuid,
		Operator: condition.Equal,
		Value:    row.Uuid,
	})
	if err != nil {
		return nil, err
	}

	return &types.EditResponse{}, nil
}
