<script setup lang="tsx">
import { computed, ref } from 'vue';
import { NButton, NPopconfirm, NTag } from 'naive-ui';
import {
  ExportAttendanceMonthReport,
  GetAttendanceMonthReport,
  LockAttendanceMonthReport,
  RecalcAttendanceMonthReport,
  UnlockAttendanceMonthReport
} from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { useTable } from '@/hooks/common/table';
import MonthReportSearch from './modules/month-report-search.vue';
import MonthReportDetailModal from './modules/month-report-detail-modal.vue';
import MonthReportLockRecordsModal from './modules/month-report-lock-records-modal.vue';

defineOptions({
  name: 'MonthReport'
});

const appStore = useAppStore();

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
  apiFn: GetAttendanceMonthReport,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    yearMonth: undefined,
    orgUuid: undefined,
    attGroupName: undefined,
    name: undefined,
    onlyAbnormal: undefined,
    onlyChanged: undefined
  },
  columns: () => [
    {
      key: 'index',
      title: $t('common.index'),
      align: 'center',
      width: 64
    },
    {
      key: 'orgName',
      title: $t('page.hrm.attendance.monthReport.orgName'),
      align: 'center',
      width: 120,
      render: row => row.orgName || '-'
    },
    {
      key: 'attGroupName',
      title: $t('page.hrm.attendance.monthReport.attGroupName'),
      align: 'center',
      width: 110,
      render: row => row.attGroupName || '-'
    },
    {
      key: 'employeeName',
      title: $t('page.hrm.attendance.monthReport.employeeName'),
      align: 'center',
      width: 90
    },
    {
      key: 'expectedDays',
      title: $t('page.hrm.attendance.monthReport.expectedDays'),
      align: 'center',
      width: 84
    },
    {
      key: 'attendDays',
      title: $t('page.hrm.attendance.monthReport.attendDays'),
      align: 'center',
      width: 84
    },
    {
      key: 'lateCount',
      title: $t('page.hrm.attendance.monthReport.lateCount'),
      align: 'center',
      width: 72,
      render: row =>
        row.lateCount > 0 ? (
          <NTag type="warning" size="small">
            {row.lateCount}
          </NTag>
        ) : (
          row.lateCount
        )
    },
    {
      key: 'earlyCount',
      title: $t('page.hrm.attendance.monthReport.earlyCount'),
      align: 'center',
      width: 72,
      render: row =>
        row.earlyCount > 0 ? (
          <NTag type="warning" size="small">
            {row.earlyCount}
          </NTag>
        ) : (
          row.earlyCount
        )
    },
    {
      key: 'missingCount',
      title: $t('page.hrm.attendance.monthReport.missingCount'),
      align: 'center',
      width: 72,
      render: row =>
        row.missingCount > 0 ? (
          <NTag type="error" size="small">
            {row.missingCount}
          </NTag>
        ) : (
          row.missingCount
        )
    },
    {
      key: 'absentDays',
      title: $t('page.hrm.attendance.monthReport.absentDays'),
      align: 'center',
      width: 72,
      render: row =>
        row.absentDays > 0 ? (
          <NTag type="error" size="small">
            {row.absentDays}
          </NTag>
        ) : (
          row.absentDays
        )
    },
    {
      key: 'workdayOtH',
      title: $t('page.hrm.attendance.monthReport.workdayOtH'),
      align: 'center',
      width: 110
    },
    {
      key: 'restdayOtH',
      title: $t('page.hrm.attendance.monthReport.restdayOtH'),
      align: 'center',
      width: 110
    },
    {
      key: 'holidayOtH',
      title: $t('page.hrm.attendance.monthReport.holidayOtH'),
      align: 'center',
      width: 110
    },
    {
      key: 'compAccrueH',
      title: $t('page.hrm.attendance.monthReport.compAccrueH'),
      align: 'center',
      width: 100
    },
    {
      key: 'changed',
      title: $t('page.hrm.attendance.monthReport.lockStatus'),
      align: 'center',
      width: 110,
      render: row => renderLockStatus(row)
    },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      width: 100,
      render: row => (
        <NButton v-auth={['hrm:monthreport:detail']} type="primary" ghost size="small" onClick={() => handleDetail(row)}>
          {$t('page.hrm.attendance.monthReport.dayDetail')}
        </NButton>
      )
    }
  ]
});

// 锁定状态（列表响应携带 locked 全局状态；变更行打「锁定后变更」标记）
const locked = ref(false);
const lockHint = ref('');

function renderLockStatus(row: Api.Hrm.MonthReportRow) {
  if (row.changed) {
    return (
      <NTag type="warning" size="small">
        {$t('page.hrm.attendance.monthReport.changed')}
      </NTag>
    );
  }
  if (locked.value) {
    return (
      <NTag type="info" size="small">
        {$t('page.hrm.attendance.monthReport.lockedTag')}
      </NTag>
    );
  }
  return (
    <NTag type="default" size="small">
      {$t('page.hrm.attendance.monthReport.unlockedTag')}
    </NTag>
  );
}

// useTable apiFn 包装：同步 locked/lockHint 到页面状态
const fetchList = async (params?: any) => {
  const result = await GetAttendanceMonthReport(params);
  if (!result.error && result.data) {
    locked.value = result.data.locked;
    lockHint.value = result.data.lockHint;
  }
  return result;
};
void fetchList;

