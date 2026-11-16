declare namespace Api {
  namespace Hrm {
    /**
     * namespace Performance
     *
     * backend api module: "Hrm.Performance"（绩效管理 — HRM-40 指标库与考核模板）
     */
    namespace Performance {
      /**
       * 指标类型
       *
       * - "kpi": 结果指标(KPI)
       * - "behavior": 行为能力
       */
      type IndicatorType = 'kpi' | 'behavior';

      /**
       * 评分规则
       *
       * - "formula": 公式评分
       * - "manual": 手工评分
       */
      type ScoreRule = 'formula' | 'manual';

      /**
       * 数据来源
       *
       * - "system": 系统取数
       * - "manual": 手工填报
       */
      type DataSource = 'system' | 'manual';

      /**
       * 启用状态（与 manage 域一致）
       *
       * - "1": 启用
       * - "2": 停用
       */
      type CommonStatus = Api.Common.EnableStatus;

      /** 指标库指标 */
      type Indicator = Common.CommonRecord<{
        /** 指标编码 */
        code: string;
        /** 指标名称 */
        name: string;
        /** 指标类型 */
        indicatorType: IndicatorType;
        /** 计量单位 */
        unit: string;
        /** 评分规则 */
        scoreRule: ScoreRule;
        /** 数据来源 */
        dataSource: DataSource;
        /** 默认权重（百分比） */
        defaultWeight: number;
        /** 备注 */
        remark: string;
      }>;

      /** 指标库 search params */
      type IndicatorSearchParams = CommonType.RecordNullable<
        Pick<Indicator, 'code' | 'name' | 'indicatorType' | 'scoreRule' | 'dataSource' | 'status'> &
          Common.CommonSearchParams
      >;

      /** 指标库 list */
      type IndicatorList = Common.PaginatingQueryRecord<Indicator>;

      type AddIndicatorRequest = {
        code: string;
        name: string;
        indicatorType: IndicatorType;
        unit: string;
        scoreRule: ScoreRule;
        dataSource: DataSource;
        defaultWeight: number;
        status: CommonStatus;
        remark: string;
      };

      type EditIndicatorRequest = AddIndicatorRequest & {
        uuid: string | undefined;
      };

      // eslint-disable-next-line @typescript-eslint/no-empty-object-type
      type AddIndicatorResponse = {};
      // eslint-disable-next-line @typescript-eslint/no-empty-object-type
      type EditIndicatorResponse = {};
      // eslint-disable-next-line @typescript-eslint/no-empty-object-type
      type DeleteIndicatorResponse = {};

      /** 考核模板-指标权重配置（提交用） */
      type TemplateItemRequest = {
        /** 指标 uuid */
        indicatorUuid: string;
        /** 权重（百分比），模板内权重之和须为 100% */
        weight: number;
      };

      /** 考核模板-指标权重配置（详情展示，带指标库快照） */
      type TemplateItem = {
        uuid: string;
        indicatorUuid: string;
        weight: number;
        indicatorCode: string;
        indicatorName: string;
        indicatorType: IndicatorType;
        unit: string;
        scoreRule: ScoreRule;
        dataSource: DataSource;
      };

      /** 考核模板 */
      type Template = Common.CommonRecord<{
        /** 模板名称 */
        name: string;
        /** 备注 */
        remark: string;
        /** 模板内指标权重配置（列表行为空数组，详情返回） */
        items: TemplateItem[];
      }>;

      /** 考核模板 search params */
      type TemplateSearchParams = CommonType.RecordNullable<
        Pick<Template, 'name' | 'status'> & Common.CommonSearchParams
      >;

      /** 考核模板 list */
      type TemplateList = Common.PaginatingQueryRecord<Template>;

      /** 考核模板详情（含指标权重配置） */
      type TemplateDetail = Template;

      type GetTemplateDetailRequest = {
        uuid: string;
      };

      type AddTemplateRequest = {
        name: string;
        status: CommonStatus;
        remark: string;
        items: TemplateItemRequest[];
      };

      type EditTemplateRequest = {
        uuid: string | undefined;
        name: string;
        status: CommonStatus;
        remark: string;
        items: TemplateItemRequest[];
      };

      // eslint-disable-next-line @typescript-eslint/no-empty-object-type
      type GetTemplateDetailResponse = {};
      // eslint-disable-next-line @typescript-eslint/no-empty-object-type
      type AddTemplateResponse = {};
      // eslint-disable-next-line @typescript-eslint/no-empty-object-type
      type EditTemplateResponse = {};
      // eslint-disable-next-line @typescript-eslint/no-empty-object-type
      type DeleteTemplateResponse = {};

      /**
       * 周期类型（HRM-41 绩效方案）
       *
       * - "month": 月度
       * - "quarter": 季度
       * - "half_year": 半年
       * - "year": 年度
       */
      type SchemePeriodType = 'month' | 'quarter' | 'half_year' | 'year';

      /**
       * 评估模式
       *
       * - "kpi": KPI
       * - "kpi_behavior": KPI+行为混合
       */
      type SchemeAssessmentMode = 'kpi' | 'kpi_behavior';

      /**
       * 评分模式
       *
       * - "percent": 百分制
       * - "grade": 等级制
       * - "amount": 金额制
       */
      type SchemeScoreMode = 'percent' | 'grade' | 'amount';

      /**
       * 联薪方式（HRM-49 绩效联薪契约字段，取值勿改动）
       *
       * - "coefficient": 系数制
       * - "amount": 金额制
       */
      type SchemePayLinkMode = 'coefficient' | 'amount';

      /** 绩效方案 */
      type Scheme = Common.CommonRecord<{
        /** 方案名称 */
        name: string;
        /** 关联考核模板 uuid */
        templateUuid: string;
        /** 关联考核模板名称（展示快照） */
        templateName: string;
        /** 周期类型 */
        periodType: SchemePeriodType;
        /** 评估模式 */
        assessmentMode: SchemeAssessmentMode;
        /** 评分模式 */
        scoreMode: SchemeScoreMode;
        /** 联薪方式 */
        payLinkMode: SchemePayLinkMode;
        /** 备注 */
        remark: string;
      }>;

      /** 绩效方案 search params */
      type SchemeSearchParams = CommonType.RecordNullable<Pick<Scheme, 'name' | 'status'> & Common.CommonSearchParams>;

      /** 绩效方案 list */
      type SchemeList = Common.PaginatingQueryRecord<Scheme>;

      type AddSchemeRequest = {
        name: string;
        templateUuid: string;
        periodType: SchemePeriodType;
        assessmentMode: SchemeAssessmentMode;
        scoreMode: SchemeScoreMode;
        payLinkMode: SchemePayLinkMode;
        status: CommonStatus;
        remark: string;
      };

      type EditSchemeRequest = AddSchemeRequest & {
        uuid: string | undefined;
      };

      // eslint-disable-next-line @typescript-eslint/no-empty-object-type
      type AddSchemeResponse = {};
      // eslint-disable-next-line @typescript-eslint/no-empty-object-type
      type EditSchemeResponse = {};
      // eslint-disable-next-line @typescript-eslint/no-empty-object-type
      type DeleteSchemeResponse = {};

      // ---- 绩效批次（HRM-42，演示站 9.2 考核批次工作台） ----

      /**
       * 批次类型
       *
       * - "month": 月度
       * - "quarter": 季度
       */
      type PeriodType = 'month' | 'quarter';

      /**
       * 批次生命周期（GLOSSARY「绩效批次」，forward-only）
       *
       * - "0": 草稿
       * - "1": 进行中
       * - "2": 待确认
       * - "3": 待校准
       * - "4": 结果已发布
       * - "5": 已发布
       */
      type PeriodStatus = '0' | '1' | '2' | '3' | '4' | '5';

      /** 绩效结果状态：1=正常 2=已作废 */
      type PerfResultStatus = '1' | '2';

      /** 是/否 0|1（绩效结果行布尔列） */
      type PerfYesNo = '0' | '1';

      /** 适用范围多维筛选条件（存后端 scope JSON） */
      type PeriodScope = {
        /** 公司（org_type=1）uuids；空 = 不限 */
        companyUuids: string[];
        /** 部门 uuids（含全部下级组织）；空 = 不限 */
        orgUuids: string[];
        /** 职位；空 = 不限 */
        positions: string[];
        /** 职务；空 = 不限 */
        posts: string[];
        /** 人员状态（'1'-'6'）；空 = 不限 */
        personStatuses: string[];
      };

      /** 绩效批次 */
      type PerfPeriod = {
        uuid: string;
        /** 批次编号 PERF-年-流水 */
        periodNo: string;
        /** 批次名称 */
        name: string;
        /** 批次类型 */
        periodType: PeriodType;
        /** 绩效方案 uuid */
        schemeUuid: string;
        /** 绩效方案名称（展示快照） */
        schemeName: string;
        /** 适用范围 JSON 字符串（PeriodScope 序列化） */
        scope: string;
        /** 周期开始 YYYY-MM-DD */
        periodStart: string;
        /** 周期结束 YYYY-MM-DD */
        periodEnd: string;
        /** 结果发布日期 YYYY-MM-DD */
        resultPublishDate: string;
        /** 参与人数（快照） */
        participantCount: number;
        /** 生命周期 */
        status: PeriodStatus;
        remark: string;
        createTime: string;
        updateTime: string;
      };

      /** 绩效批次 search params */
      type PerfPeriodSearchParams = CommonType.RecordNullable<
        Pick<PerfPeriod, 'name' | 'status'> & Common.CommonSearchParams
      >;

      /** 绩效批次 list */
      type PerfPeriodList = Common.PaginatingQueryRecord<PerfPeriod>;

      type AddPerfPeriodRequest = {
        name: string;
        periodType: PeriodType;
        schemeUuid: string;
        scope: string;
        periodStart: string;
        periodEnd: string;
        resultPublishDate: string;
        remark: string;
      };

      type EditPerfPeriodRequest = AddPerfPeriodRequest & {
        uuid: string | undefined;
      };

      // eslint-disable-next-line @typescript-eslint/no-empty-object-type
      type AddPerfPeriodResponse = {};
      // eslint-disable-next-line @typescript-eslint/no-empty-object-type
      type EditPerfPeriodResponse = {};
      // eslint-disable-next-line @typescript-eslint/no-empty-object-type
      type DeletePerfPeriodResponse = {};
      // eslint-disable-next-line @typescript-eslint/no-empty-object-type
      type PublishPerfPeriodResponse = {};
      // eslint-disable-next-line @typescript-eslint/no-empty-object-type
      type AdvancePerfPeriodResponse = {};

      // ---- 考核任务链（HRM-43，演示站 9.2 五个填报列表页共用） ----

      /**
       * 考核任务当前节点（GLOSSARY「考核任务」五阶段）
       *
       * - "1": 目标确认（被考核人）
       * - "2": 自评填报（被考核人）
       * - "3": 上级评审（有 review 码者）
       * - "4": 校准填报（校准人）
       * - "5": 结果确认（被考核人）
       */
      type AssessmentNode = '1' | '2' | '3' | '4' | '5';

      /**
       * 考核任务状态
       *
       * - "0": 进行中
       * - "1": 已完成
       * - "2": 申诉中
       */
      type AssessmentTaskStatus = '0' | '1' | '2';

      /**
       * 申诉状态
       *
       * - "0": 待处理
       * - "1": 已裁决-维持原结果
       * - "2": 已裁决-调整得分
       */
      type AssessmentAppealStatus = '0' | '1' | '2';

      /** 考核任务-模板快照项（items+权重，生成任务时固化） */
      type AssessmentItemSnapshot = {
        indicatorUuid: string;
        indicatorCode: string;
        indicatorName: string;
        indicatorType: IndicatorType;
        unit: string;
        scoreRule: ScoreRule;
        dataSource: DataSource;
        weight: number;
      };

      /** 逐项得分（自评/评审提交项） */
      type ScoreItem = {
        indicatorUuid: string;
        score: number;
      };

      /** 考核任务 */
      type PerfAssessment = {
        uuid: string;
        /** 考核单号 PA-年-流水 */
        assessmentNo: string;
        periodUuid: string;
        /** 批次编号（展示快照） */
        periodNo: string;
        /** 批次名称（展示快照） */
        periodName: string;
        employeeUuid: string;
        /** 被考核人姓名（展示快照） */
        employeeName: string;
        templateUuid: string;
        /** 考核模板名称（展示快照） */
        templateName: string;
        /** 模板快照 items（列表行为空数组，详情返回） */
        items: AssessmentItemSnapshot[];
        /** 当前节点 */
        currentNode: AssessmentNode;
        /** 任务状态 */
        taskStatus: AssessmentTaskStatus;
        /** 最终得分（校准后） */
        finalScore: number;
        /** 最终等级 */
        finalGrade: string;
        selfScore: number;
        selfGrade: string;
        reviewScore: number;
        reviewGrade: string;
        /** 申诉状态（''=无申诉） */
        appealStatus: '' | AssessmentAppealStatus;
        createTime: string;
        updateTime: string;
      };

      /** 考核任务 search params */
      type AssessmentSearchParams = CommonType.RecordNullable<
        Pick<PerfAssessment, 'periodUuid' | 'employeeUuid' | 'currentNode' | 'taskStatus'> &
          Common.CommonSearchParams
      >;

      /** 考核任务 list */
      type AssessmentList = Common.PaginatingQueryRecord<PerfAssessment>;

      /** 考核任务详情 */
      type AssessmentDetail = PerfAssessment & {
        planTime: string;
        planBy: string;
        selfItems: ScoreItem[];
        selfComment: string;
        selfTime: string;
        selfBy: string;
        reviewItems: ScoreItem[];
        reviewComment: string;
        reviewTime: string;
        reviewBy: string;
        calibComment: string;
        calibTime: string;
        calibBy: string;
        resultTime: string;
        resultBy: string;
      };

      // eslint-disable-next-line @typescript-eslint/no-empty-object-type
      type ConfirmPlanResponse = {};
      // eslint-disable-next-line @typescript-eslint/no-empty-object-type
      type SubmitSelfReviewResponse = {};
      // eslint-disable-next-line @typescript-eslint/no-empty-object-type
      type SubmitLeaderReviewResponse = {};
      // eslint-disable-next-line @typescript-eslint/no-empty-object-type
      type SubmitCalibrationResponse = {};
      // eslint-disable-next-line @typescript-eslint/no-empty-object-type
      type ConfirmResultResponse = {};
      // eslint-disable-next-line @typescript-eslint/no-empty-object-type
      type AppealResponse = {};
      // eslint-disable-next-line @typescript-eslint/no-empty-object-type
      type HandleAppealResponse = {};

      type HandleAppealRequest = {
        uuid: string;
        /** 裁决："1"=维持原结果 "2"=调整得分 */
        decision: AssessmentAppealStatus;
        comment: string;
        adjustedScore?: number;
        adjustedGrade?: string;
      };

      // ---- 考核任务工作台（HRM-44，单聚合端点） ----

      /** 工作台待办我的角色 */
      type WorkbenchTodoRole = 'owner' | 'reviewer' | 'calibrator';

      /** 工作台待办行（最近待办 / 申诉单列共用） */
      type WorkbenchTodo = {
        uuid: string;
        /** 考核单号 PA-年-流水 */
        assessmentNo: string;
        periodUuid: string;
        /** 批次编号（展示快照） */
        periodNo: string;
        /** 批次名称（展示快照） */
        periodName: string;
        employeeUuid: string;
        /** 被考核人姓名（展示快照） */
        employeeName: string;
        /** 我的角色 owner=被考核人 reviewer=评审 calibrator=校准/裁决 */
        myRole: WorkbenchTodoRole;
        /** 当前节点 */
        currentNode: AssessmentNode;
        /** 任务状态 */
        taskStatus: AssessmentTaskStatus;
        /** 申诉状态（申诉中任务携带） */
        appealStatus: '' | AssessmentAppealStatus;
        updateTime: string;
      };

      /** 工作台汇总卡（六卡，按当前用户角色口径统计待办任务数） */
      type WorkbenchCards = {
        /** 目标确认待办（被考核人，节点1） */
        planConfirm: number;
        /** 自评填报待办（被考核人，节点2） */
        selfReview: number;
        /** 上级评审待办（考核人，节点3） */
        leaderReview: number;
        /** 校准填报待办（校准人，节点4） */
        calibration: number;
        /** 结果确认待办（被考核人，节点5） */
        resultConfirm: number;
        /** 申诉处理待办（申诉中） */
        appeal: number;
      };

      /** 考核任务工作台聚合响应 */
      type AssessmentWorkbench = {
        cards: WorkbenchCards;
        /** 最近待办列表（更新时间倒序，默认前 20） */
        todos: WorkbenchTodo[];
        /** 申诉中的任务单列 */
        appeals: WorkbenchTodo[];
      };

      // ---- 批次结果审批与绩效结果（HRM-45，演示站 9.3 结果） ----

      /** HR211 批次结果审批单 */
      type PerfResultApproval = {
        uuid: string;
        /** 审批单号 HR211-{YYYYMMDD}{5位流水} */
        billNo: string;
        /** 绩效批次 uuid */
        periodUuid: string;
        /** 批次编号快照 */
        periodNo: string;
        /** 批次名称快照 */
        periodName: string;
        /** 公司名称快照（JSON 数组字符串） */
        companySnapshot: string;
        /** 部门名称快照（JSON 数组字符串） */
        orgSnapshot: string;
        /** 人数 */
        participantCount: number;
        /** 平均得分 */
        avgScore: number;
        /** 绩效金额合计 */
        totalAmount: number;
        /** 单据状态（共享单据状态机） */
        status: Api.Hrm.BillStatus;
        submitBy: string;
        submitTime: string;
        approveBy: string;
        approveTime: string;
        opinion: string;
        createTime: string;
        updateTime: string;
      };

      /** 批次结果审批 search params */
      type PerfResultApprovalSearchParams = CommonType.RecordNullable<
        Pick<PerfResultApproval, 'periodUuid' | 'status'> & Common.CommonSearchParams
      >;

      /** 批次结果审批 list */
      type PerfResultApprovalList = Common.PaginatingQueryRecord<PerfResultApproval>;

      type SubmitPerfResultApprovalRequest = {
        periodUuid: string;
        opinion?: string;
      };

      // eslint-disable-next-line @typescript-eslint/no-empty-object-type
      type SubmitPerfResultApprovalResponse = {};
      // eslint-disable-next-line @typescript-eslint/no-empty-object-type
      type SignPerfResultResponse = {};

      /** 绩效结果行 */
      type PerfResult = {
        uuid: string;
        /** 审批单 uuid */
        approvalUuid: string;
        /** 考核任务 uuid */
        assessmentUuid: string;
        /** 绩效批次 uuid */
        periodUuid: string;
        /** 批次编号快照 */
        periodNo: string;
        /** 员工 uuid */
        employeeUuid: string;
        /** 员工姓名快照 */
        employeeName: string;
        /** 公司 uuid 快照 */
        companyUuid: string;
        /** 公司名称快照 */
        companyName: string;
        /** 部门 uuid 快照 */
        orgUuid: string;
        /** 部门名称快照 */
        orgName: string;
        /** 得分 */
        score: number;
        /** 等级（A/B/C/D 或等级制原值） */
        grade: string;
        /** 绩效系数 */
        coefficient: number;
        /** 绩效金额 */
        perfAmount: number;
        /** 奖金金额 */
        bonusAmount: number;
        /** 结果状态：1=正常 2=已作废 */
        resultStatus: PerfResultStatus;
        /** 已发布：0=否 1=是 */
        published: PerfYesNo;
        /** 已签收：0=否 1=是 */
        signed: PerfYesNo;
        /** 签收时间 */
        signTime: string;
        /** 签收人 */
        signBy: string;
        /** 已同步薪酬：0=否 1=是（HRM-49 写入） */
        salarySynced: PerfYesNo;
        createTime: string;
        updateTime: string;
      };

      /** 绩效结果行 search params（HRM-46 增补 公司/部门快照 + 已同步薪酬筛选） */
      type PerfResultSearchParams = CommonType.RecordNullable<
        Pick<PerfResult, 'periodUuid' | 'employeeUuid' | 'signed' | 'companyUuid' | 'orgUuid' | 'salarySynced'> & {
          mine?: boolean;
        } & Common.CommonSearchParams
      >;

      /** 绩效结果 list */
      type PerfResultList = Common.PaginatingQueryRecord<PerfResult>;

      // ---- 绩效驾驶舱（HRM-46，演示站 9.3） ----
      // 口径（pre-decided，见 hrm-46-notes.md）：进行中批次数=批次 status='1'；
      // 考核完成率=task_status='1'/总任务；逾期待办=批次 period_end<今天 且任务未完成；
      // 风险人数=已发布结果 score<低分阈值 去重员工数；平均得分=已发布结果行均值；
      // 申诉中=task_status='2'。

      /** 驾驶舱请求（公司/部门/批次/低分阈值筛选） */
      type PerfDashboardSearchParams = {
        companyUuid?: string;
        orgUuid?: string;
        periodUuid?: string;
        lowScore?: number;
      };

      /** 驾驶舱指标卡 */
      type PerfDashboardMetrics = {
        /** 进行中批次数 */
        ongoingPeriodCount: number;
        /** 考核完成率（0-100） */
        completionRate: number;
        /** 逾期待办数（任务级） */
        overdueTaskCount: number;
        /** 风险人数（score<阈值 去重员工） */
        riskEmployeeCount: number;
        /** 平均得分 */
        avgScore: number;
        /** 申诉中数 */
        appealCount: number;
        /** 生效低分阈值（回显） */
        lowScore: number;
      };

      /** 低分明细行（风险预警 Tab1；signed 后端为 int64 0/1） */
      type PerfDashboardLowScore = Omit<
        Pick<PerfResult, 'periodUuid' | 'periodNo' | 'employeeUuid' | 'employeeName' | 'companyName' | 'orgName' | 'grade' | 'signed'>,
        'signed'
      > & {
        score: number;
        coefficient: number;
        perfAmount: number;
        signed: number;
      };

      /** 逾期任务明细行（风险预警 Tab2） */
      type PerfDashboardOverdueTask = {
        periodUuid: string;
        periodNo: string;
        periodEnd: string;
        employeeUuid: string;
        employeeName: string;
        currentNode: string;
        taskStatus: string;
        assessmentNo: string;
      };

      /** 申诉中名单行（简化"申诉积压"） */
      type PerfDashboardAppeal = {
        periodUuid: string;
        periodNo: string;
        employeeUuid: string;
        employeeName: string;
        assessmentNo: string;
        currentNode: string;
      };

      /** 驾驶舱聚合响应（单端点） */
      type PerfDashboardResponse = {
        metrics: PerfDashboardMetrics;
        /** 风险预警 Tab1：低分明细 */
        lowScores: PerfDashboardLowScore[];
        /** 风险预警 Tab2：逾期任务明细 */
        overdueTasks: PerfDashboardOverdueTask[];
        /** 申诉中名单 */
        appeals: PerfDashboardAppeal[];
      };
    }
  }
}
