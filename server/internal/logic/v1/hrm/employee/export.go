package employee

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/xuri/excelize/v2"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/employee"
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

// Export 导出员工 (spec user story 13): writes every employee into the shared
// excelize template (组织/单位 resolved by name) and returns it base64-encoded.
func (l *Export) Export() (resp *types.ExportResponse, err error) {
	rows, err := l.svcCtx.Model.HrmEmployee.FindByCondition(l.ctx, nil,
		condition.NewChain().OrderByAsc(hrm_employee.EmployeeNo).Build()...)
	if err != nil {
		return nil, err
	}

	orgs, err := l.svcCtx.Model.HrmOrg.FindByCondition(l.ctx, nil)
	if err != nil {
		return nil, err
	}
	uuidToName := make(map[string]string, len(orgs))
	for _, org := range orgs {
		uuidToName[org.Uuid] = org.OrgName
	}

	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	if err = f.SetSheetRow(sheet, "A1", &importHeader); err != nil {
		return nil, err
	}

	for i, row := range rows {
		record := []string{
			row.EmployeeNo,
			row.Name,
			genderLabel(row.Gender),
			row.Phone,
			row.Email,
			uuidToName[row.OrgUuid],
			uuidToName[row.CompanyUuid],
			row.Position,
			row.Post,
			personStatusLabel(row.PersonStatus),
			row.HireDate,
			row.FormalDate,
			row.Birthday,
			row.Nation,
			row.PoliticalStatus,
			maritalStatusLabel(row.MaritalStatus),
			row.NativePlace,
			row.IdCard,
			row.HouseholdAddress,
			row.Address,
			row.EmergencyContact,
			row.EmergencyPhone,
			row.BankName,
			row.BankAccount,
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
		FileName: fmt.Sprintf("员工档案_%s.xlsx", time.Now().Format("20060102150405")),
		Content:  base64.StdEncoding.EncodeToString(buf.Bytes()),
	}, nil
}
