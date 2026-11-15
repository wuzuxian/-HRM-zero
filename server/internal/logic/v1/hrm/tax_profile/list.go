package tax_profile

import (
	"context"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_tax_profile"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/tax_profile"
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

func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	rows, total, err := l.svcCtx.Model.HrmTaxProfile.PageByCondition(l.ctx, nil,
		condition.NewChain().
			Page(req.Current, req.Size).
			Like(hrm_tax_profile.EmployeeName, "%"+req.EmployeeName+"%", condition.WithSkip(req.EmployeeName == "")).
			Equal(hrm_tax_profile.ResidentType, req.ResidentType, condition.WithSkip(req.ResidentType == "")).
			Equal(hrm_tax_profile.Status, req.Status, condition.WithSkip(req.Status == "")).
			OrderByDesc(hrm_tax_profile.Id).
			Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.HrmTaxProfile, 0, len(rows))
	for _, row := range rows {
		records = append(records, types.HrmTaxProfile{
			Uuid:                row.Uuid,
			EmployeeUuid:        row.EmployeeUuid,
			EmployeeName:        row.EmployeeName,
			IdCard:              row.IdCard,
			ResidentType:        row.ResidentType,
			TaxCity:             row.TaxCity,
			DefaultSpecialExtra: row.DefaultSpecialExtra,
			Status:              row.Status,
			Remark:              row.Remark,
			CreateTime:          row.CreateTime.Format(time.DateTime),
			UpdateTime:          row.UpdateTime.Format(time.DateTime),
		})
	}

	return &types.ListResponse{PageResponse: types.PageResponse{
		Current: req.Current,
		Size:    req.Size,
		Total:   total,
	}, Records: records}, nil
}
