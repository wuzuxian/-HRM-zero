package salary_subject

import (
	"context"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_subject"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_subject"
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

// List 薪酬主体分页列表，按创建时间倒序，支持按主体名称/状态筛选.
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	rows, total, err := l.svcCtx.Model.HrmSalarySubject.PageByCondition(l.ctx, nil,
		condition.NewChain().
			Page(req.Current, req.Size).
			Like(hrm_salary_subject.SubjectName, "%"+req.SubjectName+"%", condition.WithSkip(req.SubjectName == "")).
			Equal(hrm_salary_subject.Status, req.Status, condition.WithSkip(req.Status == "")).
			OrderByDesc(hrm_salary_subject.Id).
			Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.HrmSalarySubject, 0, len(rows))
	for _, row := range rows {
		records = append(records, types.HrmSalarySubject{
			Uuid:        row.Uuid,
			SubjectName: row.SubjectName,
			OrgUuid:     row.OrgUuid,
			OrgName:     row.OrgName,
			PayAccount:  row.PayAccount,
			PayBank:     row.PayBank,
			Status:      row.Status,
			Remark:      row.Remark,
			CreateTime:  row.CreateTime.Format(time.DateTime),
			UpdateTime:  row.UpdateTime.Format(time.DateTime),
		})
	}

	return &types.ListResponse{PageResponse: types.PageResponse{
		Current: req.Current,
		Size:    req.Size,
		Total:   total,
	}, Records: records}, nil
}
