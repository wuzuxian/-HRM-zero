package certificate

import (
	"context"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_certificate"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/certificate"
)

type Issue struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewIssue(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Issue {
	return &Issue{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Issue 发放（行操作/批量）：待发（0）→已发（1），置发证时间；已发/已撤销拒绝.
func (l *Issue) Issue(req *types.IssueRequest) (resp *types.IssueResponse, err error) {
	for _, uuid := range req.Uuids {
		row, err := l.svcCtx.Model.HrmTrainingCertificate.FindOneByUuid(l.ctx, nil, uuid)
		if err != nil {
			return nil, err
		}
		if row.Status != CertStatusPending {
			return nil, statusInvalid(row.Status, row.CertNo)
		}
		if err = l.svcCtx.Model.HrmTrainingCertificate.UpdateFieldsByCondition(l.ctx, nil,
			map[string]any{"status": CertStatusIssued, "issue_time": time.Now()},
			condition.NewChain().Equal(hrm_training_certificate.Uuid, uuid).Build()...); err != nil {
			return nil, err
		}
	}

	l.Infof("certificate: 证书已发放 count=%d", len(req.Uuids))
	return &types.IssueResponse{}, nil
}
