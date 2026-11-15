<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useNaiveForm } from '@/hooks/common/form';
import { CancelLeave } from '@/service/api';
import { $t } from '@/locales';

defineOptions({
  name: 'LeaveCancelModal'
});

interface Props {
  /** the bill row being settled (销假) */
  rowData?: Api.Hrm.HrmLeave | null;
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

const title = computed(() =>
  `${$t('page.hrm.attendance.leaveCancel.cancelLeave')} — ${props.rowData?.billNo ?? ''}`
);

const model = reactive({
  actualRange: null as [number, number] | null
});

/** 剩余可休 = 预计天数（供提示） */
const planDays = computed(() => props.rowData?.planDays ?? 0);

function handleInitModel() {
  model.actualRange = null;
}

function rangeToBillTime(ts: number) {
  const d = new Date(ts);
  const pad = (n: number) => `${n}`.padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:00`;
}

async function handleSubmit() {
  if (!props.rowData) return;
  if (!model.actualRange || model.actualRange.length !== 2) {
    window.$message?.error($t('page.hrm.attendance.leaveCancel.form.actualRangeRequired'));
    return;
  }
  confirmStartLoading();
  const { error } = await CancelLeave({
    uuid: props.rowData.uuid,
    actualStartTime: rangeToBillTime(model.actualRange[0]),
    actualEndTime: rangeToBillTime(model.actualRange[1])
  });
  confirmEndLoading();
  if (!error) {
    window.$message?.success($t('page.hrm.attendance.leaveCancel.cancelSuccess'));
    visible.value = false;
    emit('submitted');
  }
}

const settleHint = ref('');

watch(visible, () => {
  if (visible.value) {
    handleInitModel();
    settleHint.value = `${$t('page.hrm.attendance.leaveCancel.settleHintPrefix')} ${planDays.value} ${$t(
      'page.hrm.attendance.leaveCancel.settleHintSuffix'
    )}`;
    restoreValidation();
  }
});
</script>

<template>
  <NModal v-model:show="visible" :title="title" preset="card" class="w-560px">
    <NForm ref="formRef" label-placement="left" :label-width="100">
      <NFormItem :label="$t('page.hrm.attendance.leaveCancel.actualRange')" path="actualRange">
        <NDatePicker
          v-model:value="model.actualRange"
          type="datetimerange"
          clearable
          :start-placeholder="$t('page.hrm.attendance.leaveCancel.form.actualStart')"
          :end-placeholder="$t('page.hrm.attendance.leaveCancel.form.actualEnd')"
        />
      </NFormItem>
      <div class="pb-8px text-12px text-gray-400">
        {{ settleHint }}
        <div>{{ $t('page.hrm.attendance.leaveCancel.settleRule') }}</div>
      </div>
    </NForm>
    <template #footer>
      <NSpace justify="end">
        <NButton @click="visible = false">{{ $t('common.cancel') }}</NButton>
        <NButton type="primary" :loading="confirmLoading" @click="handleSubmit">
          {{ $t('page.hrm.attendance.leaveCancel.cancelLeave') }}
        </NButton>
      </NSpace>
    </template>
  </NModal>
</template>

<style scoped></style>
