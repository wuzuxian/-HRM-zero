package errcodes

import "github.com/jzero-io/jzero/core/status"

// HRM-34 薪酬方案业务错误码（1061x；1060x 已被加班单占用——feat/hrm 并行期把加班 1057x 重编到 1060x，
// 本票开工时取的 1060x 与之撞码，合并后让位改取 1061x）。
const (
	// SalarySchemeNotFoundCode 薪酬方案不存在
	SalarySchemeNotFoundCode = 10610
	// SalarySchemeInvalidCode 薪酬方案参数非法（名称为空/适用人员状态为空/结算周期/发薪日/币种/状态枚举越界）
	SalarySchemeInvalidCode = 10611
	// SalarySchemeSubjectNotFoundCode 关联薪酬主体不存在（subject_uuid 须指向 hrm_salary_subject 已有行）
	SalarySchemeSubjectNotFoundCode = 10612
	// SalarySchemeItemNotFoundCode 方案引用的薪资项目不存在（item_uuid 须指向 hrm_salary_item 已有行）
	SalarySchemeItemNotFoundCode = 10613
	// SalarySchemeItemDuplicateCode 同一方案内薪资项目重复勾选
	SalarySchemeItemDuplicateCode = 10614
	// SalarySchemeNotEditableCode 薪酬方案不可编辑/删除（HRM-37 薪资档案/批次引用后由消费侧声明，本侧预留）
	SalarySchemeNotEditableCode = 10615
)

func init() {
	status.Register(SalarySchemeNotFoundCode)
	status.Register(SalarySchemeInvalidCode)
	status.Register(SalarySchemeSubjectNotFoundCode)
	status.Register(SalarySchemeItemNotFoundCode)
	status.Register(SalarySchemeItemDuplicateCode)
	status.Register(SalarySchemeNotEditableCode)
}
