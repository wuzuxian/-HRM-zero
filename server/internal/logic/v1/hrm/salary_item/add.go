package salary_item

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_item"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_item"
)

type Add struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewAdd(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Add {
	return &Add{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Add 新增薪资项目（GLOSSARY「薪资项目」）: 项目编码唯一；预置项目只能由种子迁移产生（is_preset=0）.
func (l *Add) Add(req *types.AddRequest) (resp *types.AddResponse, err error) {
	if req.Status == "" {
		req.Status = "1"
	}
	if err = validateItemCommon(req.ItemCode, req.ItemName, req.ItemType, req.ItemClass,
		req.ValueMode, req.Formula, req.Status); err != nil {
		return nil, err
	}

	// 项目编码唯一
	_, err = l.svcCtx.Model.HrmSalaryItem.FindOneByItemCode(l.ctx, nil, req.ItemCode)
	if err != nil && err != sqlx.ErrNotFound {
		return nil, err
	} else if err == nil {
		return nil, status.ErrorMessage(errcodes.SalaryItemCodeDuplicateCode, "项目编码 "+req.ItemCode+" 已存在")
	}

	_, err = l.svcCtx.Model.HrmSalaryItem.Insert(l.ctx, nil, &hrm_salary_item.HrmSalaryItem{
		Uuid:      uuid.New().String(),
		ItemCode:  req.ItemCode,
		ItemName:  req.ItemName,
		ItemType:  req.ItemType,
		ItemClass: req.ItemClass,
		ValueMode: req.ValueMode,
		CalcOrder: req.CalcOrder,
		OrderNum:  req.OrderNum,
		Formula:   normalizeFormula(req.ItemType, req.Formula),
		IsPreset:  0,
		Status:    req.Status,
		Remark:    req.Remark,
	})
	if err != nil {
		return nil, err
	}

	return &types.AddResponse{}, nil
}
