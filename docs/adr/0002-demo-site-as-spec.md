# HRM 功能规格以演示站走查为准，开源源码仅作字段佐证

ruoyi-office 仓库中开源的 `yudao-module-hrm` 只覆盖员工档案 + 4 类人事单据，招聘/培训/奖惩/薪酬/绩效/假勤为演示站闭源功能。我们决定以演示站（ruoyioffice.com）逐页走查结果作为功能规格基准（见 `docs/design/hrm/demo-inventory.md`），开源源码只用于佐证字段与业务规则；表结构、API、页面全部按本项目（jzero-admin）规范全新设计，不照搬 Java 代码。

## Consequences

- 数据库表不必与 `hrm_*` 开源表一一对应；对照开源代码排障时只能参考业务语义，不能参考 schema。
- 演示站后续更新可能造成规格漂移；再对齐时以 `demo-inventory.md` 的走查日期（2026-11-15）为版本锚点。
- 演示站的 AI 功能（企业管理智能体、邮箱线索入库、文本识别补录）明确裁剪，保留 Excel 导入与手工维护。
