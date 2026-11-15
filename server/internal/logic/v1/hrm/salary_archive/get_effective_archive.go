package salary_archive

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/salary_scheme"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_archive"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_scheme"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_archive"
	schemeTypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_scheme"
)

type GetEffectiveArchive struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewGetEffectiveArchive(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *GetEffectiveArchive {
	return &GetEffectiveArchive{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// GetEffectiveArchive HRM-37 试算消费契约（导出包级函数 GetEffectiveArchive 供试算引擎复用）:
// 按 试算月份（YYYY-MM）找员工当时有效的薪资档案 ——
//
//	生效日期 ≤ 月末日 AND (失效日期为空 OR 失效日期 ≥ 月初日) AND status='1'（停用档案不返回）。
//
// 多条命中（理论上被重叠守卫阻止，但历史数据可能并存）取 生效日期最晚 的一条。
// 命中档案的方案可为空（未绑定方案）——此时 Items 为空数组，由试算侧决定兜底口径；
// 有方案时串 salary_scheme.GetSchemeItemsForScheme 取项目清单（calc_order 升序，含覆盖值）。
//
// 找不到生效档案 → SalaryArchiveNotFoundCode（试算侧据此报「员工无有效薪资档案」）。
func (l *GetEffectiveArchive) GetEffectiveArchive(req *types.GetEffectiveArchiveRequest) (resp *types.GetEffectiveArchiveResponse, err error) {
	return GetEffectiveArchiveForEmployee(l.ctx, l.svcCtx, req.EmployeeUuid, req.YearMonth)
}

// GetEffectiveArchiveForEmployee HRM-37 消费契约实现（包级导出；HTTP handler 与试算引擎共用）。
// yearMonth 格式 "YYYY-MM"；解析失败 → SalaryArchiveInvalidCode。
func GetEffectiveArchiveForEmployee(ctx context.Context, svcCtx *svc.ServiceContext, employeeUuid, yearMonth string) (*types.GetEffectiveArchiveResponse, error) {
	if employeeUuid == "" || len(yearMonth) != 7 {
		return nil, status.ErrorMessage(errcodes.SalaryArchiveInvalidCode, "employee_uuid 与 yearMonth(YYYY-MM) 不能为空")
	}
	monthFirst, monthLast, err := monthBounds(yearMonth)
	if err != nil {
		return nil, status.ErrorMessage(errcodes.SalaryArchiveInvalidCode, "yearMonth 格式必须是 YYYY-MM")
	}
	first := monthFirst.Format(dateLayout)
	last := monthLast.Format(dateLayout)

	rows, err := hrm_salary_archive.NewHrmSalaryArchiveModel(svcCtx.SqlxConn).FindByCondition(ctx, nil,
		condition.NewChain().
			Equal(hrm_salary_archive.EmployeeUuid, employeeUuid).
			Equal(hrm_salary_archive.Status, "1").
			Build()...)
	if err != nil {
		return nil, err
	}

	// 区间过滤（SQL 方言差异大，应用层判定）:
	//   effective_date <= 月末 AND (expire_date = '' OR expire_date >= 月初)
	var hit *hrm_salary_archive.HrmSalaryArchive
	for _, row := range rows {
		if row.EffectiveDate > last {
			continue
		}
		if row.ExpireDate != "" && row.ExpireDate < first {
			continue
		}
		// 取生效日期最晚的档案（跨档案切换时后档覆盖前档）
		if hit == nil || row.EffectiveDate > hit.EffectiveDate {
			hit = row
		}
	}
	if hit == nil {
		return nil, status.ErrorMessage(errcodes.SalaryArchiveNotFoundCode, "员工在 "+yearMonth+" 无有效的薪资档案")
	}

	record := archiveRowToRecord(hit)
	resp := &types.GetEffectiveArchiveResponse{
		Archive: record,
		Items:   []types.HrmEffectiveItem{},
	}

	if hit.SchemeUuid == "" {
		// 未绑定方案：方案摘要回填档案快照（subject），Items 留空
		resp.Scheme = types.HrmEffectiveScheme{
			SubjectUuid: hit.SubjectUuid,
			SubjectName: hit.SubjectName,
		}
		return resp, nil
	}

	scheme, err := hrm_salary_scheme.NewHrmSalarySchemeModel(svcCtx.SqlxConn).FindOneByUuid(ctx, nil, hit.SchemeUuid)
	if err != nil {
		if errors.Is(err, hrm_salary_scheme.ErrNotFound) {
			// 方案被删除：档案只读降级，返回方案空摘要+空项目（试算侧按无方案口径处理）
			resp.Scheme = types.HrmEffectiveScheme{
				SubjectUuid: hit.SubjectUuid,
				SubjectName: hit.SubjectName,
			}
			return resp, nil
		}
		return nil, err
	}
	resp.Scheme = types.HrmEffectiveScheme{
		Uuid:        scheme.Uuid,
		SchemeName:  scheme.SchemeName,
		SubjectUuid: scheme.SubjectUuid,
		SubjectName: scheme.SubjectName,
		SettleCycle: scheme.SettleCycle,
		PayDay:      scheme.PayDay,
		Currency:    scheme.Currency,
	}

	items, err := salary_scheme.GetSchemeItemsForScheme(ctx, svcCtx, hit.SchemeUuid)
	if err != nil {
		return nil, err
	}
	resp.Items = effectiveItemsFromSchemeItems(items)
	return resp, nil
}

// effectiveItemsFromSchemeItems 方案项目行 → 试算项目行（字段透传，顺序保持 calc_order 升序）.
func effectiveItemsFromSchemeItems(items []schemeTypes.HrmSalarySchemeItem) []types.HrmEffectiveItem {
	out := make([]types.HrmEffectiveItem, 0, len(items))
	for _, item := range items {
		out = append(out, types.HrmEffectiveItem{
			ItemUuid:        item.ItemUuid,
			ItemCode:        item.ItemCode,
			ItemName:        item.ItemName,
			ItemType:        item.ItemType,
			ItemClass:       item.ItemClass,
			ValueMode:       item.ValueMode,
			CalcOrder:       item.CalcOrder,
			Status:          item.Status,
			OverrideAmount:  item.OverrideAmount,
			OverrideFormula: item.OverrideFormula,
		})
	}
	return out
}

// monthBounds "YYYY-MM" → (月初 1 日, 月末日)（Local 时区）.
func monthBounds(yearMonth string) (time.Time, time.Time, error) {
	t, err := time.ParseInLocation("2006-01", yearMonth, time.Local)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	first := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.Local)
	last := first.AddDate(0, 1, -1)
	return first, last, nil
}
