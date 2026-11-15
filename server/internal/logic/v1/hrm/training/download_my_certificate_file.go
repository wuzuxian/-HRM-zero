package training

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_certificate"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/training"
)

type DownloadMyCertificateFile struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewDownloadMyCertificateFile(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *DownloadMyCertificateFile {
	return &DownloadMyCertificateFile{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// DownloadMyCertificateFile 证书扫描件预览/下载：返回 base64 内容
// （与组织导出同一返回模式）。仅限证书本人访问。
func (l *DownloadMyCertificateFile) DownloadMyCertificateFile(req *types.DownloadMyCertificateFileRequest) (resp *types.DownloadMyCertificateFileResponse, err error) {
	userUuid := currentUserUuid(l.ctx)

	row, err := l.svcCtx.Model.HrmTrainingCertificate.FindOneByCondition(l.ctx, nil,
		condition.NewChain().
			Equal(hrm_training_certificate.Uuid, req.Uuid).
			Equal(hrm_training_certificate.UserUuid, userUuid).
			Build()...)
	if err != nil {
		if err == hrm_training_certificate.ErrNotFound {
			return nil, wrapBizErr(ErrCertificateNotFound)
		}
		return nil, err
	}

	if row.FileName == "" || !row.FileContent.Valid || row.FileContent.String == "" {
		return nil, wrapBizErr(ErrCertificateFileNotFound)
	}

	return &types.DownloadMyCertificateFileResponse{
		FileName: row.FileName,
		Content:  row.FileContent.String,
	}, nil
}
