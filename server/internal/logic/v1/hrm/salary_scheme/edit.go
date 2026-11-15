package salary_scheme

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_scheme_item"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero/core/status"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_scheme"
)

type Edit struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewEdit(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Edit {
	return &Edit{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Edit 编辑薪酬方案：字段更新 + 项目配置整体替换（旧配置全部作废，事务内）。
// 关联表全删全插在同一个事务里，失败则整体回滚。
func (l *Edit) Edit(req *types.EditRequest) (resp *types.EditResponse, err error) {
	if req.Status == "" {
		req.Status = "1"
	}
	req.SettleCycle, req.PayDay, req.Currency = normalizeSchemeDefaults(req.SettleCycle, req.PayDay, req.Currency)
	if err = validateSchemeCommon(req.SchemeName, req.PersonStatus, req.SettleCycle, req.PayDay, req.Status); err != nil {
		return nil, err
	}
	row, err := l.svcCtx.Model.HrmSalaryScheme.FindOneByUuid(l.ctx, nil, req.Uuid)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.SalarySchemeNotFoundCode, "薪酬方案不存在")
		}
		return nil, err
	}
	subject, err := resolveSubject(l.ctx, l.svcCtx.SqlxConn, req.SubjectUuid)
	if err != nil {
		return nil, err
	}
	items, err := validateSchemeItems(l.ctx, l.svcCtx.SqlxConn, req.Items)
	if err != nil {
		return nil, err
	}

	row.SchemeName = req.SchemeName
	row.SubjectUuid = req.SubjectUuid
	row.SubjectName = subject.SubjectName
	row.PersonStatus = joinPersonStatus(req.PersonStatus)
	row.SettleCycle = req.SettleCycle
	row.PayDay = req.PayDay
	row.Currency = req.Currency
	row.Status = req.Status
	row.Remark = req.Remark

	err = l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(_ context.Context, session sqlx.Session) error {
		if err := l.svcCtx.Model.HrmSalaryScheme.Update(l.ctx, session, row); err != nil {
			return err
		}
		if err := l.svcCtx.Model.HrmSalarySchemeItem.DeleteByCondition(l.ctx, session,
			condition.NewChain().Equal(hrm_salary_scheme_item.SchemeUuid, req.Uuid).Build()...); err != nil {
			return err
		}
		return insertSchemeItems(l.ctx, l.svcCtx.SqlxConn, session, req.Uuid, items)
	})
	if err != nil {
		return nil, err
	}

	return &types.EditResponse{}, nil
}
