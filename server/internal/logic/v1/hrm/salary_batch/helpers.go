package salary_batch

import (
	"context"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/jzero-io/jzero/core/status"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
)

// loadOrgNames 一次性解析部门名称快照（org uuid → org_name，试算明细 org_name 列用）。
func loadOrgNames(ctx context.Context, svcCtx *svc.ServiceContext) (map[string]string, error) {
	orgs, err := svcCtx.Model.HrmOrg.FindByCondition(ctx, nil, condition.NewChain().Build()...)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(orgs))
	for _, o := range orgs {
		out[o.Uuid] = o.OrgName
	}
	return out, nil
}


// isArchiveNotFound 判定 HRM-35 GetEffectiveArchiveForEmployee 的「无有效档案」错误
//（errcode 10620）。status 错误无法 errors.Is 穿透，按 Code() 判定。
func isArchiveNotFound(err error) bool {
	if err == nil {
		return false
	}
	if sc := status.FromError(err); sc != nil {
		return sc.Code() == status.Code(10620)
	}
	return false
}
