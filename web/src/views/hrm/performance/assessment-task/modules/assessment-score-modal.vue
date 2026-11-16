<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import type { InputInst } from 'naive-ui';
import { $t } from '@/locales';
import {
  GetPerfAssessmentDetail,
  SubmitPerfAssessmentCalibration,
  SubmitPerfAssessmentLeaderReview,
  SubmitPerfAssessmentSelfReview
} from '@/service/api';

defineOptions({
  name: 'AssessmentScoreModal'
});

/**
 * 考核任务评分弹窗（HRM-43 共用）：
 *   self — 自评填报（节点2，逐项评分权重汇总）
 *   review — 上级评审（节点3）
 *   calibrate — 校准填报（节点4，调整最终得分/等级）
 *   view — 只读详情（模板快照 items + 各阶段填报记录）
 */
type ScoreMode = 'self' | 'review' | 'calibrate' | 'view';

const props = defineProps<{
  mode: ScoreMode;
  rowData: Api.Hrm.Performance.PerfAssessment | null;
}>();

interface Emits {
  (e: 'update:visible', value: boolean): void;
  (e: 'submitted'): void;
}

const emit = defineEmits<Emits>();

const visible = defineModel<boolean>('visible', {
  default: false
});

const title = computed(() => {
  const keys: Record<ScoreMode, App.I18n.I18nKey> = {
    self: 'page.hrm.performance.assessment.selfModalTitle',
    review: 'page.hrm.performance.assessment.reviewModalTitle',
    calibrate: 'page.hrm.performance.assessment.calibrateModalTitle',
    view: 'page.hrm.performance.assessment.detailModalTitle'
  };
  return $t(keys[props.mode]);
});

const detail = ref<Api.Hrm.Performance.AssessmentDetail | null>(null);
const detailLoading = ref(false);

const scores = reactive<Record<string, number>>({});
const selfComment = ref('');
const reviewComment = ref('');
const finalScore = ref<number | null>(null);
const finalGrade = ref('');
const calibComment = ref('');

async function loadDetail() {
  if (!props.rowData?.uuid) return;
  detailLoading.value = true;
  const { data: res, error } = await GetPerfAssessmentDetail(props.rowData.uuid);
  detailLoading.value = false;
  if (error || !res) return;
  detail.value = res;
  Object.keys(scores).forEach(key => delete scores[key]);
  const targetItems =
    props.mode === 'review'
      ? (res.reviewItems ?? [])
      : (res.selfItems ?? []);
  targetItems.forEach(item => {
    scores[item.indicatorUuid] = item.score;
  });
  selfComment.value = res.selfComment || '';
  reviewComment.value = res.reviewComment || '';
  finalScore.value = props.mode === 'calibrate' ? (res.finalScore || res.reviewScore || null) : res.finalScore;
  finalGrade.value = res.finalGrade || '';
  calibComment.value = res.calibComment || '';
}

watch(visible, newVal => {
  if (newVal) {
    loadDetail();
  } else {
    detail.value = null;
    selfComment.value = '';
    reviewComment.value = '';
    finalScore.value = null;
    finalGrade.value = '';
    calibComment.value = '';
  }
});

const items = computed<Api.Hrm.Performance.AssessmentItemSnapshot[]>(() => detail.value?.items ?? []);

const showItemsScoring = computed(() => props.mode === 'self' || props.mode === 'review' || props.mode === 'view');

const submitLoading = ref(false);

async function handleSubmit() {
  if (!props.rowData) return;
  if (props.mode === 'calibrate') {
    if (finalScore.value === null) {
      window.$message?.error($t('page.hrm.performance.assessment.finalScoreRequired'));
      return;
    }
    submitLoading.value = true;
    const { error } = await SubmitPerfAssessmentCalibration(
      props.rowData.uuid,
      finalScore.value,
      finalGrade.value || undefined,
      calibComment.value
    );
    submitLoading.value = false;
    if (!error) {
      window.$message?.success($t('common.updateSuccess'));
      emit('submitted');
      visible.value = false;
    }
    return;
  }

  const itemScores = items.value.map(item => ({
    indicatorUuid: item.indicatorUuid,
    score: scores[item.indicatorUuid] ?? 0
  }));
  submitLoading.value = true;
  const { error } =
    props.mode === 'self'
      ? await SubmitPerfAssessmentSelfReview(props.rowData.uuid, itemScores, selfComment.value)
      : await SubmitPerfAssessmentLeaderReview(props.rowData.uuid, itemScores, reviewComment.value);
  submitLoading.value = false;
  if (!error) {
    window.$message?.success($t('common.updateSuccess'));
    emit('submitted');
    visible.value = false;
  }
}
</script>

