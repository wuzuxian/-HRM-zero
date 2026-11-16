package salary_scheme

import (
	"context"
	"errors"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_scheme"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_scheme"
)

type GetSchemeDetail struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewGetSchemeDetail(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *GetSchemeDetail {
	return &GetSchemeDetail{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// GetSchemeDetail 返回一个薪酬方案 + 其全量项目配置（按 calc_order 排序，配置弹窗回显用）。
func (l *GetSchemeDetail) GetSchemeDetail(req *types.GetSchemeDetailRequest) (resp *types.GetSchemeDetailResponse, err error) {
	row, err := l.svcCtx.Model.HrmSalaryScheme.FindOneByUuid(l.ctx, nil, req.Uuid)
	if err != nil {
		if errors.Is(err, hrm_salary_scheme.ErrNotFound) {
			return nil, status.ErrorMessage(errcodes.SalarySchemeNotFoundCode, "薪酬方案不存在")
		}
		return nil, err
	}

	completeness, itemCount, err := completenessFor(l.ctx, l.svcCtx.SqlxConn, row.Uuid)
	if err != nil {
		return nil, err
	}
	record := schemeRowToRecord(row)
	record.ItemCount = itemCount
	record.Completeness = completeness

	items, err := fetchSchemeItems(l.ctx, l.svcCtx.SqlxConn, row.Uuid)
	if err != nil {
		return nil, err
	}

	return &types.GetSchemeDetailResponse{
		Scheme: record,
		Items:  items,
	}, nil
}
