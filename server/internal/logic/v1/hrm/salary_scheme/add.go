package salary_scheme

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_scheme"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_scheme"
)

type Add struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewAdd(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Add {
	return &Add{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Add 新增薪酬方案（GLOSSARY「薪酬方案」）: 主体名称以快照落库；
// 方案项目配置整体写入关联表（先存方案、后配置项目，空 items 合法）。
func (l *Add) Add(req *types.AddRequest) (resp *types.AddResponse, err error) {
	if req.Status == "" {
		req.Status = "1"
	}
	req.SettleCycle, req.PayDay, req.Currency = normalizeSchemeDefaults(req.SettleCycle, req.PayDay, req.Currency)
	if err = validateSchemeCommon(req.SchemeName, req.PersonStatus, req.SettleCycle, req.PayDay, req.Status); err != nil {
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

	personStatus := joinPersonStatus(req.PersonStatus)
	schemeUuid := uuid.New().String()
	err = l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(_ context.Context, session sqlx.Session) error {
		if err := l.svcCtx.Model.HrmSalaryScheme.InsertV2(l.ctx, session, &hrm_salary_scheme.HrmSalaryScheme{
			Uuid:         schemeUuid,
			SchemeName:   req.SchemeName,
			SubjectUuid:  req.SubjectUuid,
			SubjectName:  subject.SubjectName,
			PersonStatus: personStatus,
			SettleCycle:  req.SettleCycle,
			PayDay:       req.PayDay,
			Currency:     req.Currency,
			Status:       req.Status,
			Remark:       req.Remark,
		}); err != nil {
			return err
		}
		return insertSchemeItems(l.ctx, l.svcCtx.SqlxConn, session, schemeUuid, items)
	})
	if err != nil {
		return nil, err
	}

	return &types.AddResponse{}, nil
}
