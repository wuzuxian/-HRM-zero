package certificate

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_certificate"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/certificate"
)

type DownloadFile struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewDownloadFile(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *DownloadFile {
	return &DownloadFile{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// DownloadFile 证书扫描件预览/下载：返回 base64 内容（与我的培训同一返回模式）.
func (l *DownloadFile) DownloadFile(req *types.DownloadFileRequest) (resp *types.DownloadFileResponse, err error) {
	row, err := l.svcCtx.Model.HrmTrainingCertificate.FindOneByUuid(l.ctx, nil, req.Uuid)
	if err != nil {
		if err == hrm_training_certificate.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.CertificateNotFoundCode, "证书不存在")
		}
		return nil, err
	}

	if row.FileName == "" || !row.FileContent.Valid || row.FileContent.String == "" {
		return nil, status.ErrorMessage(errcodes.CertificateFileNotFoundCode, "证书扫描件不存在")
	}

	return &types.DownloadFileResponse{
		FileName: row.FileName,
		Content:  row.FileContent.String,
	}, nil
}
