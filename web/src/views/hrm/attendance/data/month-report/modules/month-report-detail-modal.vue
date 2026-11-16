<script setup lang="tsx">
import { ref, watch } from 'vue';
import { NTag } from 'naive-ui';
import { GetAttendanceMonthReportDetail } from '@/service/api';
import { $t } from '@/locales';

defineOptions({
  name: 'MonthReportDetailModal'
});

interface Props {
  visible: boolean;
  row: Api.Hrm.MonthReportRow | null;
  yearMonth?: string;
}

const props = defineProps<Props>();

interface Emits {
  (e: 'update:visible', value: boolean): void;
}

const emit = defineEmits<Emits>();

const loading = ref(false);
const records = ref<Api.Hrm.MonthReportDetailDay[]>([]);

const show = ref(props.visible);
watch(
  () => props.visible,
  val => {
    show.value = val;
    if (val && props.row) {
      fetchDetail();
    }
  }
);
watch(show, val => {
  if (!val) emit('update:visible', false);
});

async function fetchDetail() {
  if (!props.row) return;
  loading.value = true;
  const { data, error } = await GetAttendanceMonthReportDetail({
    employeeUuid: props.row.employeeUuid,
    yearMonth: props.row.yearMonth
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
  { key: 'punchDate', title: $t('page.hrm.attendance.monthReport.detailDate'), align: 'center' as const, width: 110 },
  { key: 'weekday', title: $t('page.hrm.attendance.punchRecord.weekday'), align: 'center' as const, width: 70 },
  { key: 'settleStatus', title: $t('page.hrm.attendance.punchRecord.settleStatus'), align: 'center' as const, render: (row: Api.Hrm.MonthReportDetailDay) => renderSettle(row.settleStatus) },
  { key: 'clockInTime', title: $t('page.hrm.attendance.punchRecord.clockInTime'), align: 'center' as const, width: 160, render: (row: Api.Hrm.MonthReportDetailDay) => row.clockInTime || '-' },
  { key: 'clockOutTime', title: $t('page.hrm.attendance.punchRecord.clockOutTime'), align: 'center' as const, width: 160, render: (row: Api.Hrm.MonthReportDetailDay) => row.clockOutTime || '-' },
  { key: 'settleLateMinutes', title: $t('page.hrm.attendance.punchRecord.lateMinutes'), align: 'center' as const, width: 100 },
  { key: 'settleEarlyMinutes', title: $t('page.hrm.attendance.punchRecord.earlyMinutes'), align: 'center' as const, width: 100 },
  { key: 'settleWorkHours', title: $t('page.hrm.attendance.punchRecord.workHours'), align: 'center' as const, width: 100 }
]);
</script>

<template>
  <NModal
    v-model:show="show"
    :title="$t('page.hrm.attendance.monthReport.detailTitle')"
    preset="card"
    class="w-960px"
  >
    <div class="mb-8px text-gray-500">
      {{ row?.employeeName }}
      <span v-if="yearMonth">（{{ yearMonth }}）</span>
    </div>
    <NDataTable :columns="columns" :data="records" size="small" :loading="loading" :max-height="420" :scroll-x="900" />
    <template #footer>
      <NSpace justify="end">
        <NButton @click="show = false">{{ $t('common.close') }}</NButton>
      </NSpace>
    </template>
  </NModal>
</template>
