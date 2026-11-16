package salary_payslip

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_payslip"
)

type Publish struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewPublish(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Publish {
	return &Publish{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Publish 按批次发布工资条：从批次试算明细行聚合生成每员工工资条。
// 守卫：批次须已确认（confirm_status='1'，HRM-38 薪资确认审批写回）。
// 幂等：重发布覆盖旧明细（UNIQUE(batch_uuid,employee_uuid) 先删后插）。
func (l *Publish) Publish(req *types.PublishRequest) (resp *types.PublishResponse, err error) {
	batch, err := findBatchByUuid(l.ctx, l.svcCtx.Model, req.BatchUuid)
	if err != nil {
		return nil, err
	}
	if batch.ConfirmStatus != "1" {
		return nil, statusNotConfirmed("批次尚未确认，不能发布工资条")
	}

	var count int64
	if err := l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		var txErr error
		count, txErr = generateForBatch(ctx, session, l.svcCtx.Model, batch)
		return txErr
	}); err != nil {
		return nil, err
	}
	return &types.PublishResponse{BatchUuid: batch.Uuid, Count: count}, nil
}
