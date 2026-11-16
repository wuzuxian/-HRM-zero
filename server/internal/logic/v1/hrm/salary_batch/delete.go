package salary_batch

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_batch"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_batch_item"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_batch"
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

// Delete 批删批次（连同试算明细行一并删除；确认/发薪/工资条进度归 HRM-38/39，本票不做拦截）。
func (l *Delete) Delete(req *types.DeleteRequest) (resp *types.DeleteResponse, err error) {
	if len(req.Uuids) == 0 {
		return nil, statusInvalid("uuids 不能为空")
	}
	txErr := l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		batchModel := hrm_salary_batch.NewHrmSalaryBatchModel(sqlx.NewSqlConnFromSession(session))
		itemModel := hrm_salary_batch_item.NewHrmSalaryBatchItemModel(sqlx.NewSqlConnFromSession(session))
		if err := batchModel.DeleteByCondition(ctx, session,
			condition.NewChain().In(hrm_salary_batch.Uuid, req.Uuids).Build()...); err != nil {
			return err
		}
		return itemModel.DeleteByCondition(ctx, session,
			condition.NewChain().In(hrm_salary_batch_item.BatchUuid, req.Uuids).Build()...)
	})
	if txErr != nil {
		return nil, txErr
	}
	l.Logger.Infof("salary batches deleted %d", len(req.Uuids))
	return &types.DeleteResponse{}, nil
}
