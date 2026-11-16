<script setup lang="tsx">
import { computed, ref } from 'vue';
import { NButton, NTag } from 'naive-ui';
import { GetPunchStatistics } from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { useTable } from '@/hooks/common/table';
import PunchStatisticsSearch from './modules/punch-statistics-search.vue';
import PunchStatisticsDetailModal from './modules/punch-statistics-detail-modal.vue';

defineOptions({
  name: 'PunchStatistics'
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
  apiFn: GetPunchStatistics,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    startDate: undefined,
    endDate: undefined,
    name: undefined,
    onlyAbnormal: undefined
  },
  columns: () => [
    {
      key: 'index',
      title: $t('common.index'),
      align: 'center',
      width: 64
    },
    {
      key: 'employeeName',
      title: $t('page.hrm.attendance.punchStatistics.employeeName'),
      align: 'center',
      width: 100
    },
    {
      key: 'orgName',
      title: $t('page.hrm.attendance.punchStatistics.orgName'),
      align: 'center',
      width: 120,
      render: row => row.orgName || '-'
    },
    {
      key: 'expectedDays',
      title: $t('page.hrm.attendance.punchStatistics.expectedDays'),
      align: 'center',
      width: 90
    },
    {
      key: 'attendDays',
      title: $t('page.hrm.attendance.punchStatistics.attendDays'),
      align: 'center',
      width: 90
    },
    {
      key: 'absentDays',
      title: $t('page.hrm.attendance.punchStatistics.absentDays'),
      align: 'center',
      width: 90,
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
      key: 'lateCount',
      title: $t('page.hrm.attendance.punchStatistics.lateCount'),
      align: 'center',
      width: 80,
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
      title: $t('page.hrm.attendance.punchStatistics.earlyCount'),
      align: 'center',
      width: 80,
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
      title: $t('page.hrm.attendance.punchStatistics.missingCount'),
      align: 'center',
      width: 80,
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
      key: 'abnormalTotal',
      title: $t('page.hrm.attendance.punchStatistics.abnormalTotal'),
      align: 'center',
      width: 90,
      render: row =>
        row.abnormalTotal > 0 ? (
          <NTag type="error" size="small">
            {row.abnormalTotal}
          </NTag>
        ) : (
          row.abnormalTotal
        )
    },
    {
      key: 'abnormalSummary',
      title: $t('page.hrm.attendance.punchStatistics.abnormalSummary'),
      align: 'center',
      minWidth: 160,
      render: row => row.abnormalSummary || '-'
    },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      width: 100,
      render: row => (
        <NButton v-auth={['hrm:punch:statisticsDetail']} type="primary" ghost size="small" onClick={() => handleDetail(row)}>
          {$t('page.hrm.attendance.punchStatistics.viewDetail')}
        </NButton>
      )
    }
  ]
});

// 统计卡（随列表响应更新；卡片与分页无关，对筛选后的员工集合聚合）。
const cards = ref<Api.Hrm.PunchStatsCard>({
  statEmployees: 0,
  abnormalEmployees: 0,
  lateCount: 0,
  earlyCount: 0,
  missingCount: 0,
  recordCount: 0
});

const statCardItems = computed(() => [
  { label: $t('page.hrm.attendance.punchStatistics.cardEmployees'), value: cards.value.statEmployees },
  { label: $t('page.hrm.attendance.punchStatistics.cardAbnormalEmployees'), value: cards.value.abnormalEmployees },
  { label: $t('page.hrm.attendance.punchStatistics.cardLate'), value: cards.value.lateCount },
  { label: $t('page.hrm.attendance.punchStatistics.cardEarly'), value: cards.value.earlyCount },
  { label: $t('page.hrm.attendance.punchStatistics.cardMissing'), value: cards.value.missingCount },
  { label: $t('page.hrm.attendance.punchStatistics.cardRecords'), value: cards.value.recordCount }
]);

async function fetchCards() {
  const { data: statsData, error } = await GetPunchStatistics({
    current: 1,
    size: 1,
    startDate: searchParams.startDate,
    endDate: searchParams.endDate,
    name: searchParams.name,
    onlyAbnormal: searchParams.onlyAbnormal
  });
  if (error || !statsData) return;
  cards.value = statsData.cards;
}

function refresh() {
  fetchCards();
  getData();
}

// 查看明细弹窗（手动管理 modal，HRM 页面不用 useTableOperate）。
const detailVisible = ref(false);
const detailPerson = ref<Api.Hrm.PunchPersonStats | null>(null);

function handleDetail(row: Api.Hrm.PunchPersonStats) {
  detailPerson.value = { ...row };
  detailVisible.value = true;
}
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <PunchStatisticsSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <NCard :bordered="false" size="small" class="card-wrapper">
      <NGrid cols="2 s:3 m:6" responsive="screen" :x-gap="8" :y-gap="8">
        <NGi v-for="item in statCardItems" :key="item.label">
          <NStatistic :label="item.label" :value="item.value" />
        </NGi>
      </NGrid>
    </NCard>
    <NCard
      :title="$t('page.hrm.attendance.punchStatistics.title')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <div class="flex items-center gap-8px">
          <NButton type="primary" ghost @click="refresh">
            {{ $t('common.refresh') }}
          </NButton>
          <TableHeaderOperation
            v-model:columns="columnChecks"
            :loading="loading"
            @refresh="refresh"
          />
        </div>
      </template>
      <NDataTable
        :columns="columns"
        :data="data"
        size="small"
        :flex-height="!appStore.isMobile"
        :scroll-x="1300"
        :loading="loading"
        remote
        :pagination="mobilePagination"
      />
    </NCard>
    <PunchStatisticsDetailModal
      v-model:visible="detailVisible"
      :person="detailPerson"
      :start-date="searchParams.startDate ?? undefined"
      :end-date="searchParams.endDate ?? undefined"
    />
  </div>
</template>
