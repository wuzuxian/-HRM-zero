package salary_archive

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_archive"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_archive"
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

// List 员工薪资档案分页列表（demo-inventory 8.2），按创建时间倒序；
// 筛选: 员工姓名（快照列 Like）/员工ID（工号快照 Like）/薪酬主体/薪酬方案/状态。
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	rows, total, err := l.svcCtx.Model.HrmSalaryArchive.PageByCondition(l.ctx, nil,
		condition.NewChain().
			Page(req.Current, req.Size).
			Like(hrm_salary_archive.EmployeeName, "%"+req.EmployeeName+"%", condition.WithSkip(req.EmployeeName == "")).
			Like(hrm_salary_archive.EmployeeNo, "%"+req.EmployeeNo+"%", condition.WithSkip(req.EmployeeNo == "")).
			Equal(hrm_salary_archive.SubjectUuid, req.SubjectUuid, condition.WithSkip(req.SubjectUuid == "")).
			Equal(hrm_salary_archive.SchemeUuid, req.SchemeUuid, condition.WithSkip(req.SchemeUuid == "")).
			Equal(hrm_salary_archive.Status, req.Status, condition.WithSkip(req.Status == "")).
			OrderByDesc(hrm_salary_archive.Id).
			Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.HrmSalaryArchive, 0, len(rows))
	for _, row := range rows {
		records = append(records, archiveRowToRecord(row))
	}

	return &types.ListResponse{PageResponse: types.PageResponse{
		Current: req.Current,
		Size:    req.Size,
		Total:   total,
	}, Records: records}, nil
}
