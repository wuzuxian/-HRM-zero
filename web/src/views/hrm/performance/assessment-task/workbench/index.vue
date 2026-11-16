<script setup lang="tsx">
import { computed, onMounted, ref } from 'vue';
import { NButton, NCard, NDataTable, NGi, NGrid, NStatistic, NTag } from 'naive-ui';
import type { DataTableColumns } from 'naive-ui';
import { GetMyAssessmentWorkbench } from '@/service/api';
import { $t } from '@/locales';
import { useRouterPush } from '@/hooks/common/router';
import { useAuth } from '@/hooks/business/auth';
import {
  perfAssessmentAppealStatusRecord,
  perfAssessmentAppealStatusTag,
  perfAssessmentNodeRecord,
  perfAssessmentNodeTag,
  perfAssessmentTaskStatusRecord,
  perfAssessmentTaskStatusTag
} from '@/constants/hrm-performance';

defineOptions({
  name: 'AssessmentTaskWorkbench'
});

/**
 * 考核任务工作台（HRM-44，demo-inventory 9.2 /hrm/performance/assessment-task/workbench）：
 * 当前登录人视角的只读聚合页 —
 *   - 全周期待办汇总卡（六卡：目标确认/自评填报/上级评审/校准填报/结果确认/申诉），
 *     数值=当前用户在各节点的实际待办任务数；卡片点击跳转对应填报列表页。
 *   - 最近待办列表（类型/批次/员工/当前节点/当前处理人/操作，更新时间倒序前 20）。
 *   - 申诉中的任务单列（按节点跳转，可发起申诉/裁决）。
 */

const { routerPushByKey } = useRouterPush();
const { hasAuth } = useAuth();

// ---- 数据加载 ----
const loading = ref(false);
const workbench = ref<Api.Hrm.Performance.AssessmentWorkbench | null>(null);

async function loadData() {
  loading.value = true;
  const { data, error } = await GetMyAssessmentWorkbench();
  loading.value = false;
  if (error) return;
  workbench.value = data;
}

onMounted(() => {
  loadData();
});

// ---- 汇总卡（点击跳转对应填报列表页） ----
const cardItems = [
  { key: 'planConfirm', routeKey: 'hrm_performance_assessment-task_plan-list' },
  { key: 'selfReview', routeKey: 'hrm_performance_assessment-task_self-review-list' },
  {
    key: 'leaderReview',
    routeKey: 'hrm_performance_assessment-task_leader-review-list',
    authCode: 'hrm:perf_assessment:review'
  },
  {
    key: 'calibration',
    routeKey: 'hrm_performance_assessment-task_leader-review-list',
    authCode: 'hrm:perf_assessment:calibrate'
  },
  { key: 'resultConfirm', routeKey: 'hrm_performance_assessment-task_result-confirm-task-list' },
  { key: 'appeal', routeKey: 'hrm_performance_assessment-task_result-confirm-task-list' }
] as const;

type CardKey = (typeof cardItems)[number]['key'];
type CardItem = Extract<(typeof cardItems)[number], { key: CardKey }>;

const cardValues = computed(() => workbench.value?.cards);

function cardLabel(key: CardKey) {
  return $t(`page.hrm.performance.workbench.${key}` as App.I18n.I18nKey);
}

function cardClickable(item: CardItem) {
  if ('authCode' in item && item.authCode && !hasAuth(item.authCode)) return false;
  return true;
}

function goCard(item: CardItem) {
  if (!cardClickable(item)) return;
  routerPushByKey(item.routeKey);
}

// ---- 待办表列（最近待办 / 申诉单列共用） ----
function renderNodeTag(node: Api.Hrm.Performance.AssessmentNode) {
  return (
    <NTag type={perfAssessmentNodeTag(node)} size="small">
      {$t(perfAssessmentNodeRecord[node])}
    </NTag>
  );
}

function renderStatusTag(status: Api.Hrm.Performance.AssessmentTaskStatus) {
  return (
    <NTag type={perfAssessmentTaskStatusTag(status)} size="small">
      {$t(perfAssessmentTaskStatusRecord[status])}
    </NTag>
  );
}

function renderAppealTag(status: Api.Hrm.Performance.WorkbenchTodo['appealStatus']) {
  if (!status) return null;
  return (
    <NTag type={perfAssessmentAppealStatusTag(status)} size="small">
      {$t(perfAssessmentAppealStatusRecord[status])}
    </NTag>
  );
}

/** 当前处理人：按节点映射处理角色（本人待办显示「我」） */
function handlerLabel(row: Api.Hrm.Performance.WorkbenchTodo) {
  if (row.myRole === 'owner') return $t('page.hrm.performance.workbench.handlerMe');
  if (row.myRole === 'reviewer') return $t('page.hrm.performance.workbench.handlerReviewer');
  return $t('page.hrm.performance.workbench.handlerCalibrator');
}

/** 按节点跳到对应填报页 */
function stageRouteKey(row: Api.Hrm.Performance.WorkbenchTodo) {
  switch (row.currentNode) {
    case '1':
      return 'hrm_performance_assessment-task_plan-list';
    case '2':
      return 'hrm_performance_assessment-task_self-review-list';
    case '3':
    case '4':
      return 'hrm_performance_assessment-task_leader-review-list';
    default:
      return 'hrm_performance_assessment-task_result-confirm-task-list';
  }
}

