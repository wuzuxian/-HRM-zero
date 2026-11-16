<script setup lang="tsx">
import { computed, onMounted, ref } from 'vue';
import { NTag } from 'naive-ui';
import { GetOrgTree, GetPerfDashboard, GetPerfPeriodList } from '@/service/api';
import { $t } from '@/locales';
import {
  perfAssessmentNodeRecord,
  perfAssessmentNodeTag,
  perfAssessmentTaskStatusRecord
} from '@/constants/hrm-performance';

defineOptions({
  name: 'PerfDashboard'
});

// 绩效驾驶舱（HRM-46，演示站 9.3 /hrm/performance/dashboard）。
// 单聚合端点 GetPerfDashboard：指标卡（进行中批次/考核完成率/逾期待办/风险人数/
// 平均得分/申诉中）+ 风险预警明细（Tab1 低分明细 / Tab2 逾期任务明细 + 申诉中名单列，
// "连续低分/申诉积压"简化为申诉中名单）。

const loading = ref(false);

const metrics = ref<Api.Hrm.Performance.PerfDashboardMetrics>({
  ongoingPeriodCount: 0,
  completionRate: 100,
  overdueTaskCount: 0,
  riskEmployeeCount: 0,
  avgScore: 0,
  appealCount: 0,
  lowScore: 60
});

const lowScores = ref<Api.Hrm.Performance.PerfDashboardLowScore[]>([]);
const overdueTasks = ref<Api.Hrm.Performance.PerfDashboardOverdueTask[]>([]);
const appeals = ref<Api.Hrm.Performance.PerfDashboardAppeal[]>([]);

/** 筛选（公司/部门/批次/低分阈值） */
const searchParams = ref<Api.Hrm.Performance.PerfDashboardSearchParams>({
  companyUuid: undefined,
  orgUuid: undefined,
  periodUuid: undefined,
  lowScore: undefined
});

const activeTab = ref<'lowScores' | 'overdueTasks'>('lowScores');

// ---- 指标卡 ----

const metricCardItems = computed(() => [
  { label: $t('page.hrm.performance.dashboard.metric.ongoingPeriods'), value: metrics.value.ongoingPeriodCount },
  { label: $t('page.hrm.performance.dashboard.metric.completionRate'), value: `${metrics.value.completionRate}%` },
  { label: $t('page.hrm.performance.dashboard.metric.overdueTasks'), value: metrics.value.overdueTaskCount },
  { label: $t('page.hrm.performance.dashboard.metric.riskEmployees'), value: metrics.value.riskEmployeeCount },
  { label: $t('page.hrm.performance.dashboard.metric.avgScore'), value: metrics.value.avgScore },
  { label: $t('page.hrm.performance.dashboard.metric.appeals'), value: metrics.value.appealCount }
]);

// ---- 筛选下拉（组织树 + 批次） ----

const orgTreeData = ref<Api.Hrm.Org[]>([]);

async function getOrgTree() {
  const { error, data } = await GetOrgTree();
  if (!error && data) {
    orgTreeData.value = data.records;
  }
}

const companyOptions = computed<CommonType.Option<string>[]>(() =>
  orgTreeData.value
    .filter(item => item.orgType === '1')
    .map(item => ({ label: item.orgName, value: item.uuid }))
);

const orgOptions = computed<CommonType.Option<string>[]>(() => {
  const out: CommonType.Option<string>[] = [];
  function walk(node: Api.Hrm.Org) {
    if (node.orgType === '2') {
      out.push({ label: node.orgName, value: node.uuid });
    }
    (node.children ?? []).forEach(walk);
  }
  orgTreeData.value.forEach(walk);
  return out;
});

const periodOptions = ref<CommonType.Option<string>[]>([]);

async function getPeriodOptions() {
  const { error, data } = await GetPerfPeriodList({ current: 1, size: 1000 });
  if (!error && data) {
    periodOptions.value = data.records.map(item => ({
      label: `${item.periodNo} ${item.name}`,
      value: item.uuid
    }));
  }
}

// ---- 数据加载 ----

async function getDashboard() {
  loading.value = true;
  const { error, data } = await GetPerfDashboard({
    companyUuid: searchParams.value.companyUuid ?? undefined,
    orgUuid: searchParams.value.orgUuid ?? undefined,
    periodUuid: searchParams.value.periodUuid ?? undefined,
    lowScore: searchParams.value.lowScore ?? undefined
  });
  loading.value = false;
  if (!error && data) {
    metrics.value = data.metrics;
    lowScores.value = data.lowScores;
    overdueTasks.value = data.overdueTasks;
    appeals.value = data.appeals;
  }
}

