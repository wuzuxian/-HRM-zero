package bill

// 共享单据导出（final-review F5, spec 用户故事 23: 每类人事单据支持导出）。
//
// 口径：按 billType 直查各单据表，与页面列表同一套 mine 过滤 — 默认「我的单据」
// （create_by=当前用户），管理员级角色可看全部（hrmuser.CanSeeAllBills，spec
// 用户故事 4）。excelize base64 返回（org/monthreport 导出惯例）。
// 最小可用范围：人事管理四类单据 201 入职 / 202 转正 / 203 调动 / 204 离职；
// 其余类型的导出已由各自域覆盖（薪资/工资条/考勤等，demo-inventory）。
// （不直接 import 各类型 logic 包 — bill 是底座包，单向依赖，避免 import cycle。）
//
// 权限：authx casbin 按 route2code 生成码 v1:hrm:bill:exportBillList 精确校验
// （按钮码种子见迁移 56_bill_export.up.sql，人事管理目录下）。

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

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	hrmuser "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/hrmuser"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_entry_bill"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_regularization_bill"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_resignation_bill"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_transfer_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/bill"
	"github.com/jzero-io/jzero/core/status"
)

type ExportBillList struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewExportBillList(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *ExportBillList {
	return &ExportBillList{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// MaxExportRows 单次导出行数上限（防误操作全表拉爆；列表分页单次导出足量）。
const MaxExportRows = 10000

// billExportRow 导出统一行（四类单据列表展示列的公共子集）。
type billExportRow struct {
	BillNo       string // 单据编号
	Status       string // 单据状态
	EmployeeNo   string // 员工工号（入职单为空，员工姓名即申请人信息）
	EmployeeName string // 员工姓名
	CreatorName  string // 申请人
	CreateTime   string // 创建时间
}

// billStatusName 单据状态中文展示（GLOSSARY 单据状态机）。
func billStatusName(s string) string {
	switch s {
	case "unsubmitted":
		return "未提交"
	case "approving":
		return "审批中"
	case "approved":
		return "审批通过"
	case "rejected":
		return "审批不通过"
	case "revoked":
		return "已撤销"
	default:
		return s
	}
}

func (l *ExportBillList) ExportBillList(req *types.ExportBillListRequest) (resp *types.ExportBillListResponse, err error) {
	rows, err := l.exportRows(req)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	header := []string{"单据编号", "单据状态", "员工工号", "员工姓名", "申请人", "创建时间"}
	if err := f.SetSheetRow(sheet, "A1", &header); err != nil {
		return nil, err
	}
	for i, row := range rows {
		line := []string{
			row.BillNo, billStatusName(row.Status), row.EmployeeNo,
			row.EmployeeName, row.CreatorName, row.CreateTime,
		}
		cell, cellErr := excelize.CoordinatesToCellName(1, i+2)
		if cellErr != nil {
			return nil, cellErr
		}
		if err := f.SetSheetRow(sheet, cell, &line); err != nil {
			return nil, err
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return &types.ExportBillListResponse{
		FileName: fmt.Sprintf("hrm-bills-%s-%s.xlsx", req.BillType, time.Now().Format("20060102150405")),
		Content:  base64.StdEncoding.EncodeToString(buf.Bytes()),
	}, nil
}

// mineChain mine 过滤链（与各类型 list.go 同语义：默认仅本人单据，管理员级可看全部）。
func (l *ExportBillList) mineChain(req *types.ExportBillListRequest) condition.Chain {
	user := hrmuser.Current(l.ctx)
	onlyMine := req.Mine == 1 || !hrmuser.CanSeeAllBills(l.ctx, l.svcCtx, user)
	return condition.NewChain().Equal("create_by", user.Uuid, condition.WithSkip(!onlyMine))
}

// exportRows 按 billType 查询该类型全量单据。
func (l *ExportBillList) exportRows(req *types.ExportBillListRequest) ([]billExportRow, error) {
	switch req.BillType {
	case "201": // 员工入职单
		chain := l.mineChain(req).OrderByDesc(hrm_entry_bill.Id)
		rows, err := l.svcCtx.Model.HrmEntryBill.FindByCondition(l.ctx, nil, chain.Build()...)
		if err != nil {
			return nil, err
		}
		out := make([]billExportRow, 0, len(rows))
		for _, r := range rows {
			out = append(out, billExportRow{r.BillNo, r.Status, "", r.Name, r.CreatorName, r.CreateTime.Format("2006-01-02 15:04:05")})
		}
		return out, nil
	case "202": // 员工转正单
		chain := l.mineChain(req).OrderByDesc(hrm_regularization_bill.Id)
		rows, err := l.svcCtx.Model.HrmRegularizationBill.FindByCondition(l.ctx, nil, chain.Build()...)
		if err != nil {
			return nil, err
		}
		out := make([]billExportRow, 0, len(rows))
		for _, r := range rows {
			out = append(out, billExportRow{r.BillNo, r.Status, r.EmployeeNo, r.EmployeeName, r.CreatorName, r.CreateTime.Format("2006-01-02 15:04:05")})
		}
		return out, nil
	case "203": // 人事调动单
		chain := l.mineChain(req).OrderByDesc(hrm_transfer_bill.Id)
		rows, err := l.svcCtx.Model.HrmTransferBill.FindByCondition(l.ctx, nil, chain.Build()...)
		if err != nil {
			return nil, err
		}
		out := make([]billExportRow, 0, len(rows))
		for _, r := range rows {
			out = append(out, billExportRow{r.BillNo, r.Status, r.EmployeeNo, r.EmployeeName, r.CreatorName, r.CreateTime.Format("2006-01-02 15:04:05")})
		}
		return out, nil
	case "204": // 员工离职单
		chain := l.mineChain(req).OrderByDesc(hrm_resignation_bill.Id)
		rows, err := l.svcCtx.Model.HrmResignationBill.FindByCondition(l.ctx, nil, chain.Build()...)
		if err != nil {
			return nil, err
		}
		out := make([]billExportRow, 0, len(rows))
		for _, r := range rows {
			out = append(out, billExportRow{r.BillNo, r.Status, r.EmployeeNo, r.EmployeeName, r.CreatorName, r.CreateTime.Format("2006-01-02 15:04:05")})
		}
		return out, nil
	default:
		return nil, status.ErrorMessage(errcodes.HrmBillTypeNotRegisteredCode,
			fmt.Sprintf("单据类型 %s 不支持导出（当前支持 201/202/203/204）", req.BillType))
	}
}
