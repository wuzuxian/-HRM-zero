package tax_special_deduction

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_tax_special_deduction"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/tax_special_deduction"
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

func (l *Delete) Delete(req *types.DeleteRequest) (resp *types.DeleteResponse, err error) {
	if len(req.Uuids) == 0 {
		return &types.DeleteResponse{}, nil
	}

	err = l.svcCtx.Model.HrmTaxSpecialDeduction.DeleteByCondition(l.ctx, nil, condition.Condition{
		Field:    hrm_tax_special_deduction.Uuid,
		Operator: condition.In,
		Value:    req.Uuids,
	})
	if err != nil {
		return nil, err
	}

	return &types.DeleteResponse{}, nil
}
