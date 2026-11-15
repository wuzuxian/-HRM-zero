<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useNaiveForm } from '@/hooks/common/form';
import { ConfirmOvertimeBill } from '@/service/api';
import { $t } from '@/locales';

defineOptions({
  name: 'OvertimeBillConfirmModal'
});

interface Props {
  /** the bill row being confirmed (核定) */
  rowData?: Api.Hrm.HrmOvertimeBill | null;
}

const props = defineProps<Props>();

interface Emits {
  (e: 'submitted'): void;
}

const emit = defineEmits<Emits>();

const visible = defineModel<boolean>('visible', {
  default: false
});

const { formRef, restoreValidation } = useNaiveForm();
const { loading: confirmLoading, startLoading: confirmStartLoading, endLoading: confirmEndLoading } = useLoading();

const title = computed(() => `${$t('page.hrm.attendance.overtimeBill.confirm')} — ${props.rowData?.billNo ?? ''}`);

const model = reactive({
  confirmHours: null as number | null,
  compType: 'comp_leave' as Api.Hrm.OvertimeCompType
});

/** 申请时长（供上限提示） */
const applyHours = computed(() => props.rowData?.applyHours ?? 0);

const compTypeOptions = computed(() => [
  { label: $t('page.hrm.attendance.overtimeBill.compTypeLeave'), value: 'comp_leave' },
  { label: $t('page.hrm.attendance.overtimeBill.compTypeOther'), value: 'other' }
]);

function handleInitModel() {
  model.confirmHours = props.rowData?.confirmHours ? props.rowData.confirmHours : null;
  model.compType = props.rowData?.compType ? props.rowData.compType : 'comp_leave';
}

const hoursRule = [
  {
    validator: (_rule: unknown, value: number | null) => {
      if (value === null || value <= 0) {
        return new Error($t('page.hrm.attendance.overtimeBill.form.confirmHoursRequired'));
      }
      if (value > applyHours.value + 1e-9) {
        return new Error($t('page.hrm.attendance.overtimeBill.form.confirmHoursExceeded'));
      }
      return true;
    },
    trigger: ['blur', 'change']
  }
];

async function handleSubmit() {
  if (!props.rowData) return;
  if (model.confirmHours === null || model.confirmHours <= 0) {
    window.$message?.error($t('page.hrm.attendance.overtimeBill.form.confirmHoursRequired'));
    return;
  }
  if (model.confirmHours > applyHours.value + 1e-9) {
    window.$message?.error($t('page.hrm.attendance.overtimeBill.form.confirmHoursExceeded'));
    return;
  }
  confirmStartLoading();
  const { error } = await ConfirmOvertimeBill({
    uuid: props.rowData.uuid,
    confirmHours: model.confirmHours,
    compType: model.compType
  });
  confirmEndLoading();
  if (!error) {
    window.$message?.success($t('page.hrm.attendance.overtimeBill.confirmSuccess'));
    visible.value = false;
    emit('submitted');
  }
}

const settleHint = ref('');

watch(visible, () => {
  if (visible.value) {
    handleInitModel();
    settleHint.value = `${$t('page.hrm.attendance.overtimeBill.confirmHintPrefix')} ${applyHours.value} ${$t(
      'page.hrm.attendance.overtimeBill.confirmHintSuffix'
    )}`;
    restoreValidation();
  }
});
</script>

<template>
  <NModal v-model:show="visible" :title="title" preset="card" class="w-560px">
    <NForm ref="formRef" :model="model" label-placement="left" :label-width="100">
      <NFormItem :label="$t('page.hrm.attendance.overtimeBill.confirmHours')" path="confirmHours" :rule="hoursRule">
        <NInputNumber v-model:value="model.confirmHours" :min="0" :step="0.5" class="w-full" />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.attendance.overtimeBill.compType')" path="compType">
        <NRadioGroup v-model:value="model.compType">
          <NRadio v-for="opt in compTypeOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</NRadio>
        </NRadioGroup>
      </NFormItem>
      <div class="pb-8px text-12px text-gray-400">
        {{ settleHint }}
        <div>{{ $t('page.hrm.attendance.overtimeBill.confirmRule') }}</div>
      </div>
    </NForm>
    <template #footer>
      <NSpace justify="end">
        <NButton @click="visible = false">{{ $t('common.cancel') }}</NButton>
        <NButton type="primary" :loading="confirmLoading" @click="handleSubmit">
          {{ $t('page.hrm.attendance.overtimeBill.confirm') }}
        </NButton>
      </NSpace>
    </template>
  </NModal>
</template>

<style scoped></style>
