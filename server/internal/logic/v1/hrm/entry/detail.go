package entry

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_entry_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/entry"
)

type Detail struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewDetail(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Detail {
	return &Detail{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Detail 入职单详情：主表 + 三子表快照（工作经历/教育经历/家庭成员）.
func (l *Detail) Detail(req *types.DetailRequest) (resp *types.DetailResponse, err error) {
	row, err := l.svcCtx.Model.HrmEntryBill.FindOneByUuid(l.ctx, nil, req.Uuid)
	if err != nil {
		if err == hrm_entry_bill.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.EntryNotFoundCode, "入职单不存在")
		}
		return nil, err
	}

	works, educations, families, err := findChildren(l.ctx, l.svcCtx, row.Uuid)
	if err != nil {
		return nil, err
	}

	out := entryRecord(row)
	out.Works = works
	out.Educations = educations
	out.Families = families
	return &types.DetailResponse{Entry: out}, nil
}
