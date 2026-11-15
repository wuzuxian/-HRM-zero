<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { GetScheduleCycleList, GenerateScheduleFromCycle } from '@/service/api';
import { $t } from '@/locales';

defineOptions({
  name: 'ScheduleGenerateModal'
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

const cycleOptions = ref<{ label: string; value: string }[]>([]);
const cycleUuid = ref('');
const anchorDate = ref<number | null>(null);
const overwrite = ref(false);
const submitting = ref(false);

async function loadCycles() {
  const { data, error } = await GetScheduleCycleList({ current: 1, size: 100, status: '1' });
  if (error || !data) return;
  cycleOptions.value = data.records.map(c => ({
    label: `${c.cycleName} (${c.shiftUuids.length}${$t('page.hrm.attendance.scheduleBoard.cycleDaysUnit')})`,
    value: c.uuid
  }));
  if (cycleOptions.value.length > 0 && !cycleUuid.value) {
    cycleUuid.value = cycleOptions.value[0]!.value;
  }
}

onMounted(() => {
  loadCycles();
});

function fmtTs(ts: number) {
  const d = new Date(ts);
  const m = `${d.getMonth() + 1}`.padStart(2, '0');
  const day = `${d.getDate()}`.padStart(2, '0');
  return `${d.getFullYear()}-${m}-${day}`;
}

async function handleSubmit() {
  if (!cycleUuid.value) {
    window.$message?.error($t('page.hrm.attendance.scheduleBoard.cycleRequired'));
    return;
  }
  submitting.value = true;
  const { data, error } = await GenerateScheduleFromCycle({
    groupUuid: props.groupUuid,
    month: props.month,
    cycleUuid: cycleUuid.value,
    anchorDate: anchorDate.value ? fmtTs(anchorDate.value) : undefined,
    overwrite: overwrite.value ? '1' : undefined
  });
  submitting.value = false;
  if (error) return;
  window.$message?.success($t('page.hrm.attendance.scheduleBoard.generateSuccess', { n: data?.writtenCount ?? 0 }));
  visible.value = false;
  emit('submitted');
}
</script>

<template>
  <NModal v-model:show="visible" :title="$t('page.hrm.attendance.scheduleBoard.generate')" preset="card" class="w-480px">
    <NForm label-placement="left" :label-width="110">
      <NFormItem :label="$t('page.hrm.attendance.scheduleBoard.cycle')">
        <NSelect v-model:value="cycleUuid" :options="cycleOptions" />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.attendance.scheduleBoard.anchorDate')">
        <NDatePicker v-model:value="anchorDate" type="date" clearable value-format="yyyy-MM-dd" />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.attendance.scheduleBoard.overwrite')">
        <NSwitch v-model:value="overwrite" />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.attendance.scheduleBoard.anchorHintLabel')">
        <span class="text-xs text-gray-500">{{ $t('page.hrm.attendance.scheduleBoard.anchorHint') }}</span>
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
