package org

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_org"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/org"
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
		svcCtx: svcCtx, r: r,
	}
}

// List returns the 组织 tree. Filters (组织名称/组织类型/组织状态) keep matched
// nodes together with their ancestor chain so the tree stays navigable.
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	rows, err := l.svcCtx.Model.HrmOrg.FindByCondition(l.ctx, nil,
		condition.NewChain().OrderByAsc(hrm_org.OrderNum).Build()...)
	if err != nil {
		return nil, err
	}

	filtered := len(req.OrgName) > 0 || len(req.OrgType) > 0 || len(req.Status) > 0

	var selected []*hrm_org.HrmOrg
	if filtered {
		byUuid := make(map[string]*hrm_org.HrmOrg, len(rows))
		for _, row := range rows {
			byUuid[row.Uuid] = row
		}
		kept := make(map[string]bool, len(rows))
		var keepWithAncestors func(row *hrm_org.HrmOrg)
		keepWithAncestors = func(row *hrm_org.HrmOrg) {
			if row == nil || kept[row.Uuid] {
				return
			}
			kept[row.Uuid] = true
			if row.ParentUuid != "" {
				keepWithAncestors(byUuid[row.ParentUuid])
			}
		}
		for _, row := range rows {
			matched := (req.OrgName == "" || strings.Contains(row.OrgName, req.OrgName)) &&
				(req.OrgType == "" || row.OrgType == req.OrgType) &&
				(req.Status == "" || row.Status == req.Status)
			if matched {
				keepWithAncestors(row)
			}
		}
		for _, row := range rows {
			if kept[row.Uuid] {
				selected = append(selected, row)
			}
		}
	} else {
		selected = rows
	}

	records := buildTree(selected)

	return &types.ListResponse{Records: records}, nil
}

// buildTree assembles a flat, order_num-ordered org slice into nested tree
// records; roots are the rows whose parent is not inside the slice.
func buildTree(rows []*hrm_org.HrmOrg) []types.HrmOrg {
	type node struct {
		record   types.HrmOrg
		children []*node
	}

	nodes := make(map[string]*node, len(rows))
	list := make([]*node, 0, len(rows))
	for _, row := range rows {
		n := &node{record: types.HrmOrg{
			Uuid:       row.Uuid,
			ParentUuid: row.ParentUuid,
			OrgName:    row.OrgName,
			OrgType:    row.OrgType,
			Leader:     row.Leader,
			OrderNum:   row.OrderNum,
			Status:     row.Status,
			Remark:     row.Remark,
			CreateTime: row.CreateTime.Format(time.DateTime),
			UpdateTime: row.UpdateTime.Format(time.DateTime),
		}}
		nodes[row.Uuid] = n
		list = append(list, n)
	}

	records := make([]types.HrmOrg, 0, len(rows))
	var render func(n *node) types.HrmOrg
	render = func(n *node) types.HrmOrg {
		if len(n.children) > 0 {
			children := make([]types.HrmOrg, 0, len(n.children))
			for _, child := range n.children {
				children = append(children, render(child))
			}
			n.record.Children = children
		} else {
			n.record.Children = nil
		}
		return n.record
	}

	// link every child first, then render the roots
	var roots []*node
	for _, n := range list {
		parent, ok := nodes[n.record.ParentUuid]
		if n.record.ParentUuid != "" && ok && parent != n {
			parent.children = append(parent.children, n)
		} else {
			roots = append(roots, n)
		}
	}
	for _, n := range roots {
		records = append(records, render(n))
	}
	return records
}
