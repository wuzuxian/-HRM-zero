<script setup lang="tsx">
import { ref, watch } from 'vue';
import { GetAttendanceMonthLockRecords } from '@/service/api';
import { $t } from '@/locales';

defineOptions({
  name: 'MonthReportLockRecordsModal'
});

interface Props {
  visible: boolean;
  yearMonth?: string;
}

const props = defineProps<Props>();

interface Emits {
  (e: 'update:visible', value: boolean): void;
}

const emit = defineEmits<Emits>();

const loading = ref(false);
const records = ref<Api.Hrm.MonthReportLockRecord[]>([]);

const show = ref(props.visible);
watch(
  () => props.visible,
  val => {
    show.value = val;
    if (val) {
      fetchRecords();
    }
  }
);
watch(show, val => {
  if (!val) emit('update:visible', false);
});

async function fetchRecords() {
  loading.value = true;
  const { data, error } = await GetAttendanceMonthLockRecords({ yearMonth: props.yearMonth });
  loading.value = false;
  if (error) return;
  records.value = data?.records ?? [];
}

const columns = ref([
  { key: 'createTime', title: $t('page.hrm.attendance.monthReport.lockRecordTime'), align: 'center' as const, width: 170 },
  { key: 'yearMonth', title: $t('page.hrm.attendance.monthReport.yearMonth'), align: 'center' as const, width: 100 },
  { key: 'actionName', title: $t('page.hrm.attendance.monthReport.lockRecordAction'), align: 'center' as const, width: 110 },
  { key: 'operatorName', title: $t('page.hrm.attendance.monthReport.lockRecordOperator'), align: 'center' as const, width: 120 },
  { key: 'remark', title: $t('page.hrm.attendance.monthReport.lockRecordRemark'), align: 'center' as const, render: (row: Api.Hrm.MonthReportLockRecord) => row.remark || '-' }
]);
</script>

<template>
  <NModal
    v-model:show="show"
    :title="$t('page.hrm.attendance.monthReport.lockRecords')"
    preset="card"
    class="w-820px"
  >
    <NDataTable :columns="columns" :data="records" size="small" :loading="loading" :max-height="420" />
    <template #footer>
      <NSpace justify="end">
        <NButton @click="show = false">{{ $t('common.close') }}</NButton>
      </NSpace>
    </template>
  </NModal>
</template>
