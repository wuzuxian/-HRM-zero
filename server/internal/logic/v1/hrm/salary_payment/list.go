package salary_payment

import (
	"context"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	model "github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_payment"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_payment"
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

// List 发薪登记分页列表（batchUuid/payMode 筛选；demo 8.4 列：批次编号/名称、
// 薪酬主体、发薪金额、发薪方式、回执编号、到账时间）。
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	chain := condition.NewChain().
		Equal(model.BatchUuid, req.BatchUuid, condition.WithSkip(req.BatchUuid == "")).
		Equal(model.PayMode, req.PayMode, condition.WithSkip(req.PayMode == "")).
		OrderByDesc(model.Id).
		Page(req.Current, req.Size)

	rows, total, err := l.svcCtx.Model.HrmSalaryPayment.PageByCondition(l.ctx, nil, chain.Build()...)
	if err != nil {
		return nil, err
	}
	records := make([]types.HrmSalaryPayment, 0, len(rows))
	for _, row := range rows {
		records = append(records, paymentRowToRecord(row))
	}
	return &types.ListResponse{PageResponse: types.PageResponse{
		Current: req.Current,
		Size:    req.Size,
		Total:   total,
	}, Records: records}, nil
}

// paymentRowToRecord 登记行 → API record.
func paymentRowToRecord(row *model.HrmSalaryPayment) types.HrmSalaryPayment {
	return types.HrmSalaryPayment{
		Uuid:         row.Uuid,
		BatchUuid:    row.BatchUuid,
		BatchNo:      row.BatchNo,
		BatchName:    row.BatchName,
		SubjectUuid:  row.SubjectUuid,
		SubjectName:  row.SubjectName,
		PayAmount:    row.PayAmount,
		PayMode:      row.PayMode,
		ReceiptNo:    row.ReceiptNo,
		ArriveTime:   row.ArriveTime,
		RegisterBy:   row.RegisterBy,
		RegisterName: row.RegisterName,
		RegisterTime: row.RegisterTime,
		CreateTime:   row.CreateTime.Format(time.DateTime),
		UpdateTime:   row.UpdateTime.Format(time.DateTime),
	}
}
