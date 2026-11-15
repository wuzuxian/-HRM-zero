package training

import (
	"errors"

	"github.com/jzero-io/jzero/core/status"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
)

// 我的培训（HRM-31）sentinel errors，映射到 errcodes 1036x。
var (
	// ErrCertificateNotFound: 证书不存在或无权访问（我的培训仅限证书本人）。
	ErrCertificateNotFound = errors.New("证书不存在或无权访问")
	// ErrCertificateFileNotFound: 证书扫描件不存在（未上传或已清除）。
	ErrCertificateFileNotFound = errors.New("证书扫描件不存在")
)

// wrapBizErr maps 我的培训 sentinel errors to registered business error codes;
// unknown errors pass through unchanged.
func wrapBizErr(err error) error {
	switch {
	case errors.Is(err, ErrCertificateNotFound):
		return status.Wrap(errcodes.TrainingCertificateNotFoundCode, err)
	case errors.Is(err, ErrCertificateFileNotFound):
		return status.Wrap(errcodes.TrainingCertificateFileNotFoundCode, err)
	default:
		return err
	}
}
