package org

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_org"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/org"
)

// MemberCountSource 返回"本级人数"数据源：org uuid → 直属该组织的在册员工数。
// 缺失的 org uuid 视为 0。
//
// 人数汇总口径：
//   - 本级人数（MemberCount）= 数据源给出的该组织直接在册员工数；
//   - 汇总人数（TotalMemberCount）= 本级人数 + 全部下级组织的汇总人数（自底向上累加）。
type MemberCountSource func(ctx context.Context, orgUuids []string) (map[string]int64, error)

// defaultMemberCounts is the production 本级人数 source: 在册员工 grouped by
// 所属组织 (hrm_employee.org_uuid). 员工档案 (HRM-03) has not landed yet, so the
// table may not exist — that transitional state is logged and counted as zero
// members instead of failing the whole chart; no change is needed here once
// hrm_employee arrives.
func defaultMemberCounts(svcCtx *svc.ServiceContext) MemberCountSource {
	return func(ctx context.Context, orgUuids []string) (map[string]int64, error) {
		// ServiceContext (core-engine) is the nil-able embedded pointer; logic
		// tests construct the outer struct with it unset.
		if svcCtx == nil || svcCtx.ServiceContext == nil || len(orgUuids) == 0 {
			return map[string]int64{}, nil
		}
		return countEmployeesByOrg(ctx, svcCtx.SqlxConn, orgUuids)
	}
}

type OrgChart struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request

	// memberCounts 本级人数来源；tests may swap this seam.
	memberCounts MemberCountSource
}

func NewOrgChart(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *OrgChart {
	return &OrgChart{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,

		memberCounts: defaultMemberCounts(svcCtx),
	}
}

// OrgChart returns the 可视化组织架构图 tree: every node carries 负责人、
// 本级人数、汇总人数 so the frontend can render and expand/collapse it.
func (l *OrgChart) OrgChart() (resp *types.OrgChartResponse, err error) {
	rows, err := l.svcCtx.Model.HrmOrg.FindByCondition(l.ctx, nil,
		condition.NewChain().OrderByAsc(hrm_org.OrderNum).Build()...)
	if err != nil {
		return nil, err
	}

	orgUuids := make([]string, 0, len(rows))
	for _, row := range rows {
		orgUuids = append(orgUuids, row.Uuid)
	}
	direct, err := l.memberCounts(l.ctx, orgUuids)
	if err != nil {
		return nil, err
	}

	records := buildChartTree(rows, direct)

	return &types.OrgChartResponse{Records: records}, nil
}

// buildChartTree assembles the flat org slice (order_num ordered) into chart
// nodes, rolling member counts bottom-up: each node's 汇总人数 = its 本级人数
// plus the 汇总人数 of every child. The source only guarantees direct counts,
// so subtotals are computed here.
func buildChartTree(rows []*hrm_org.HrmOrg, direct map[string]int64) []types.HrmOrgChartNode {
	type node struct {
		record   types.HrmOrgChartNode
		children []*node
	}

	nodes := make(map[string]*node, len(rows))
	list := make([]*node, 0, len(rows))
	for _, row := range rows {
		n := &node{record: types.HrmOrgChartNode{
			Uuid:       row.Uuid,
			ParentUuid: row.ParentUuid,
			OrgName:    row.OrgName,
			OrgType:    row.OrgType,
			Leader:     row.Leader,
			OrderNum:   row.OrderNum,
			Status:     row.Status,
			// 本级人数；汇总人数 filled in by render (bottom-up).
			MemberCount: direct[row.Uuid],
		}}
		nodes[row.Uuid] = n
		list = append(list, n)
	}

	var roots []*node
	for _, n := range list {
		parent, ok := nodes[n.record.ParentUuid]
		if n.record.ParentUuid != "" && ok && parent != n {
			parent.children = append(parent.children, n)
		} else {
			roots = append(roots, n)
		}
	}

	var render func(n *node) types.HrmOrgChartNode
	render = func(n *node) types.HrmOrgChartNode {
		total := n.record.MemberCount
		if len(n.children) > 0 {
			children := make([]types.HrmOrgChartNode, 0, len(n.children))
			for _, child := range n.children {
				childRecord := render(child)
				total += childRecord.TotalMemberCount
				children = append(children, childRecord)
			}
			n.record.Children = children
		} else {
			n.record.Children = nil
		}
		n.record.TotalMemberCount = total
		return n.record
	}

	records := make([]types.HrmOrgChartNode, 0, len(rows))
	for _, n := range roots {
		records = append(records, render(n))
	}
	return records
}
