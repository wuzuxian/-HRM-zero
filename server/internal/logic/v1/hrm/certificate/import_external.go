package certificate

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_certificate"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/certificate"
)

type ImportExternal struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewImportExternal(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *ImportExternal {
	return &ImportExternal{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// ImportExternal 导入外部证书：source='2'（无班次来源，template_uuid 留空），
// 证书号手动录入（外部编号体系，全局唯一），导入即发放（status='1' 已发）.
func (l *ImportExternal) ImportExternal(req *types.ImportExternalRequest) (resp *types.ImportExternalResponse, err error) {
	if strings.TrimSpace(req.UserUuid) == "" {
		return nil, status.ErrorMessage(errcodes.CertificateInvalidCode, "学员必填")
	}
	if strings.TrimSpace(req.CertName) == "" {
		return nil, status.ErrorMessage(errcodes.CertificateInvalidCode, "证书名称不能为空")
	}
	if strings.TrimSpace(req.CertNo) == "" {
		return nil, status.ErrorMessage(errcodes.CertificateInvalidCode, "证书号不能为空")
	}

	// 证书号全局唯一（hrm_training_certificate.cert_no UNIQUE 语义的业务前置检查）
	if _, err = l.svcCtx.Model.HrmTrainingCertificate.FindOneByCondition(l.ctx, nil,
		condition.NewChain().Equal(hrm_training_certificate.CertNo, req.CertNo).Build()...); err == nil {
		return nil, status.ErrorMessage(errcodes.CertificateNoDuplicateCode, "证书号已存在: "+req.CertNo)
	} else if err != hrm_training_certificate.ErrNotFound {
		return nil, err
	}

	if _, err = l.svcCtx.Model.HrmTrainingCertificate.Insert(l.ctx, nil, &hrm_training_certificate.HrmTrainingCertificate{
		Uuid:        uuid.New().String(),
		CertNo:      req.CertNo,
		CertName:    req.CertName,
		CertType:    req.CertType,
		Source:      CertSourceImport,
		UserUuid:    req.UserUuid,
		IssueTime:   nullTime(time.Now()),
		Status:      CertStatusIssued,
		FileName:    req.FileName,
		FileContent: nullString(req.FileContent),
	}); err != nil {
		return nil, err
	}

	l.Infof("certificate: 外部证书已导入 certNo=%s user=%s", req.CertNo, req.UserUuid)
	return &types.ImportExternalResponse{}, nil
}
