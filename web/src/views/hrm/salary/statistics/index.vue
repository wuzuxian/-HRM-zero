<script setup lang="tsx">
import { computed, ref } from 'vue';
import { NTag } from 'naive-ui';
import { GetSalaryLedger } from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { useTable } from '@/hooks/common/table';
import LedgerSearch from './modules/ledger-search.vue';

defineOptions({ name: 'SalaryStatistics' });

const appStore = useAppStore();

// ===== 台账表（useTable 管理分页筛选；apiFn 包装同步更新统计卡） =====

const statCards = ref({
  employeeCount: 0,
  earningTotal: 0,
  deductionTotal: 0,
  netTotal: 0,
  laborCostTotal: 0
});

// GetSalaryLedger 返回 {records,...} 之外还带统计卡字段；包装成 useTable 需要的分页形状
async function ledgerTableApiFn(params: Api.Hrm.SalaryLedgerSearchParams) {
  const res = await GetSalaryLedger(params);
  if (!res.error && res.data) {
    statCards.value = {
      employeeCount: res.data.employeeCount,
      earningTotal: res.data.earningTotal,
      deductionTotal: res.data.deductionTotal,
      netTotal: res.data.netTotal,
      laborCostTotal: res.data.laborCostTotal
    };
  }
  return res as { data: Api.Common.PaginatingQueryRecord<Api.Hrm.SalaryLedgerRow> | null; error: unknown };
}

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
  apiFn: ledgerTableApiFn,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    batchUuid: null,
    employeeName: null,
    orgName: null,
    onlyAbnormal: null
  },
  columns: () => [
    {
      key: 'index',
      title: $t('common.index'),
      align: 'center',
      width: 64
    },
    {
      key: 'batchNo',
      title: $t('page.hrm.salary.statistics.batch'),
      align: 'center',
      minWidth: 140,
      render: row => `${row.batchNo} ${row.batchName}`
    },
    {
      key: 'employeeName',
      title: $t('page.hrm.salary.statistics.employee'),
      align: 'center',
      minWidth: 110,
      render: row => `${row.employeeName}（${row.employeeNo}）`
    },
    {
      key: 'yearMonth',
      title: $t('page.hrm.salary.statistics.yearMonth'),
      align: 'center',
      width: 90
    },
    {
      key: 'orgName',
      title: $t('page.hrm.salary.statistics.org'),
      align: 'center',
      minWidth: 100
    },
    {
      key: 'earningTotal',
      title: $t('page.hrm.salary.statistics.earningTotal'),
      align: 'center',
      width: 100
    },
    {
      key: 'deductionTotal',
      title: $t('page.hrm.salary.statistics.deductionTotal'),
      align: 'center',
      width: 100
    },
    {
      key: 'netTotal',
      title: $t('page.hrm.salary.statistics.netTotal'),
      align: 'center',
      width: 100
    },
    {
      key: 'laborCostTotal',
      title: $t('page.hrm.salary.statistics.laborCostTotal'),
      align: 'center',
      width: 110
    },
    {
      key: 'abnormal',
      title: $t('page.hrm.salary.statistics.abnormal'),
      align: 'center',
      minWidth: 160,
      render: row =>
        row.abnormal ? (
          <NTag type="error">{row.abnormal}</NTag>
        ) : (
          <NTag type="default">{$t('page.hrm.salary.statistics.normal')}</NTag>
        )
    },
    {
      key: 'paid',
      title: $t('page.hrm.salary.statistics.paid'),
      align: 'center',
      width: 80,
      render: row => (row.paid === 1 ? $t('common.yesOrNo.yes') : $t('common.yesOrNo.no'))
    },
    {
      key: 'payslipped',
      title: $t('page.hrm.salary.statistics.payslipped'),
      align: 'center',
      width: 80,
      render: row => (row.payslipped === 1 ? $t('common.yesOrNo.yes') : $t('common.yesOrNo.no'))
    }
  ]
});

const statCardItems = computed(() => [
  { label: $t('page.hrm.salary.statistics.employeeCount'), value: statCards.value.employeeCount },
  { label: $t('page.hrm.salary.statistics.earningTotal'), value: statCards.value.earningTotal },
  { label: $t('page.hrm.salary.statistics.deductionTotal'), value: statCards.value.deductionTotal },
  { label: $t('page.hrm.salary.statistics.netTotal'), value: statCards.value.netTotal },
  { label: $t('page.hrm.salary.statistics.laborCostTotal'), value: statCards.value.laborCostTotal }
]);
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <LedgerSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />

    <NGrid cols="2 s:3 m:5" responsive="screen" :x-gap="12" :y-gap="12">
      <NGridItem v-for="card in statCardItems" :key="card.label">
        <NCard :bordered="false" size="small" class="card-wrapper">
          <div class="text-14px text-gray-500">{{ card.label }}</div>
          <div class="mt-8px text-22px font-bold">{{ card.value }}</div>
        </NCard>
      </NGridItem>
    </NGrid>

    <NCard
      :title="$t('page.hrm.salary.statistics.pageTitle')"
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
        :scroll-x="1500"
        :loading="loading"
        remote
        :row-key="row => `${row.batchUuid}-${row.employeeUuid}`"
        :pagination="mobilePagination"
        class="sm:h-full"
      />
    </NCard>
  </div>
</template>

<style scoped></style>
