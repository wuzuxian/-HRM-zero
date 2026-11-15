package errcodes

import "github.com/jzero-io/jzero/core/status"

// HRM-40 绩效指标库与考核模板业务错误码。
const (
	// HrmPerfIndicatorCodeDuplicatedCode 指标编码已存在
	HrmPerfIndicatorCodeDuplicatedCode = 10350
	// HrmPerfIndicatorNotFoundCode 指标不存在
	HrmPerfIndicatorNotFoundCode = 10351
	// HrmPerfIndicatorInvalidCode 指标字段非法（编码/名称为空、枚举越界、权重越界）
	HrmPerfIndicatorInvalidCode = 10352
	// HrmPerfIndicatorInUseCode 指标已被考核模板引用，不能删除
	HrmPerfIndicatorInUseCode = 10353
	// HrmPerfTemplateNotFoundCode 考核模板不存在
	HrmPerfTemplateNotFoundCode = 10354
	// HrmPerfTemplateItemInvalidCode 模板指标配置非法（指标不存在/权重不大于0/指标重复）
	HrmPerfTemplateItemInvalidCode = 10355
	// HrmPerfTemplateWeightSumCode 模板指标权重之和必须等于 100%
	HrmPerfTemplateWeightSumCode = 10356
)

func init() {
	status.Register(HrmPerfIndicatorCodeDuplicatedCode)
	status.Register(HrmPerfIndicatorNotFoundCode)
	status.Register(HrmPerfIndicatorInvalidCode)
	status.Register(HrmPerfIndicatorInUseCode)
	status.Register(HrmPerfTemplateNotFoundCode)
	status.Register(HrmPerfTemplateItemInvalidCode)
	status.Register(HrmPerfTemplateWeightSumCode)
}
