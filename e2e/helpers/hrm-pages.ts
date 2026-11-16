/**
 * 全部 62 个 HRM 叶子页面路由（与 web/src/views/hrm/**\/index.vue 一一对应）。
 * 增删页面时同步维护；冒烟层遍历此清单。
 */
export const HRM_PAGES: { route: string; module: string; name: string }[] = [
  // 组织
  { route: '/hrm/org/orginfo', module: 'org', name: '组织管理' },
  { route: '/hrm/org/org-chart', module: 'org', name: '组织架构图' },
  // 假勤管理
  { route: '/hrm/attendance/setting/group', module: 'attendance', name: '考勤组' },
  { route: '/hrm/attendance/setting/attendance-config', module: 'attendance', name: '组织默认规则' },
  { route: '/hrm/attendance/setting/shift', module: 'attendance', name: '班次管理' },
  { route: '/hrm/attendance/setting/leave-type', module: 'attendance', name: '假期类型' },
  { route: '/hrm/attendance/setting/overtime-rule', module: 'attendance', name: '加班规则' },
  { route: '/hrm/attendance/apply/leave-cancel-list', module: 'attendance', name: '请假销假' },
  { route: '/hrm/attendance/apply/overtime-bill-list', module: 'attendance', name: '加班申请' },
  { route: '/hrm/attendance/apply/card-replacement-list', module: 'attendance', name: '补卡管理' },
  { route: '/hrm/attendance/data/leave-balance', module: 'attendance', name: '假期余额' },
  { route: '/hrm/attendance/data/punch-record', module: 'attendance', name: '打卡记录' },
  { route: '/hrm/attendance/data/punch-statistics', module: 'attendance', name: '考勤统计' },
  { route: '/hrm/attendance/data/month-report', module: 'attendance', name: '考勤月报' },
  { route: '/hrm/attendance/my/attendance-punch', module: 'attendance', name: '考勤打卡' },
  { route: '/hrm/attendance/my/dashboard', module: 'attendance', name: '个人看板' },
  { route: '/hrm/attendance/scheduling/schedule-cycle', module: 'attendance', name: '排班周期' },
  { route: '/hrm/attendance/scheduling/schedule-board', module: 'attendance', name: '排班表' },
  // 人事档案
  { route: '/hrm/employee/employee-archive-list', module: 'employee', name: '员工档案' },
  { route: '/hrm/employee/employee-contract-expiring', module: 'employee', name: '劳动合同到期预警' },
  { route: '/hrm/employee/employee-contract-renewable', module: 'employee', name: '待续签合同' },
  { route: '/hrm/employee/certificate-expiring', module: 'employee', name: '证照到期预警' },
  // 人事管理
  { route: '/hrm/employee-relation/entry-list', module: 'employee-relation', name: '员工入职' },
  { route: '/hrm/employee-relation/regular-list', module: 'employee-relation', name: '员工转正' },
  { route: '/hrm/employee-relation/resignation-list', module: 'employee-relation', name: '员工离职' },
  { route: '/hrm/employee-relation/transfer-list', module: 'employee-relation', name: '人事调动' },
  // 绩效管理
  { route: '/hrm/performance/setting/indicator', module: 'performance', name: '指标库' },
  { route: '/hrm/performance/setting/template', module: 'performance', name: '模板管理' },
  { route: '/hrm/performance/scheme', module: 'performance', name: '绩效方案' },
  { route: '/hrm/performance/period-list', module: 'performance', name: '绩效批次工作台' },
  { route: '/hrm/performance/assessment-task/workbench', module: 'performance', name: '考核任务工作台' },
  { route: '/hrm/performance/assessment-task/plan-list', module: 'performance', name: '目标确认' },
  { route: '/hrm/performance/assessment-task/self-review-list', module: 'performance', name: '自评填报' },
  { route: '/hrm/performance/assessment-task/leader-review-list', module: 'performance', name: '评审填报' },
  { route: '/hrm/performance/assessment-task/result-confirm-task-list', module: 'performance', name: '结果确认' },
  { route: '/hrm/performance/result-confirm-list', module: 'performance', name: '批次结果审批' },
  { route: '/hrm/performance/result-list', module: 'performance', name: '绩效结果' },
  { route: '/hrm/performance/dashboard', module: 'performance', name: '绩效驾驶舱' },
  // 招聘管理
  { route: '/hrm/recruitment/requirement-list', module: 'recruitment', name: '招聘需求' },
  { route: '/hrm/recruitment/position-list', module: 'recruitment', name: '招聘职位' },
  { route: '/hrm/recruitment/candidate-list', module: 'recruitment', name: '候选人中心' },
  { route: '/hrm/recruitment/offer-list', module: 'recruitment', name: '录用审批' },
  // 薪酬管理
  { route: '/hrm/salary/setting/subject', module: 'salary', name: '薪酬主体' },
  { route: '/hrm/salary/setting/item', module: 'salary', name: '薪资项目' },
  { route: '/hrm/salary/setting/scheme', module: 'salary', name: '薪酬方案' },
  { route: '/hrm/salary/archive-list', module: 'salary', name: '员工薪资档案' },
  { route: '/hrm/salary/batch', module: 'salary', name: '薪资批次' },
  { route: '/hrm/salary/confirm-list', module: 'salary', name: '薪资确认审批' },
  { route: '/hrm/salary/payment-list', module: 'salary', name: '发薪登记' },
  { route: '/hrm/salary/payslip', module: 'salary', name: '工资条管理' },
  { route: '/hrm/salary/tax/tax-profile', module: 'salary', name: '员工个税档案' },
  { route: '/hrm/salary/tax/special-deduction', module: 'salary', name: '专项附加扣除' },
  { route: '/hrm/salary/tax/tax-opening', module: 'salary', name: '个税累计期初' },
  { route: '/hrm/salary/statistics', module: 'salary', name: '薪资统计' },
  // 培训管理
  { route: '/hrm/training/plan', module: 'training', name: '培训计划' },
  { route: '/hrm/training/session', module: 'training', name: '培训班次' },
  { route: '/hrm/training/cert-template', module: 'training', name: '证书模板' },
  { route: '/hrm/training/cert-issue', module: 'training', name: '证书发放' },
  { route: '/hrm/training/my', module: 'training', name: '我的培训' },
  { route: '/hrm/training/signin/manage', module: 'training', name: '培训签到管理' },
  { route: '/hrm/training/signin/scan', module: 'training', name: '扫码签到落地页' },
  // 奖惩管理
  { route: '/hrm/reward/bill', module: 'reward', name: '奖惩单' },
];
