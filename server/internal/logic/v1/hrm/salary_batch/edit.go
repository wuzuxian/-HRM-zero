package salary_batch

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_batch"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_batch_item"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_batch"
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

// Edit 编辑批次（名称/类型/年月/备注；主体不可换——换主体=换员工集，需删除重建）。
// 年月变更后已试算结果与新年月不匹配：重置试算汇总为未试算态（明细保留供追溯，由重新试算覆盖）。
func (l *Edit) Edit(req *types.EditRequest) (resp *types.EditResponse, err error) {
	if req.Uuid == "" {
		return nil, statusInvalid("uuid 不能为空")
	}
	if req.BatchName == "" {
		return nil, statusInvalid("批次名称不能为空")
	}
	if req.BatchType == "" {
		req.BatchType = BatchTypeMonthly
	}
	if req.BatchType != BatchTypeMonthly {
		return nil, statusInvalid("批次类型仅支持 monthly（月度批次）")
	}
	batch, err := findBatchByUuid(l.ctx, l.svcCtx, req.Uuid)
	if err != nil {
		return nil, err
	}
	_, _, err = validateBatchMonth(req.BatchYear, req.BatchMonth)
	if err != nil {
		return nil, err
	}

	if req.BatchYear != batch.BatchYear || req.BatchMonth != batch.BatchMonth || req.BatchType != batch.BatchType {
		// 唯一范围守卫（排除自身）
		existing, err := l.svcCtx.Model.HrmSalaryBatch.FindOneByCondition(l.ctx, nil,
			condition.NewChain().
				Equal(hrm_salary_batch.SubjectUuid, batch.SubjectUuid).
				Equal(hrm_salary_batch.BatchYear, req.BatchYear).
				Equal(hrm_salary_batch.BatchMonth, req.BatchMonth).
				Equal(hrm_salary_batch.BatchType, req.BatchType).
				Build()...)
		if err != nil && err != hrm_salary_batch.ErrNotFound {
			return nil, err
		}
		if existing != nil && existing.Uuid != req.Uuid {
			return nil, statusDuplicate("该主体下该年月已存在同类型批次")
		}
	}

	// 编号跟随月份首日重算语义：保持原编号不变（单据编号落库后不改，GLOSSARY 口径），
	// 年月变更仅影响试算取数月份；已试算数据按未试算态展示（trial_status 重置）。
	resetTrial := req.BatchYear != batch.BatchYear || req.BatchMonth != batch.BatchMonth
	fields := map[string]any{
		"batch_name":  req.BatchName,
		"batch_type":  req.BatchType,
		"batch_year":  req.BatchYear,
		"batch_month": req.BatchMonth,
		"remark":      req.Remark,
	}
	if resetTrial {
		fields["trial_status"] = "0"
		fields["headcount"] = 0
		fields["earning_total"] = 0
		fields["deduction_total"] = 0
		fields["net_total"] = 0
		fields["warnings"] = ""
	}
	if err := l.svcCtx.Model.HrmSalaryBatch.UpdateFieldsByCondition(l.ctx, nil, fields,
		condition.NewChain().Equal(hrm_salary_batch.Uuid, req.Uuid).Build()...); err != nil {
		return nil, err
	}
	if resetTrial {
		if err := l.svcCtx.Model.HrmSalaryBatchItem.DeleteByCondition(l.ctx, nil,
			condition.NewChain().Equal(hrm_salary_batch_item.BatchUuid, req.Uuid).Build()...); err != nil {
			return nil, err
		}
	}
	return &types.EditResponse{}, nil
}
