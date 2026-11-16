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
    }
  }
}
