package certificate

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_cert_template"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_certificate"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_session"
	"github.com/jzero-io/jzero-admin/server/internal/model/manage_user"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/certificate"
)

type List struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewList(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *List {
	return &List{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// List 证书发放列表（管理端全量视图）：支持证书号/名称（模糊）/类型/状态/来源/班次筛选，
// 富化模板名称/班次名称/学员账号姓名（批量 IN 查询）.
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	rows, total, err := l.svcCtx.Model.HrmTrainingCertificate.PageByCondition(l.ctx, nil,
		condition.NewChain().
			Page(req.Current, req.Size).
			OrderByDesc(hrm_training_certificate.Id).
			Like(hrm_training_certificate.CertNo, "%"+strings.TrimSpace(req.CertNo)+"%",
				condition.WithSkip(strings.TrimSpace(req.CertNo) == "")).
			Like(hrm_training_certificate.CertName, "%"+strings.TrimSpace(req.CertName)+"%",
				condition.WithSkip(strings.TrimSpace(req.CertName) == "")).
			Equal(hrm_training_certificate.CertType, req.CertType, condition.WithSkip(req.CertType == "")).
			Equal(hrm_training_certificate.Status, req.Status, condition.WithSkip(req.Status == "")).
			Equal(hrm_training_certificate.Source, req.Source, condition.WithSkip(req.Source == "")).
			Equal(hrm_training_certificate.SessionUuid, req.SessionUuid, condition.WithSkip(req.SessionUuid == "")).
			Build()...)
	if err != nil {
		return nil, err
	}

	// batch enrichment: templates / sessions / users
	templateUuids := make([]string, 0, len(rows))
	sessionUuids := make([]string, 0, len(rows))
	userUuids := make([]string, 0, len(rows))
	for _, row := range rows {
		if row.TemplateUuid != "" {
			templateUuids = append(templateUuids, row.TemplateUuid)
		}
		if row.SessionUuid != "" {
			sessionUuids = append(sessionUuids, row.SessionUuid)
		}
		userUuids = append(userUuids, row.UserUuid)
	}
	templateRows, terr := l.svcCtx.Model.HrmTrainingCertTemplate.FindByCondition(l.ctx, nil,
		condition.NewChain().In(hrm_training_cert_template.Uuid, templateUuids).Build()...)
	templateNames := map[string]string{}
	if terr == nil {
		for _, r := range templateRows {
			templateNames[r.Uuid] = r.TemplateName
		}
	}
	sessionRows, serr := l.svcCtx.Model.HrmTrainingSession.FindByCondition(l.ctx, nil,
		condition.NewChain().In(hrm_training_session.Uuid, sessionUuids).Build()...)
	sessionNames := map[string]string{}
	if serr == nil {
		for _, r := range sessionRows {
			sessionNames[r.Uuid] = r.SessionName
		}
	}
	userRows, uerr := l.svcCtx.Model.ManageUser.FindByCondition(l.ctx, nil,
		condition.NewChain().In(manage_user.Uuid, userUuids).Build()...)
	userNames := map[string]*manage_user.ManageUser{}
	if uerr == nil {
		for _, u := range userRows {
			userNames[u.Uuid] = u
		}
	}

	records := make([]types.HrmCertificate, 0, len(rows))
	for _, row := range rows {
		issueTime := ""
		if row.IssueTime.Valid {
			issueTime = row.IssueTime.Time.Format(time.DateTime)
		}
		rec := types.HrmCertificate{
			Uuid:         row.Uuid,
			CertNo:       row.CertNo,
			CertName:     row.CertName,
			CertType:     row.CertType,
			TemplateUuid: row.TemplateUuid,
			TemplateName: templateNames[row.TemplateUuid],
			Source:       row.Source,
			SessionUuid:  row.SessionUuid,
			SessionName:  sessionNames[row.SessionUuid],
			UserUuid:     row.UserUuid,
			IssueTime:    issueTime,
			Status:       row.Status,
			FileName:     row.FileName,
			HasFile:      row.FileName != "" && row.FileContent.Valid && row.FileContent.String != "",
			CreateTime:   row.CreateTime.Format(time.DateTime),
			UpdateTime:   row.UpdateTime.Format(time.DateTime),
		}
		if u := userNames[row.UserUuid]; u != nil {
			rec.Username = u.Username
			rec.NickName = u.Nickname
		}
		records = append(records, rec)
	}

	return &types.ListResponse{PageResponse: types.PageResponse{
		Current: req.Current,
		Size:    req.Size,
		Total:   total,
	}, Records: records}, nil
}
