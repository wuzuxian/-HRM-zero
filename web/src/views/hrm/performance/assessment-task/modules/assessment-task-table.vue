<script setup lang="tsx">
import { reactive, ref } from 'vue';
import { NButton, NTag } from 'naive-ui';
import { ConfirmPerfAssessmentPlan, ConfirmPerfAssessmentResult, GetPerfAssessmentList } from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { useAuth } from '@/hooks/business/auth';
import {
  perfAssessmentAppealStatusRecord,
  perfAssessmentAppealStatusTag,
  perfAssessmentNodeRecord,
  perfAssessmentNodeTag,
  perfAssessmentTaskStatusRecord,
  perfAssessmentTaskStatusTag
} from '@/constants/hrm-performance';
import { useTable } from '@/hooks/common/table';
import AssessmentScoreModal from './assessment-score-modal.vue';
import AssessmentAppealModal from './assessment-appeal-modal.vue';

defineOptions({
  name: 'AssessmentTaskTable'
});

/**
 * 考核任务列表（HRM-43 五个填报列表页共用一套组件模式： 考核单号/当前节点/任务状态/自评得分/评审得分/最终得分等级列 + 阶段操作按钮）。 stage 决定行操作与默认节点过滤： plan — 目标确认页（节点1：确认目标）
 * self-review — 自评填报页（节点2：逐项评分） leader-review — 评审填报页（节点3 评审 + 节点4 校准） result — 结果确认页（节点5：确认结果 / 发起申诉入口）
 */
const props = defineProps<{
  stage: 'plan' | 'self-review' | 'leader-review' | 'result';
}>();

const appStore = useAppStore();
const { hasAuth } = useAuth();

const presetNode: Record<string, string> = {
  plan: '1',
  'self-review': '2',
  'leader-review': '',
  result: '5'
};

const {
  columns,
  columnChecks,
  data,
  getData,
  getDataByPage,
  loading,
  mobilePagination,
  searchParams,
  resetSearchParams
} = useTable({
  apiFn: GetPerfAssessmentList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    periodUuid: null,
    employeeUuid: null,
    currentNode: (presetNode[props.stage] || null) as Api.Hrm.Performance.AssessmentNode | null,
    taskStatus: null
  },
  columns: () => [
    {
      key: 'index',
      title: $t('common.index'),
      align: 'center',
      width: 64
    },
    {
      key: 'assessmentNo',
      title: $t('page.hrm.performance.assessment.assessmentNo'),
      align: 'center',
      width: 130
    },
    {
      key: 'periodName',
      title: $t('page.hrm.performance.assessment.periodName'),
      align: 'center',
      minWidth: 150,
      render: row => `${row.periodName}（${row.periodNo}）`
    },
    {
      key: 'employeeName',
      title: $t('page.hrm.performance.assessment.employeeName'),
      align: 'center',
      width: 90
    },
    {
      key: 'templateName',
      title: $t('page.hrm.performance.assessment.templateName'),
      align: 'center',
      minWidth: 110
    },
    {
      key: 'currentNode',
      title: $t('page.hrm.performance.assessment.currentNode'),
      align: 'center',
      width: 100,
      render: row => (
        <NTag type={perfAssessmentNodeTag(row.currentNode)}>{$t(perfAssessmentNodeRecord[row.currentNode])}</NTag>
      )
    },
    {
      key: 'taskStatus',
      title: $t('page.hrm.performance.assessment.taskStatus'),
      align: 'center',
      width: 90,
      render: row => (
        <NTag type={perfAssessmentTaskStatusTag(row.taskStatus)}>
          {$t(perfAssessmentTaskStatusRecord[row.taskStatus])}
        </NTag>
      )
    },
    {
      key: 'selfScore',
      title: $t('page.hrm.performance.assessment.selfScore'),
      align: 'center',
      width: 90,
      render: row => renderScoreGrade(row.selfScore, row.selfGrade)
    },
    {
      key: 'reviewScore',
      title: $t('page.hrm.performance.assessment.reviewScore'),
      align: 'center',
      width: 90,
      render: row => renderScoreGrade(row.reviewScore, row.reviewGrade)
    },
    {
      key: 'finalScore',
      title: $t('page.hrm.performance.assessment.finalScore'),
      align: 'center',
      width: 110,
      render: row => renderScoreGrade(row.finalScore, row.finalGrade)
    },
    {
      key: 'appealStatus',
      title: $t('page.hrm.performance.assessment.appealStatus'),
      align: 'center',
      width: 100,
      render: row => {
        if (!row.appealStatus) return null;
        return (
          <NTag type={perfAssessmentAppealStatusTag(row.appealStatus)}>
            {$t(perfAssessmentAppealStatusRecord[row.appealStatus])}
          </NTag>
        );
      }
    },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      // 结果确认页最多同排 3 按钮（申诉裁决/确认结果/发起申诉/详情），200px 放不下会被截断
      width: 300,
      render: row => <div class="flex-center gap-8px">{renderOperates(row)}</div>
    }
  ]
});

function renderScoreGrade(score: number, grade: string) {
  if (score === 0 && !grade) return null;
  const parts: string[] = [];
  if (grade) parts.push(grade);
  if (score !== 0) parts.push(String(score));
  return parts.join(' / ') || null;
}

const loadingStatus = reactive<Record<string, boolean>>({});

