<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { NRadioGroup, NRadio } from 'naive-ui';
import { $t } from '@/locales';
import { useAuth } from '@/hooks/business/auth';
import { AppealPerfAssessment, HandlePerfAssessmentAppeal } from '@/service/api';

defineOptions({
  name: 'AssessmentAppealModal'
});

/**
 * 申诉弹窗（HRM-43 结果确认页申诉入口）：
 *   发起模式 — 被考核人填写原因发起申诉；
 *   裁决模式 — 有 hrm:perf_assessment:calibrate 码者对申诉中任务裁决（维持/调整得分→关闭）。
 */
const props = defineProps<{
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

const { hasAuth } = useAuth();

// 任务申诉中（taskStatus '2'）且有裁决码 → 裁决模式
const isHandleMode = computed(() => {
  const row = props.rowData;
  return !!row && row.taskStatus === '2' && hasAuth('hrm:perf_assessment:calibrate');
});

const title = computed(() =>
  $t(isHandleMode.value ? 'page.hrm.performance.assessment.handleAppeal' : 'page.hrm.performance.assessment.appeal')
);

const reason = ref('');
const decision = ref<Api.Hrm.Performance.AssessmentAppealStatus>('1');
const adjustedScore = ref<number | null>(null);
const adjustedGrade = ref('');
const comment = ref('');
const submitLoading = ref(false);

watch(visible, newVal => {
  if (newVal) {
    reason.value = '';
    decision.value = '1';
    adjustedScore.value = props.rowData?.finalScore ?? null;
    adjustedGrade.value = props.rowData?.finalGrade ?? '';
    comment.value = '';
  }
});

async function handleSubmit() {
  if (!props.rowData) return;
  if (!isHandleMode.value) {
    if (!reason.value.trim()) {
      window.$message?.error($t('page.hrm.performance.assessment.appealReasonRequired'));
      return;
    }
    submitLoading.value = true;
    const { error } = await AppealPerfAssessment(props.rowData.uuid, reason.value);
    submitLoading.value = false;
    if (!error) {
      window.$message?.success($t('page.hrm.performance.assessment.appealSuccess'));
      emit('submitted');
      visible.value = false;
    }
    return;
  }

  if (decision.value === '2' && adjustedScore.value === null) {
    window.$message?.error($t('page.hrm.performance.assessment.adjustedScoreRequired'));
    return;
  }
  submitLoading.value = true;
  const { error } = await HandlePerfAssessmentAppeal({
    uuid: props.rowData.uuid,
    decision: decision.value,
    comment: comment.value,
    adjustedScore: decision.value === '2' ? (adjustedScore.value ?? 0) : undefined,
    adjustedGrade: decision.value === '2' ? adjustedGrade.value || undefined : undefined
  });
  submitLoading.value = false;
  if (!error) {
    window.$message?.success($t('page.hrm.performance.assessment.handleAppealSuccess'));
    emit('submitted');
    visible.value = false;
  }
}
</script>

<template>
  <NModal v-model:show="visible" preset="card" :title="title" class="w-560px">
    <NForm label-placement="left" :label-width="90">
      <template v-if="!isHandleMode">
        <NFormItem :label="$t('page.hrm.performance.assessment.assessmentNo')" required>
          <span>{{ rowData?.assessmentNo }}</span>
        </NFormItem>
        <NFormItem :label="$t('page.hrm.performance.assessment.appealReason')" required>
          <NInput
            v-model:value="reason"
            type="textarea"
            :rows="4"
            :placeholder="$t('page.hrm.performance.assessment.appealReasonPlaceholder')"
          />
        </NFormItem>
      </template>
      <template v-else>
        <NFormItem :label="$t('page.hrm.performance.assessment.decision')" required>
          <NRadioGroup v-model:value="decision">
            <NRadio value="1">{{ $t('page.hrm.performance.assessment.decisionUphold') }}</NRadio>
            <NRadio value="2">{{ $t('page.hrm.performance.assessment.decisionAdjust') }}</NRadio>
          </NRadioGroup>
        </NFormItem>
        <template v-if="decision === '2'">
          <NFormItem :label="$t('page.hrm.performance.assessment.adjustedScore')" required>
            <NInputNumber v-model:value="adjustedScore" :min="0" :precision="2" class="w-200px" />
          </NFormItem>
          <NFormItem :label="$t('page.hrm.performance.assessment.finalGrade')">
            <NInput v-model:value="adjustedGrade" />
          </NFormItem>
        </template>
        <NFormItem :label="$t('page.hrm.performance.assessment.decisionComment')">
          <NInput v-model:value="comment" type="textarea" :rows="3" />
        </NFormItem>
      </template>
    </NForm>

    <template #footer>
      <NSpace justify="end">
        <NButton @click="visible = false">{{ $t('common.cancel') }}</NButton>
        <NButton type="primary" :loading="submitLoading" @click="handleSubmit">
          {{ $t('common.confirm') }}
        </NButton>
      </NSpace>
    </template>
  </NModal>
</template>

<style scoped></style>
