package tax_special_deduction

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/xuri/excelize/v2"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_tax_special_deduction"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/tax_special_deduction"
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

func (l *Export) Export() (resp *types.ExportResponse, err error) {
	rows, err := l.svcCtx.Model.HrmTaxSpecialDeduction.FindByCondition(l.ctx, nil,
		condition.NewChain().
			OrderByAsc(hrm_tax_special_deduction.EmployeeUuid).
			OrderByAsc(hrm_tax_special_deduction.DeductYear).
			OrderByAsc(hrm_tax_special_deduction.DeductMonth).
			OrderByAsc(hrm_tax_special_deduction.Id).
			Build()...)
	if err != nil {
		return nil, err
	}

	employees, err := l.svcCtx.Model.HrmEmployee.FindByCondition(l.ctx, nil)
	if err != nil {
		return nil, err
	}
	idCardByUuid := make(map[string]string, len(employees))
	for _, emp := range employees {
		idCardByUuid[emp.Uuid] = emp.IdCard
	}

	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	if err = f.SetSheetRow(sheet, "A1", &importHeader); err != nil {
		return nil, err
	}

	for i, row := range rows {
		record := []string{
			row.EmployeeName,
			idCardByUuid[row.EmployeeUuid],
			row.DeductYear,
			strconv.FormatInt(row.DeductMonth, 10),
			deductTypeName(row.DeductType),
			strconv.FormatFloat(row.Amount, 'f', 2, 64),
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
		FileName: fmt.Sprintf("专项附加扣除_%s.xlsx", time.Now().Format("20060102150405")),
		Content:  base64.StdEncoding.EncodeToString(buf.Bytes()),
	}, nil
}