function renderOperates(row: Api.Hrm.Performance.PerfAssessment) {
  const ongoing = row.taskStatus === '0';
  const buttons: any[] = [];
  if (props.stage === 'plan' && ongoing && row.currentNode === '1') {
    buttons.push(
      <NButton size="small" type="primary" ghost onClick={() => handleConfirmPlan(row)}>
        {$t('page.hrm.performance.assessment.confirmPlan')}
      </NButton>
    );
  }
  if (props.stage === 'self-review' && ongoing && row.currentNode === '2') {
    buttons.push(
      <NButton size="small" type="primary" ghost onClick={() => openScoreModal('self', row)}>
        {$t('page.hrm.performance.assessment.fillSelfReview')}
      </NButton>
    );
  }
  if (props.stage === 'leader-review' && ongoing && row.currentNode === '3') {
    buttons.push(
      <NButton size="small" type="primary" ghost onClick={() => openScoreModal('review', row)}>
        {$t('page.hrm.performance.assessment.fillLeaderReview')}
      </NButton>
    );
  }
  if (props.stage === 'leader-review' && ongoing && row.currentNode === '4') {
    buttons.push(
      <NButton size="small" type="warning" ghost onClick={() => openScoreModal('calibrate', row)}>
        {$t('page.hrm.performance.assessment.fillCalibration')}
      </NButton>
    );
  }
  if (props.stage === 'result' && ongoing && row.currentNode === '5') {
    buttons.push(
      <NButton size="small" type="primary" ghost onClick={() => handleConfirmResult(row)}>
        {$t('page.hrm.performance.assessment.confirmResult')}
      </NButton>,
      <NButton size="small" type="error" ghost onClick={() => openAppealModal(row)}>
        {$t('page.hrm.performance.assessment.appeal')}
      </NButton>
    );
  }
  // 申诉中（taskStatus '2'）：持校准码者在结果确认页打开裁决弹窗（弹窗自动切 handleMode）
  if (props.stage === 'result' && row.taskStatus === '2' && hasAuth('hrm:perf_assessment:calibrate')) {
    buttons.unshift(
      <NButton size="small" type="warning" ghost onClick={() => openAppealModal(row)}>
        {$t('page.hrm.performance.assessment.handleAppeal')}
      </NButton>
    );
  }
  // 详情始终可点（只读弹窗展示快照 items + 各阶段记录）
  buttons.push(
    <NButton size="small" onClick={() => openScoreModal('view', row)}>
      {$t('page.hrm.performance.assessment.detailModalTitle')}
    </NButton>
  );
  return buttons;
}

async function handleConfirmPlan(row: Api.Hrm.Performance.PerfAssessment) {
  loadingStatus[row.uuid] = true;
  const { error } = await ConfirmPerfAssessmentPlan(row.uuid);
  loadingStatus[row.uuid] = false;
  if (!error) {
    window.$message?.success($t('page.hrm.performance.assessment.confirmPlanSuccess'));
    getData();
  }
}

async function handleConfirmResult(row: Api.Hrm.Performance.PerfAssessment) {
  loadingStatus[row.uuid] = true;
  const { error } = await ConfirmPerfAssessmentResult(row.uuid);
  loadingStatus[row.uuid] = false;
  if (!error) {
    window.$message?.success($t('page.hrm.performance.assessment.confirmResultSuccess'));
    getData();
  }
}

// ---- 评分弹窗（自评/评审/校准共用；view 只读） ----

const scoreVisible = ref(false);
const scoreMode = ref<'self' | 'review' | 'calibrate' | 'view'>('view');
const scoringRow = ref<Api.Hrm.Performance.PerfAssessment | null>(null);

function openScoreModal(mode: 'self' | 'review' | 'calibrate' | 'view', row: Api.Hrm.Performance.PerfAssessment) {
  scoreMode.value = mode;
  scoringRow.value = row;
  scoreVisible.value = true;
}

// ---- 申诉弹窗（结果确认页入口；申诉中任务进入裁决模式） ----

const appealVisible = ref(false);
const appealingRow = ref<Api.Hrm.Performance.PerfAssessment | null>(null);

function openAppealModal(row: Api.Hrm.Performance.PerfAssessment) {
  appealingRow.value = row;
  appealVisible.value = true;
}

function onSubmitted() {
  getData();
}
</script>

<template>
  <NCard :bordered="false" size="small" class="sm:flex-1-hidden card-wrapper">
    <template #header>
      {{ $t(`page.hrm.performance.assessment.${stage}Title`) }}
    </template>
    <template #header-extra>
      <TableHeaderOperation v-model:columns="columnChecks" :loading="loading" @refresh="getData" />
    </template>
    <template #footer>
      <div class="flex flex-wrap items-center gap-8px">
        <NInput
          v-model:value="searchParams.periodUuid"
          clearable
          size="small"
          class="w-220px"
          :placeholder="$t('page.hrm.performance.assessment.periodFilter')"
          @clear="() => getDataByPage()"
          @keyup.enter="() => getDataByPage()"
        />
        <NButton size="small" @click="resetSearchParams">
          {{ $t('common.reset') }}
        </NButton>
        <NButton size="small" type="primary" ghost @click="() => getDataByPage()">
          {{ $t('common.search') }}
        </NButton>
      </div>
    </template>
    <NDataTable
      :columns="columns"
      :data="data"
      size="small"
      :flex-height="!appStore.isMobile"
      :scroll-x="1500"
      :loading="loading"
      remote
      :row-key="row => row.uuid"
      :pagination="mobilePagination"
      class="sm:h-full"
    />
    <AssessmentScoreModal
      v-model:visible="scoreVisible"
      :mode="scoreMode"
      :row-data="scoringRow"
      @submitted="onSubmitted"
    />
    <AssessmentAppealModal v-model:visible="appealVisible" :row-data="appealingRow" @submitted="onSubmitted" />
  </NCard>
</template>

<style scoped></style>