// 当月 yyyy-MM（搜索缺省）
const currentMonth = computed(() => {
  const d = new Date();
  return `${d.getFullYear()}-${`${d.getMonth() + 1}`.padStart(2, '0')}`;
});

function activeYearMonth() {
  return searchParams.yearMonth || currentMonth.value;
}

// ---- 锁定记录弹窗 ----
const lockRecordsVisible = ref(false);

// ---- 日明细弹窗 ----
const detailVisible = ref(false);
const detailRow = ref<Api.Hrm.MonthReportRow | null>(null);

function handleDetail(row: Api.Hrm.MonthReportRow) {
  detailRow.value = { ...row };
  detailVisible.value = true;
}

// ---- 锁定/解锁/重算/导出 ----
const lockLoading = ref(false);
const unlockLoading = ref(false);
const recalcLoading = ref(false);
const exporting = ref(false);

async function handleLock() {
  lockLoading.value = true;
  const { error } = await LockAttendanceMonthReport({ yearMonth: activeYearMonth() });
  lockLoading.value = false;
  if (error) return;
  window.$message?.success($t('page.hrm.attendance.monthReport.lockSuccess'));
  getData();
}

async function handleUnlock() {
  unlockLoading.value = true;
  const { error } = await UnlockAttendanceMonthReport({ yearMonth: activeYearMonth() });
  unlockLoading.value = false;
  if (error) return;
  window.$message?.success($t('page.hrm.attendance.monthReport.unlockSuccess'));
  getData();
}

async function handleRecalc() {
  recalcLoading.value = true;
  const { data, error } = await RecalcAttendanceMonthReport(activeYearMonth());
  recalcLoading.value = false;
  if (error) return;
  if (data?.locked && data.changedCount > 0) {
    window.$message?.warning($t('page.hrm.attendance.monthReport.recalcChanged').replace('{n}', `${data.changedCount}`));
  } else {
    window.$message?.success($t('page.hrm.attendance.monthReport.recalcSuccess'));
  }
  getData();
}

async function handleExport() {
  exporting.value = true;
  const { data: exportData, error } = await ExportAttendanceMonthReport({
    yearMonth: searchParams.yearMonth ?? undefined,
    orgUuid: searchParams.orgUuid ?? undefined,
    attGroupName: searchParams.attGroupName ?? undefined,
    name: searchParams.name ?? undefined,
    onlyAbnormal: searchParams.onlyAbnormal ?? undefined,
    onlyChanged: searchParams.onlyChanged ?? undefined
  });
  exporting.value = false;
  if (error || !exportData) return;
  const binary = window.atob(exportData.content);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i += 1) {
    bytes[i] = binary.charCodeAt(i);
  }
  const blob = new Blob([bytes], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = exportData.fileName;
  link.click();
  URL.revokeObjectURL(url);
}
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <MonthReportSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <NCard :bordered="false" size="small" class="card-wrapper">
      <div class="mb-8px flex items-center gap-8px">
        <NTag :type="locked ? 'info' : 'default'" size="small">
          {{ locked ? $t('page.hrm.attendance.monthReport.lockedTag') : $t('page.hrm.attendance.monthReport.unlockedTag') }}
        </NTag>
        <span v-if="lockHint" class="text-gray-400 text-12px">{{ lockHint }}</span>
      </div>
      <NSpace class="mb-8px">
        <NButton v-auth="'hrm:monthreport:lockRecords'" type="default" ghost @click="lockRecordsVisible = true">
          {{ $t('page.hrm.attendance.monthReport.lockRecords') }}
        </NButton>
        <NButton v-auth="'hrm:monthreport:recalc'" :loading="recalcLoading" type="default" ghost @click="handleRecalc">
          {{ $t('page.hrm.attendance.monthReport.recalc') }}
        </NButton>
        <NPopconfirm v-if="!locked" @positive-click="handleLock">
          <template #trigger>
            <NButton v-auth="'hrm:monthreport:lock'" :loading="lockLoading" type="primary" ghost>
              {{ $t('page.hrm.attendance.monthReport.lock') }}
            </NButton>
          </template>
          {{ $t('page.hrm.attendance.monthReport.lockConfirm') }}
        </NPopconfirm>
        <NPopconfirm v-else @positive-click="handleUnlock">
          <template #trigger>
            <NButton v-auth="'hrm:monthreport:unlock'" :loading="unlockLoading" type="warning" ghost>
              {{ $t('page.hrm.attendance.monthReport.unlock') }}
            </NButton>
          </template>
          {{ $t('page.hrm.attendance.monthReport.unlockConfirm') }}
        </NPopconfirm>
        <NButton v-auth="'hrm:monthreport:export'" :loading="exporting" type="primary" ghost @click="handleExport">
          {{ $t('page.hrm.attendance.monthReport.export') }}
        </NButton>
        <TableHeaderOperation
          v-model:columns="columnChecks"
          :loading="loading"
          @refresh="getData"
        />
      </NSpace>
      <NDataTable
        :columns="columns"
        :data="data"
        size="small"
        :flex-height="!appStore.isMobile"
        :scroll-x="1700"
        :loading="loading"
        remote
        :pagination="mobilePagination"
      />
    </NCard>
    <MonthReportDetailModal
      v-model:visible="detailVisible"
      :row="detailRow"
      :year-month="detailRow?.yearMonth"
    />
    <MonthReportLockRecordsModal v-model:visible="lockRecordsVisible" :year-month="activeYearMonth()" />
  </div>
</template>
