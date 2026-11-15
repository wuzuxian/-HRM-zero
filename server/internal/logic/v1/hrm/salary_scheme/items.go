package salary_scheme

import (
	"context"
	"errors"
	"sort"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_item"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_scheme"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_scheme_item"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_scheme"
)

// 方案详情 / 方案项目清单 共享实现。

// schemeItemWithSnapshot 一行关联配置 + 薪资项目实时快照（不落库，展示/试算实时取）。
type schemeItemWithSnapshot struct {
	types.HrmSalarySchemeItem
}

// fetchSchemeItems 取方案的项目配置，JOIN 薪资项目实时快照，按 calc_order ASC, id ASC 排序
// （hrm-33 契约：getSalaryItemList 顺序即计算顺序；关联表按入参顺序落库，消费时按项目计算顺序重排）。
func fetchSchemeItems(ctx context.Context, conn sqlx.SqlConn, schemeUuid string) ([]types.HrmSalarySchemeItem, error) {
	itemRows, err := hrm_salary_scheme_item.NewHrmSalarySchemeItemModel(conn).FindByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_salary_scheme_item.SchemeUuid, schemeUuid).OrderByAsc(hrm_salary_scheme_item.Id).Build()...)
	if err != nil {
		return nil, err
	}
	if len(itemRows) == 0 {
		return []types.HrmSalarySchemeItem{}, nil
	}

	itemUuids := make([]string, 0, len(itemRows))
	for _, row := range itemRows {
		itemUuids = append(itemUuids, row.ItemUuid)
	}
	items, err := hrm_salary_item.NewHrmSalaryItemModel(conn).FindByCondition(ctx, nil,
		condition.NewChain().In(hrm_salary_item.Uuid, itemUuids).
			OrderByAsc(hrm_salary_item.CalcOrder).OrderByAsc(hrm_salary_item.Id).Build()...)
	if err != nil {
		return nil, err
	}
	byUuid := make(map[string]*hrm_salary_item.HrmSalaryItem, len(items))
	for _, item := range items {
		byUuid[item.Uuid] = item
	}

	records := make([]types.HrmSalarySchemeItem, 0, len(itemRows))
	for _, row := range itemRows {
		record := types.HrmSalarySchemeItem{
			Uuid:            row.Uuid,
			SchemeUuid:      row.SchemeUuid,
			ItemUuid:        row.ItemUuid,
			OverrideAmount:  row.OverrideAmount.Float64,
			OverrideFormula: row.OverrideFormula,
		}
		if item, ok := byUuid[row.ItemUuid]; ok {
			record.ItemCode = item.ItemCode
			record.ItemName = item.ItemName
			record.ItemType = item.ItemType
			record.ItemClass = item.ItemClass
			record.ValueMode = item.ValueMode
			record.CalcOrder = item.CalcOrder
			// 停用项目 status='0' 容忍展示（hrm-33 契约 #4：试算/展示需容忍，过滤与否由消费侧定）
			record.Status = item.Status
		}
		records = append(records, record)
	}

	// 按项目计算顺序重排（calc_order ASC；同序保持关联表插入序 → 稳定排序）
	sort.SliceStable(records, func(i, j int) bool {
		return records[i].CalcOrder < records[j].CalcOrder
	})
	return records, nil
}

// GetSchemeItems HRM-37 试算消费契约：按方案取项目清单+计算规则（calc_order 升序，含覆盖值）。
func GetSchemeItemsForScheme(ctx context.Context, svcCtx *svc.ServiceContext, schemeUuid string) ([]types.HrmSalarySchemeItem, error) {
	if schemeUuid == "" {
		return nil, status.ErrorMessage(errcodes.SalarySchemeInvalidCode, "scheme_uuid 不能为空")
	}
	if _, err := svcCtx.Model.HrmSalaryScheme.FindOneByUuid(ctx, nil, schemeUuid); err != nil {
		if errors.Is(err, hrm_salary_scheme.ErrNotFound) {
			return nil, status.ErrorMessage(errcodes.SalarySchemeNotFoundCode, "薪酬方案不存在")
		}
		return nil, err
	}
	return fetchSchemeItems(ctx, svcCtx.SqlxConn, schemeUuid)
}
