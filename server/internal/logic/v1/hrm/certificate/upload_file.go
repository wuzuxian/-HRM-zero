package certificate

import (
	"context"
	"net/http"
	"strings"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_certificate"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/certificate"
)

type UploadFile struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewUploadFile(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *UploadFile {
	return &UploadFile{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// UploadFile 上传/替换证书扫描件（行操作）：file_name + file_content(base64).
// 已撤销证书仍可补传扫描件（存档用途），状态不变.
func (l *UploadFile) UploadFile(req *types.UploadFileRequest) (resp *types.UploadFileResponse, err error) {
	if strings.TrimSpace(req.FileName) == "" {
		return nil, status.ErrorMessage(errcodes.CertificateInvalidCode, "扫描件文件名不能为空")
	}
	if strings.TrimSpace(req.FileContent) == "" {
		return nil, status.ErrorMessage(errcodes.CertificateInvalidCode, "扫描件内容不能为空")
	}

	if _, err = l.svcCtx.Model.HrmTrainingCertificate.FindOneByUuid(l.ctx, nil, req.Uuid); err != nil {
		if err == hrm_training_certificate.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.CertificateNotFoundCode, "证书不存在")
		}
		return nil, err
	}

	if err = l.svcCtx.Model.HrmTrainingCertificate.UpdateFieldsByCondition(l.ctx, nil,
		map[string]any{"file_name": req.FileName, "file_content": req.FileContent},
		conditionUuid(req.Uuid)...); err != nil {
		return nil, err
	}

	l.Infof("certificate: 证书扫描件已上传 file=%s", req.FileName)
	return &types.UploadFileResponse{}, nil
}
