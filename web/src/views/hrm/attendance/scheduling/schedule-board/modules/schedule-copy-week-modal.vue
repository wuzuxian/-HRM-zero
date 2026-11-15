<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { CopyScheduleWeek } from '@/service/api';
import { $t } from '@/locales';

defineOptions({
  name: 'ScheduleCopyWeekModal'
});

const props = defineProps<{
  visible: boolean;
  groupUuid: string;
  month: string;
}>();

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void;
  (e: 'submitted'): void;
}>();

const visible = computed({
  get: () => props.visible,
  set: value => emit('update:visible', value)
});

const sourceWeekStart = ref<number | null>(null);
const targetWeekStart = ref<number | null>(null);
const overwrite = ref(false);
const submitting = ref(false);

watch(
  () => props.visible,
  value => {
    if (value) {
      sourceWeekStart.value = null;
      targetWeekStart.value = null;
      overwrite.value = false;
    }
  }
);

function mondayOf(ts: number) {
  const d = new Date(ts);
  const day = d.getDay();
  const diff = day === 0 ? -6 : 1 - day;
  return d.getTime() + diff * 24 * 3600 * 1000;
}

function fmtTs(ts: number) {
  const d = new Date(ts);
  const m = `${d.getMonth() + 1}`.padStart(2, '0');
  const day = `${d.getDate()}`.padStart(2, '0');
  return `${d.getFullYear()}-${m}-${day}`;
}

async function handleSubmit() {
  if (!sourceWeekStart.value || !targetWeekStart.value) {
    window.$message?.error($t('page.hrm.attendance.scheduleBoard.weekRequired'));
    return;
  }
  submitting.value = true;
  const { data, error } = await CopyScheduleWeek({
    groupUuid: props.groupUuid,
    sourceWeekStart: fmtTs(mondayOf(sourceWeekStart.value)),
    targetWeekStart: fmtTs(mondayOf(targetWeekStart.value)),
    overwrite: overwrite.value ? '1' : undefined
  });
  submitting.value = false;
  if (error) return;
  window.$message?.success($t('page.hrm.attendance.scheduleBoard.copySuccess', { n: data?.writtenCount ?? 0 }));
  visible.value = false;
  emit('submitted');
}
</script>

<template>
  <NModal v-model:show="visible" :title="$t('page.hrm.attendance.scheduleBoard.copyWeek')" preset="card" class="w-480px">
    <NForm label-placement="left" :label-width="110">
      <NFormItem :label="$t('page.hrm.attendance.scheduleBoard.sourceWeek')">
        <NDatePicker v-model:value="sourceWeekStart" type="date" value-format="yyyy-MM-dd" />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.attendance.scheduleBoard.targetWeek')">
        <NDatePicker v-model:value="targetWeekStart" type="date" value-format="yyyy-MM-dd" />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.attendance.scheduleBoard.overwrite')">
        <NSwitch v-model:value="overwrite" />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.attendance.scheduleBoard.weekHintLabel')">
        <span class="text-xs text-gray-500">{{ $t('page.hrm.attendance.scheduleBoard.weekHint') }}</span>
      </NFormItem>
    </NForm>
    <template #footer>
      <NSpace justify="end">
        <NButton @click="visible = false">{{ $t('common.cancel') }}</NButton>
        <NButton type="primary" :loading="submitting" @click="handleSubmit">{{ $t('common.confirm') }}</NButton>
      </NSpace>
    </template>
  </NModal>
</template>
