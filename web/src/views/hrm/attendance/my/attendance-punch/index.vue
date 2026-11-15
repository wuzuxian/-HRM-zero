<script setup lang="tsx">
import { ref } from 'vue';
import { NButton, NCard, NTag } from 'naive-ui';
import { GetMyPunchList, Punch } from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { useTable } from '@/hooks/common/table';

defineOptions({
  name: 'AttendancePunch'
});

const appStore = useAppStore();

const { columns, columnChecks, data, getData, loading, mobilePagination, searchParams, resetSearchParams } = useTable({
  apiFn: GetMyPunchList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    startDate: undefined,
    endDate: undefined,
    clockInStatus: undefined
  },
  columns: () => [
    {
      key: 'punchDate',
      title: $t('page.hrm.attendance.punchRecord.punchDate'),
      align: 'center',
      width: 110
    },
    {
      key: 'weekday',
      title: $t('page.hrm.attendance.punchRecord.weekday'),
      align: 'center',
      width: 70
    },
    {
      key: 'clockInTime',
      title: $t('page.hrm.attendance.punchRecord.clockInTime'),
      align: 'center',
      width: 160,
      render: row => row.clockInTime || '-'
    },
    {
      key: 'clockInStatus',
      title: $t('page.hrm.attendance.punchRecord.clockInStatus'),
      align: 'center',
      width: 90,
      render: row => renderStatus(row.clockInStatus)
    },
    {
      key: 'clockInLocation',
      title: $t('page.hrm.attendance.punchRecord.clockInLocation'),
      align: 'center',
      minWidth: 130,
      render: row => row.clockInLocation || '-'
    },
    {
      key: 'clockOutTime',
      title: $t('page.hrm.attendance.punchRecord.clockOutTime'),
      align: 'center',
      width: 160,
      render: row => row.clockOutTime || '-'
    },
    {
      key: 'clockOutStatus',
      title: $t('page.hrm.attendance.punchRecord.clockOutStatus'),
      align: 'center',
      width: 90,
      render: row => renderStatus(row.clockOutStatus)
    },
    {
      key: 'settleStatus',
      title: $t('page.hrm.attendance.punchRecord.settleStatus'),
      align: 'center',
      width: 100,
      render: row => renderSettle(row.settleStatus)
    },
    {
      key: 'overtime',
      title: $t('page.hrm.attendance.punchRecord.overtime'),
      align: 'center',
      width: 70,
      render: row =>
        row.overtime === '1'
          ? $t('page.hrm.attendance.punchRecord.overtimeYes')
          : $t('page.hrm.attendance.punchRecord.overtimeNo')
    }
  ]
});

function renderStatus(status: string) {
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

// ---- 打卡动作 ----
const clocking = ref(false);

async function handlePunch(type: 'clockIn' | 'clockOut') {
  clocking.value = true;
  const { error } = await Punch({ type });
  clocking.value = false;
  if (error) return;
  window.$message?.success(
    type === 'clockIn'
      ? $t('page.hrm.attendance.attendancePunch.clockInSuccess')
      : $t('page.hrm.attendance.attendancePunch.clockOutSuccess')
  );
  getData();
}
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <NCard :bordered="false" size="small" class="card-wrapper">
      <div class="flex items-center justify-between py-12px">
        <span class="text-16px font-bold">{{ $t('page.hrm.attendance.attendancePunch.title') }}</span>
        <NSpace>
          <NButton
            v-auth="'hrm:punch:punch'"
            type="primary"
            size="large"
            :loading="clocking"
            @click="handlePunch('clockIn')"
          >
            {{ $t('page.hrm.attendance.attendancePunch.clockIn') }}
          </NButton>
          <NButton
            v-auth="'hrm:punch:punch'"
            type="warning"
            size="large"
            :loading="clocking"
            @click="handlePunch('clockOut')"
          >
            {{ $t('page.hrm.attendance.attendancePunch.clockOut') }}
          </NButton>
        </NSpace>
      </div>
    </NCard>
    <NCard
      :title="$t('page.hrm.attendance.attendancePunch.listTitle')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <TableHeaderOperation v-model:columns="columnChecks" :loading="loading" @refresh="getData" />
      </template>
      <NDataTable
        :columns="columns"
        :data="data"
        size="small"
        :flex-height="!appStore.isMobile"
        :scroll-x="1000"
        :loading="loading"
        remote
        :pagination="mobilePagination"
      />
    </NCard>
  </div>
</template>
