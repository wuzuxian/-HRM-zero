package certificate

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/certificate"
)

type Revoke struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewRevoke(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Revoke {
	return &Revoke{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// statusInvalid maps an illegal 证书状态流转 to the shared business error.
func statusInvalid(current, certNo string) error {
	return status.ErrorMessage(errcodes.CertificateStatusInvalidCode,
		"证书状态流转非法（当前 "+certStatusText(current)+"）: "+certNo)
}

func certStatusText(s string) string {
	switch s {
	case CertStatusPending:
		return "待发"
	case CertStatusIssued:
		return "已发"
	case CertStatusRevoked:
		return "已撤销"
	}
	return s
}

// Revoke 撤销（行操作/批量）：已发（1）→已撤销（2）；待发/已撤销拒绝.
// 已撤销为终态（我的培训证书列表同步展示该状态）.
func (l *Revoke) Revoke(req *types.RevokeRequest) (resp *types.RevokeResponse, err error) {
	for _, uuid := range req.Uuids {
		row, err := l.svcCtx.Model.HrmTrainingCertificate.FindOneByUuid(l.ctx, nil, uuid)
		if err != nil {
			return nil, err
		}
		if row.Status != CertStatusIssued {
			return nil, statusInvalid(row.Status, row.CertNo)
		}
		if err = l.svcCtx.Model.HrmTrainingCertificate.UpdateFieldsByCondition(l.ctx, nil,
			map[string]any{"status": CertStatusRevoked},
			conditionUuid(uuid)...); err != nil {
			return nil, err
		}
	}

	l.Infof("certificate: 证书已撤销 count=%d", len(req.Uuids))
	return &types.RevokeResponse{}, nil
}
