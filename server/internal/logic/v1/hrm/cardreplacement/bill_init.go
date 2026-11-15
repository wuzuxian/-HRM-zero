package cardreplacement

// 包初始化：注册 补卡单 类型码 206 到共享单据底座（HRM-06 registry）。
// 本包由 handler（generated routes）引用，server 编译即完成注册.

func init() {
	registerBillHandler()
}