function handleSearch() {
  getDashboard();
}

function handleReset() {
  searchParams.value = { companyUuid: undefined, orgUuid: undefined, periodUuid: undefined, lowScore: undefined };
  getDashboard();
}

// ---- 明细表列 ----

const lowScoreColumns = computed(() => [
  { key: 'index', title: $t('common.index'), align: 'center' as const, width: 60 },
  { key: 'periodNo', title: $t('page.hrm.performance.dashboard.period'), align: 'center' as const, width: 150 },
  { key: 'employeeName', title: $t('page.hrm.performance.dashboard.employee'), align: 'center' as const, width: 100 },
  { key: 'companyName', title: $t('page.hrm.performance.dashboard.company'), align: 'center' as const, width: 110 },
  { key: 'orgName', title: $t('page.hrm.performance.dashboard.org'), align: 'center' as const, width: 110 },
  {
    key: 'score',
    title: $t('page.hrm.performance.dashboard.score'),
    align: 'center' as const,
    width: 80,
    render: (row: Api.Hrm.Performance.PerfDashboardLowScore) => (
      <NTag type='error' size='small'>
        {row.score}
      </NTag>
    )
  },
  { key: 'grade', title: $t('page.hrm.performance.dashboard.grade'), align: 'center' as const, width: 70 },
  { key: 'coefficient', title: $t('page.hrm.performance.result.coefficient'), align: 'center' as const, width: 80 },
  { key: 'perfAmount', title: $t('page.hrm.performance.result.perfAmount'), align: 'center' as const, width: 110 },
  {
    key: 'signed',
    title: $t('page.hrm.performance.result.signed'),
    align: 'center' as const,
    width: 80,
    render: (row: Api.Hrm.Performance.PerfDashboardLowScore) =>
      row.signed === 1 ? (
        <NTag type='success' size='small'>
          {$t('page.hrm.performance.result.yes')}
        </NTag>
      ) : (
        <NTag type='default' size='small'>
          {$t('page.hrm.performance.result.no')}
        </NTag>
      )
  }
]);

const overdueColumns = computed(() => [
  { key: 'index', title: $t('common.index'), align: 'center' as const, width: 60 },
  { key: 'periodNo', title: $t('page.hrm.performance.dashboard.period'), align: 'center' as const, width: 150 },
  {
    key: 'periodEnd',
    title: $t('page.hrm.performance.dashboard.periodEnd'),
    align: 'center' as const,
    width: 110,
    render: (row: Api.Hrm.Performance.PerfDashboardOverdueTask) => (
      <NTag type='warning' size='small'>
        {row.periodEnd}
      </NTag>
    )
  },
  { key: 'assessmentNo', title: $t('page.hrm.performance.dashboard.assessmentNo'), align: 'center' as const, width: 170 },
  { key: 'employeeName', title: $t('page.hrm.performance.dashboard.employee'), align: 'center' as const, width: 100 },
  {
    key: 'currentNode',
    title: $t('page.hrm.performance.dashboard.currentNode'),
    align: 'center' as const,
    width: 110,
    render: (row: Api.Hrm.Performance.PerfDashboardOverdueTask) => {
      const record = perfAssessmentNodeRecord as Record<string, string>;
      const key = record[row.currentNode];
      return key ? $t(key as App.I18n.I18nKey) : row.currentNode;
    }
  },
  {
    key: 'taskStatus',
    title: $t('page.hrm.performance.dashboard.taskStatus'),
    align: 'center' as const,
    width: 90,
    render: (row: Api.Hrm.Performance.PerfDashboardOverdueTask) => {
      const record = perfAssessmentTaskStatusRecord as Record<string, string>;
      const key = record[row.taskStatus];
      return key ? $t(key as App.I18n.I18nKey) : row.taskStatus;
    }
  }
]);

const appealColumns = computed(() => [
  { key: 'index', title: $t('common.index'), align: 'center' as const, width: 60 },
  { key: 'periodNo', title: $t('page.hrm.performance.dashboard.period'), align: 'center' as const, width: 150 },
  { key: 'assessmentNo', title: $t('page.hrm.performance.dashboard.assessmentNo'), align: 'center' as const, width: 170 },
  { key: 'employeeName', title: $t('page.hrm.performance.dashboard.employee'), align: 'center' as const, width: 100 },
  {
    key: 'currentNode',
    title: $t('page.hrm.performance.dashboard.currentNode'),
    align: 'center' as const,
    width: 110,
    render: (row: Api.Hrm.Performance.PerfDashboardAppeal) => {
      const record = perfAssessmentNodeRecord as Record<string, string>;
      const key = record[row.currentNode];
      return key ? $t(key as App.I18n.I18nKey) : row.currentNode;
    }
  }
]);

