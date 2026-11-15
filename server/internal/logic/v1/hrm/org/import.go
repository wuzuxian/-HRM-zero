package org

import (
	"bytes"
	"context"
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/xuri/excelize/v2"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/org"
)

type Import struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewImport(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Import {
	return &Import{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx, r: r,
	}
}

// importHeader is the excelize template shared by Import and Export.
var importHeader = []string{"组织名称", "组织类型", "负责人", "显示顺序", "组织状态", "上级组织"}

// Import 导入组织: reads the shared template (header + one org per row) and
// resolves 上级组织 by name — either an existing org or a previously imported
// row, so a whole subtree can be imported top-down in one file.
func (l *Import) Import(req *types.ImportRequest) (resp *types.ImportResponse, err error) {
	raw, err := base64.StdEncoding.DecodeString(req.Content)
	if err != nil {
		return nil, status.ErrorMessage(errcodes.OrgImportInvalidCode, "导入文件内容不是合法的 base64")
	}

	f, err := excelize.OpenReader(bytes.NewReader(raw))
	if err != nil {
		return nil, status.ErrorMessage(errcodes.OrgImportInvalidCode, "导入文件必须是 xlsx 格式")
	}
	defer f.Close()

	rows, err := f.GetRows(f.GetSheetName(0))
	if err != nil {
		return nil, status.ErrorMessage(errcodes.OrgImportInvalidCode, "导入文件读取失败")
	}
	if len(rows) < 2 {
		return nil, status.ErrorMessage(errcodes.OrgImportInvalidCode, "导入文件没有数据行")
	}

	// known org names -> uuid, pre-seeded with existing orgs so imports can
	// attach to the current tree
	existing, err := l.svcCtx.Model.HrmOrg.FindByCondition(l.ctx, nil)
	if err != nil {
		return nil, err
	}
	nameToUuid := make(map[string]string, len(existing))
	for _, row := range existing {
		if _, ok := nameToUuid[row.OrgName]; !ok {
			nameToUuid[row.OrgName] = row.Uuid
		}
	}

	var successCount int64
	for _, row := range rows {
		if strings.TrimSpace(firstCell(row)) == importHeader[0] || firstCell(row) == "" {
			continue // header or blank line
		}
		orgName := strings.TrimSpace(firstCell(row))

		orgType, err := parseOrgType(cellAt(row, 1))
		if err != nil {
			return nil, err
		}
		orgStatus, err := parseOrgStatus(cellAt(row, 4))
		if err != nil {
			return nil, err
		}
		leader := strings.TrimSpace(cellAt(row, 2))
		orderNum, err := strconv.ParseInt(strings.TrimSpace(cellAt(row, 3)), 10, 64)
		if err != nil {
			orderNum = 0
		}

		parentName := strings.TrimSpace(cellAt(row, 5))
		parentUuid := ""
		if parentName != "" {
			u, ok := nameToUuid[parentName]
			if !ok {
				return nil, status.ErrorMessage(errcodes.OrgParentNotFoundCode, "上级组织不存在: "+parentName)
			}
			parentUuid = u
		}

		inserted, err := insertOne(l.ctx, l.svcCtx, parentUuid, orgName, orgType, leader, orderNum, orgStatus, "")
		if err != nil {
			return nil, err
		}
		nameToUuid[orgName] = inserted.Uuid
		successCount++
	}

	return &types.ImportResponse{SuccessCount: successCount}, nil
}

func firstCell(row []string) string {
	return cellAt(row, 0)
}

func cellAt(row []string, i int) string {
	if i < len(row) {
		return row[i]
	}
	return ""
}

func parseOrgType(v string) (string, error) {
	switch strings.TrimSpace(v) {
	case "公司":
		return orgTypeCompany, nil
	case "部门":
		return orgTypeDepartment, nil
	case orgTypeCompany, orgTypeDepartment:
		return strings.TrimSpace(v), nil
	default:
		return "", status.ErrorMessage(errcodes.OrgImportInvalidCode, "组织类型必须是 公司 或 部门")
	}
}

func parseOrgStatus(v string) (string, error) {
	switch strings.TrimSpace(v) {
	case "开启", "":
		return orgStatusEnabled, nil
	case "关闭":
		return orgStatusDisabled, nil
	case orgStatusEnabled, orgStatusDisabled:
		return strings.TrimSpace(v), nil
	default:
		return "", status.ErrorMessage(errcodes.OrgImportInvalidCode, "组织状态必须是 开启 或 关闭")
	}
}