<template>
  <NModal
    v-model:show="visible"
    preset="card"
    :title="title"
    class="w-720px"
  >
    <NSpin :show="detailLoading">
      <NDescriptions v-if="detail" :column="2" label-placement="left" size="small" bordered>
        <NDescriptionsItem :label="$t('page.hrm.performance.assessment.assessmentNo')">
          {{ detail.assessmentNo }}
        </NDescriptionsItem>
        <NDescriptionsItem :label="$t('page.hrm.performance.assessment.employeeName')">
          {{ detail.employeeName }}
        </NDescriptionsItem>
        <NDescriptionsItem :label="$t('page.hrm.performance.assessment.periodName')">
          {{ `${detail.periodName}（${detail.periodNo}）` }}
        </NDescriptionsItem>
        <NDescriptionsItem :label="$t('page.hrm.performance.assessment.templateName')">
          {{ detail.templateName }}
        </NDescriptionsItem>
      </NDescriptions>

      <NDivider>{{ $t('page.hrm.performance.assessment.itemsSnapshot') }}</NDivider>

      <!-- 逐项得分表（模板快照 items） -->
      <div class="mt-8px overflow-auto">
        <table class="w-full border-collapse text-sm">
          <thead>
            <tr class="bg-#f6f8fa">
              <th class="border border-#efeff5 p-6px">{{ $t('page.hrm.performance.assessment.indicatorName') }}</th>
              <th class="border border-#efeff5 p-6px w-80px">{{ $t('page.hrm.performance.assessment.weight') }}</th>
              <th class="border border-#efeff5 p-6px w-140px">
                {{ showItemsScoring ? $t('page.hrm.performance.assessment.score') : '' }}
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in items" :key="item.indicatorUuid">
              <td class="border border-#efeff5 p-6px">
                {{ item.indicatorName }}（{{ item.indicatorCode }}）
                <NTag size="tiny" :bordered="false" class="ml-4px">{{ item.indicatorType }}</NTag>
              </td>
              <td class="border border-#efeff5 p-6px text-center">{{ item.weight }}%</td>
              <td class="border border-#efeff5 p-6px text-center">
                <NInputNumber
                  v-if="props.mode === 'self' || props.mode === 'review'"
                  v-model:value="scores[item.indicatorUuid]"
                  size="small"
                  :min="0"
                  :precision="2"
                  class="w-120px"
                />
                <span v-else>{{ scores[item.indicatorUuid] ?? '-' }}</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <template v-if="props.mode === 'self'">
        <NDivider>{{ $t('page.hrm.performance.assessment.selfComment') }}</NDivider>
        <NInput v-model:value="selfComment" type="textarea" :rows="2" />
      </template>
      <template v-else-if="props.mode === 'review'">
        <NDivider>{{ $t('page.hrm.performance.assessment.reviewComment') }}</NDivider>
        <NInput v-model:value="reviewComment" type="textarea" :rows="2" />
      </template>
      <template v-else-if="props.mode === 'calibrate'">
        <NDivider>{{ $t('page.hrm.performance.assessment.calibration') }}</NDivider>
        <NForm label-placement="left" :label-width="90">
          <NFormItem :label="$t('page.hrm.performance.assessment.finalScore')">
            <NInputNumber v-model:value="finalScore" :min="0" :precision="2" class="w-200px" />
          </NFormItem>
          <NFormItem :label="$t('page.hrm.performance.assessment.finalGrade')">
            <NInput v-model:value="finalGrade" :placeholder="$t('page.hrm.performance.assessment.finalGradeAuto')" />
          </NFormItem>
          <NFormItem :label="$t('page.hrm.performance.assessment.calibComment')">
            <NInput v-model:value="calibComment" type="textarea" :rows="2" />
          </NFormItem>
        </NForm>
      </template>
    </NSpin>

    <template #footer>
      <NSpace justify="end">
        <NButton @click="visible = false">{{ $t('common.cancel') }}</NButton>
        <NButton
          v-if="props.mode !== 'view'"
          type="primary"
          :loading="submitLoading"
          @click="handleSubmit"
        >
          {{ $t('common.confirm') }}
        </NButton>
      </NSpace>
    </template>
  </NModal>
</template>

<style scoped></style>
