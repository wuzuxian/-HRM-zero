<script setup lang="tsx">
import { ref, watch } from 'vue';
import { NButton, NTag } from 'naive-ui';
import { GetPunchStatisticsDetail } from '@/service/api';
import { $t } from '@/locales';

defineOptions({
  name: 'PunchStatisticsDetailModal'
});

interface Props {
  visible: boolean;
  person: Api.Hrm.PunchPersonStats | null;
  startDate?: string;
  endDate?: string;
}

const props = defineProps<Props>();

interface Emits {
  (e: 'update:visible', value: boolean): void;
}

const emit = defineEmits<Emits>();

const loading = ref(false);
const records = ref<Api.Hrm.Punch[]>([]);

const show = ref(props.visible);
watch(
  () => props.visible,
  val => {
    show.value = val;
    if (val && props.person) {
      fetchDetail();
    }
  }
);
watch(show, val => {
  if (!val) emit('update:visible', false);
});

async function fetchDetail() {
  if (!props.person) return;
  loading.value = true;
  const { data, error } = await GetPunchStatisticsDetail({
    employeeUuid: props.person.employeeUuid,
    startDate: props.startDate,
    endDate: props.endDate
  });
  loading.value = false;
  if (error) return;
  records.value = data?.records ?? [];
}

function renderSettle(status: string) {
  const map: Record<string, { label: string; type: 'success' | 'warning' | 'error' | 'info' | 'default' }> = {
    normal: { label: $t('page.hrm.attendance.punchRecord.settleNormal'), type: 'success' },
    late: { label: $t('page.hrm.attendance.punchRecord.settleLate'), type: 'warning' },
    early: { label: $t('page.hrm.attendance.punchRecord.settleEarly'), type: 'warning' },
    missing: { label: $t('page.hrm.attendance.punchRecord.settleMissing'), type: 'error' },
    rest: { label: $t('page.hrm.attendance.punchRecord.settleRest'), type: 'default' },
    no_rule: { label: $t('page.hrm.attendance.punchRecord.settleNoRule'), type: 'info' }
  };
  const meta = map[status];
  if (!meta) return status || '-';
  return (
    <NTag type={meta.type} size="small">
      {meta.label}
    </NTag>
  );
}

const columns = ref([
  { key: 'punchDate', title: $t('page.hrm.attendance.punchRecord.punchDate'), align: 'center' as const, width: 110 },
  { key: 'weekday', title: $t('page.hrm.attendance.punchRecord.weekday'), align: 'center' as const, width: 70 },
  { key: 'clockInTime', title: $t('page.hrm.attendance.punchRecord.clockInTime'), align: 'center' as const, width: 150, render: (row: Api.Hrm.Punch) => row.clockInTime || '-' },
  { key: 'clockInStatus', title: $t('page.hrm.attendance.punchRecord.clockInStatus'), align: 'center' as const, width: 90, render: (row: Api.Hrm.Punch) => renderClockStatus(row.clockInStatus) },
  { key: 'clockOutTime', title: $t('page.hrm.attendance.punchRecord.clockOutTime'), align: 'center' as const, width: 150, render: (row: Api.Hrm.Punch) => row.clockOutTime || '-' },
  { key: 'clockOutStatus', title: $t('page.hrm.attendance.punchRecord.clockOutStatus'), align: 'center' as const, width: 90, render: (row: Api.Hrm.Punch) => renderClockStatus(row.clockOutStatus) },
  { key: 'settleStatus', title: $t('page.hrm.attendance.punchRecord.settleStatus'), align: 'center' as const, render: (row: Api.Hrm.Punch) => renderSettle(row.settleStatus) },
  { key: 'settleLateMinutes', title: $t('page.hrm.attendance.punchRecord.lateMinutes'), align: 'center' as const, width: 100 },
  { key: 'settleEarlyMinutes', title: $t('page.hrm.attendance.punchRecord.earlyMinutes'), align: 'center' as const, width: 100 },
  { key: 'settleWorkHours', title: $t('page.hrm.attendance.punchRecord.workHours'), align: 'center' as const, width: 100 }
]);

function renderClockStatus(status: string) {
  if (status === 'normal') {
    return (
      <NTag type="success" size="small">
        {$t('page.hrm.attendance.punchRecord.statusNormal')}
      </NTag>
    );
  }
  if (status === 'field') {
    return (
      <NTag type="info" size="small">
        {$t('page.hrm.attendance.punchRecord.statusField')}
      </NTag>
    );
  }
  return (
    <NTag type="error" size="small">
      {$t('page.hrm.attendance.punchRecord.statusMissing')}
    </NTag>
  );
}
</script>

<template>
  <NModal
    v-model:show="show"
    :title="$t('page.hrm.attendance.punchStatistics.detailTitle')"
    preset="card"
    class="w-960px"
  >
    <div class="mb-8px text-gray-500">
      {{ person?.employeeName }}
      <span v-if="startDate && endDate">（{{ startDate }} ~ {{ endDate }}）</span>
    </div>
    <NDataTable :columns="columns" :data="records" size="small" :loading="loading" :max-height="420" :scroll-x="1000" />
    <template #footer>
      <NSpace justify="end">
        <NButton @click="show = false">{{ $t('common.close') }}</NButton>
      </NSpace>
    </template>
  </NModal>
</template>
