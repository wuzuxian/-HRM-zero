package org

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/xuri/excelize/v2"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_org"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/org"
)

type Export struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewExport(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Export {
	return &Export{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx, r: r,
	}
}

// Export 导出组织: writes every org (flat, one row per node, 上级组织 resolved
// by name) into the shared excelize template and returns it base64-encoded.
func (l *Export) Export() (resp *types.ExportResponse, err error) {
	rows, err := l.svcCtx.Model.HrmOrg.FindByCondition(l.ctx, nil,
		condition.NewChain().OrderByAsc(hrm_org.OrderNum).Build()...)
	if err != nil {
		return nil, err
	}

	uuidToName := make(map[string]string, len(rows))
	for _, row := range rows {
		uuidToName[row.Uuid] = row.OrgName
	}

	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	if err = f.SetSheetRow(sheet, "A1", &importHeader); err != nil {
		return nil, err
	}

	for i, row := range rows {
		orgType := "部门"
		if row.OrgType == orgTypeCompany {
			orgType = "公司"
		}
		orgStatus := "关闭"
		if row.Status == orgStatusEnabled {
			orgStatus = "开启"
		}
		record := []string{
			row.OrgName,
			orgType,
			row.Leader,
			fmt.Sprintf("%d", row.OrderNum),
			orgStatus,
			uuidToName[row.ParentUuid],
		}
		cell, err := excelize.CoordinatesToCellName(1, i+2)
		if err != nil {
			return nil, err
		}
		if err = f.SetSheetRow(sheet, cell, &record); err != nil {
			return nil, err
		}
	}

	var buf bytes.Buffer
	if err = f.Write(&buf); err != nil {
		return nil, err
	}

	return &types.ExportResponse{
		FileName: fmt.Sprintf("组织管理_%s.xlsx", time.Now().Format("20060102150405")),
		Content:  base64.StdEncoding.EncodeToString(buf.Bytes()),
	}, nil
}
