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
    }
  }
}