function goTodo(row: Api.Hrm.Performance.WorkbenchTodo) {
  routerPushByKey(stageRouteKey(row));
}

const todoColumns: DataTableColumns<Api.Hrm.Performance.WorkbenchTodo> = [
  {
    key: 'type',
    title: $t('page.hrm.performance.workbench.colType'),
    align: 'center',
    width: 120,
    render: row => renderNodeTag(row.currentNode)
  },
  {
    key: 'assessmentNo',
    title: $t('page.hrm.performance.workbench.colAssessmentNo'),
    align: 'center',
    width: 150
  },
  {
    key: 'periodName',
    title: $t('page.hrm.performance.workbench.colPeriod'),
    align: 'center',
    width: 160,
    render: row => row.periodName || row.periodNo || '-'
  },
  {
    key: 'employeeName',
    title: $t('page.hrm.performance.workbench.colEmployee'),
    align: 'center',
    width: 110,
    render: row => row.employeeName || '-'
  },
  {
    key: 'currentNode',
    title: $t('page.hrm.performance.workbench.colNode'),
    align: 'center',
    width: 110,
    render: row => renderNodeTag(row.currentNode)
  },
  {
    key: 'taskStatus',
    title: $t('page.hrm.performance.workbench.colStatus'),
    align: 'center',
    width: 100,
    render: row => renderStatusTag(row.taskStatus)
  },
  {
    key: 'handler',
    title: $t('page.hrm.performance.workbench.colHandler'),
    align: 'center',
    width: 110,
    render: row => handlerLabel(row)
  },
  {
    key: 'operate',
    title: $t('page.hrm.performance.workbench.colOperate'),
    align: 'center',
    width: 100,
    render: row => (
      <NButton size="small" type="primary" ghost onClick={() => goTodo(row)}>
        {$t('page.hrm.performance.workbench.goHandle')}
      </NButton>
    )
  }
];

const appealColumns: DataTableColumns<Api.Hrm.Performance.WorkbenchTodo> = [
  {
    key: 'assessmentNo',
    title: $t('page.hrm.performance.workbench.colAssessmentNo'),
    align: 'center',
    width: 150
  },
  {
    key: 'periodName',
    title: $t('page.hrm.performance.workbench.colPeriod'),
    align: 'center',
    width: 160,
    render: row => row.periodName || row.periodNo || '-'
  },
  {
    key: 'employeeName',
    title: $t('page.hrm.performance.workbench.colEmployee'),
    align: 'center',
    width: 110,
    render: row => row.employeeName || '-'
  },
  {
    key: 'appealStatus',
    title: $t('page.hrm.performance.workbench.colAppealStatus'),
    align: 'center',
    width: 120,
    render: row => renderAppealTag(row.appealStatus)
  },
  {
    key: 'updateTime',
    title: $t('page.hrm.performance.workbench.colUpdateTime'),
    align: 'center',
    width: 170
  },
  {
    key: 'operate',
    title: $t('page.hrm.performance.workbench.colOperate'),
    align: 'center',
    width: 100,
    render: row => (
      <NButton size="small" type="primary" ghost onClick={() => goTodo(row)}>
        {$t('page.hrm.performance.workbench.goHandle')}
      </NButton>
    )
  }
];

const todos = computed(() => workbench.value?.todos ?? []);
const appeals = computed(() => workbench.value?.appeals ?? []);
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <!-- 全周期待办汇总卡 -->
    <NCard
      :title="$t('page.hrm.performance.workbench.cardsTitle')"
      :bordered="false"
      size="small"
      class="card-wrapper"
    >
      <NGrid :x-gap="16" :y-gap="12" responsive="screen" item-responsive cols="2 s:3 m:6">
        <NGi v-for="item in cardItems" :key="item.key" span="1">
          <NStatistic
            :label="cardLabel(item.key)"
            :class="cardClickable(item) ? 'cursor-pointer' : ''"
            @click="goCard(item)"
          >
            <span class="text-20px font-bold">{{ cardValues?.[item.key] ?? 0 }}</span>
          </NStatistic>
        </NGi>
      </NGrid>
    </NCard>

    <!-- 最近待办列表 -->
    <NCard
      :title="$t('page.hrm.performance.workbench.todosTitle')"
      :bordered="false"
      size="small"
      class="card-wrapper"
    >
      <NDataTable
        :columns="todoColumns"
        :data="todos"
        :loading="loading"
        :pagination="false"
        :scroll-x="900"
        size="small"
      />
    </NCard>

    <!-- 申诉中的任务单列 -->
    <NCard
      :title="$t('page.hrm.performance.workbench.appealsTitle')"
      :bordered="false"
      size="small"
      class="card-wrapper"
    >
      <NDataTable
        :columns="appealColumns"
        :data="appeals"
        :loading="loading"
        :pagination="false"
        :scroll-x="800"
        size="small"
      />
    </NCard>
  </div>
</template>

<style scoped></style>
