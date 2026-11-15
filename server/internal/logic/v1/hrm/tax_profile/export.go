package tax_profile

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

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_tax_profile"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/tax_profile"
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
		svcCtx: svcCtx,
		r:      r,
	}
}

// exportHeader 个税档案导出列（demo-inventory 8.3: 员工、身份证、居民类型、税城市、默认专项附加、状态）。
var exportHeader = []string{"姓名", "身份证号", "居民类型", "税城市", "默认专项附加", "状态", "备注"}

func residentTypeLabel(v string) string {
	if v == ResidentTypeNonResident {
		return "非居民个人"
	}
	return "居民个人"
}

func statusLabel(v string) string {
	if v == "0" {
		return "停用"
	}
	return "启用"
}

// Export 导出员工个税档案（导出按需；档案无导入入口，员工信息走 HRM-01 员工导入）。
func (l *Export) Export() (resp *types.ExportResponse, err error) {
	rows, err := l.svcCtx.Model.HrmTaxProfile.FindByCondition(l.ctx, nil,
		condition.NewChain().OrderByAsc(hrm_tax_profile.Id).Build()...)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	if err = f.SetSheetRow(sheet, "A1", &exportHeader); err != nil {
		return nil, err
	}

	for i, row := range rows {
		record := []string{
			row.EmployeeName,
			row.IdCard,
			residentTypeLabel(row.ResidentType),
			row.TaxCity,
			row.DefaultSpecialExtra,
			statusLabel(row.Status),
			row.Remark,
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
		FileName: fmt.Sprintf("员工个税档案_%s.xlsx", time.Now().Format("20060102150405")),
		Content:  base64.StdEncoding.EncodeToString(buf.Bytes()),
	}, nil
}
