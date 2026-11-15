package training

import (
	"context"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_certificate"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/training"
)

type GetMyCertificateList struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewGetMyCertificateList(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *GetMyCertificateList {
	return &GetMyCertificateList{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// GetMyCertificateList 我的培训证书：当前登录学员名下的证书列表，
// 支持按状态筛选、分页。扫描件以 hasFile + fileName 标记，预览/下载走 DownloadMyCertificateFile。
func (l *GetMyCertificateList) GetMyCertificateList(req *types.GetMyCertificateListRequest) (resp *types.GetMyCertificateListResponse, err error) {
	userUuid := currentUserUuid(l.ctx)

	rows, total, err := l.svcCtx.Model.HrmTrainingCertificate.PageByCondition(l.ctx, nil,
		condition.NewChain().
			Page(req.Current, req.Size).
			Equal(hrm_training_certificate.UserUuid, userUuid).
			Equal(hrm_training_certificate.Status, req.Status, condition.WithSkip(req.Status == "")).
			OrderByDesc(hrm_training_certificate.Id).
			Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.HrmMyTrainingCertificate, 0, len(rows))
	for _, row := range rows {
		issueTime := ""
		if row.IssueTime.Valid {
			issueTime = row.IssueTime.Time.Format(time.DateTime)
		}
		records = append(records, types.HrmMyTrainingCertificate{
			Uuid:        row.Uuid,
			CertNo:      row.CertNo,
			CertName:    row.CertName,
			CertType:    row.CertType,
			Source:      row.Source,
			SessionUuid: row.SessionUuid,
			IssueTime:   issueTime,
			Status:      row.Status,
			FileName:    row.FileName,
			HasFile:     row.FileName != "" && row.FileContent.Valid && row.FileContent.String != "",
			CreateTime:  row.CreateTime.Format(time.DateTime),
			UpdateTime:  row.UpdateTime.Format(time.DateTime),
		})
	}

	return &types.GetMyCertificateListResponse{PageResponse: types.PageResponse{
		Current: req.Current,
		Size:    req.Size,
		Total:   total,
	}, Records: records}, nil
}