onMounted(() => {
  getOrgTree();
  getPeriodOptions();
  getDashboard();
});
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <!-- 筛选卡：公司/部门/批次/低分阈值 -->
    <NCard :bordered="false" size="small" class="card-wrapper">
      <NForm label-placement="left" :label-width="90">
        <NGrid responsive="screen" item-responsive>
          <NFormItemGi
            span="24 s:12 m:6"
            :label="$t('page.hrm.performance.dashboard.company')"
            class="pr-24px"
          >
            <NSelect
              v-model:value="searchParams.companyUuid"
              :options="companyOptions"
              :placeholder="$t('page.hrm.performance.dashboard.companyFilter')"
              clearable
              filterable
            />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:6"
            :label="$t('page.hrm.performance.dashboard.org')"
            class="pr-24px"
          >
            <NSelect
              v-model:value="searchParams.orgUuid"
              :options="orgOptions"
              :placeholder="$t('page.hrm.performance.dashboard.orgFilter')"
              clearable
              filterable
            />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:6"
            :label="$t('page.hrm.performance.dashboard.period')"
            class="pr-24px"
          >
            <NSelect
              v-model:value="searchParams.periodUuid"
              :options="periodOptions"
              :placeholder="$t('page.hrm.performance.dashboard.periodFilter')"
              clearable
              filterable
            />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:6"
            :label="$t('page.hrm.performance.dashboard.lowScore')"
            class="pr-24px"
          >
            <NInputNumber
              v-model:value="searchParams.lowScore"
              :min="0"
              :max="100"
              :placeholder="$t('page.hrm.performance.dashboard.lowScoreFilter')"
              clearable
            />
          </NFormItemGi>
          <NFormItemGi span="24 m:6" class="pr-24px">
            <NSpace class="w-full" justify="end">
              <NButton @click="handleReset">
                <template #icon>
                  <icon-ic-round-refresh class="text-icon" />
                </template>
                {{ $t('common.reset') }}
              </NButton>
              <NButton type="primary" ghost @click="handleSearch">
                <template #icon>
                  <icon-ic-round-search class="text-icon" />
                </template>
                {{ $t('common.search') }}
              </NButton>
            </NSpace>
          </NFormItemGi>
        </NGrid>
      </NForm>
    </NCard>

    <!-- 指标卡 -->
    <NCard :bordered="false" size="small" class="card-wrapper">
      <NGrid cols="2 s:3 m:6" responsive="screen" :x-gap="8" :y-gap="8">
        <NGi v-for="item in metricCardItems" :key="item.label">
          <NStatistic :label="item.label" :value="item.value" />
        </NGi>
      </NGrid>
    </NCard>

    <!-- 风险预警明细：Tab1 低分明细 / Tab2 逾期任务明细；申诉中名单并列展示 -->
    <NCard
      :title="$t('page.hrm.performance.dashboard.riskTitle')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <NTabs v-model:value="activeTab" type="line" animated>
        <NTabPane name="lowScores" :tab="$t('page.hrm.performance.dashboard.tabLowScores')">
          <NDataTable
            :columns="lowScoreColumns"
            :data="lowScores"
            size="small"
            :loading="loading"
            :max-height="360"
            :row-key="(row: Api.Hrm.Performance.PerfDashboardLowScore) => `${row.periodUuid}:${row.employeeUuid}`"
            :scroll-x="1000"
          />
        </NTabPane>
        <NTabPane name="overdueTasks" :tab="$t('page.hrm.performance.dashboard.tabOverdueTasks')">
          <NDataTable
            :columns="overdueColumns"
            :data="overdueTasks"
            size="small"
            :loading="loading"
            :max-height="360"
            :row-key="(row: Api.Hrm.Performance.PerfDashboardOverdueTask) => row.assessmentNo"
            :scroll-x="900"
          />
        </NTabPane>
      </NTabs>
    </NCard>

    <!-- 申诉中名单（简化"申诉积压"口径：task_status='2' 名单列） -->
    <NCard
      :title="$t('page.hrm.performance.dashboard.appealTitle')"
      :bordered="false"
      size="small"
      class="card-wrapper"
    >
      <NDataTable
        :columns="appealColumns"
        :data="appeals"
        size="small"
        :loading="loading"
        :max-height="280"
        :row-key="(row: Api.Hrm.Performance.PerfDashboardAppeal) => row.assessmentNo"
      />
    </NCard>
  </div>
</template>

<style scoped></style>
