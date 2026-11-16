<script setup lang="tsx">
import { computed, ref } from 'vue';
import { NButton, NModal, NTag } from 'naive-ui';
import { GetSalaryBatchList, GetSalaryPayslipList, PublishSalaryPayslip } from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { useAuth } from '@/hooks/business/auth';
import { useTable } from '@/hooks/common/table';
import PayslipSearch from './modules/payslip-search.vue';

defineOptions({ name: 'SalaryPayslip' });

const appStore = useAppStore();
const { hasAuth } = useAuth();
const canPublish = computed(() => hasAuth('hrm:salary_payslip:publish'));

// ===== 发布弹窗（选批次 → 发布，手动 modal，不用 useTableOperate） =====
const publishVisible = ref(false);
const publishLoading = ref(false);
const publishBatchUuid = ref<string | null>(null);

const batchOptions = ref<CommonType.Option<string>[]>([]);

async function loadBatchOptions() {
  const { data, error } = await GetSalaryBatchList({ current: 1, size: 500 });
  if (!error && data) {
    batchOptions.value = data.records.map(row => ({
      label: `${row.batchNo} ${row.batchName}`,
      value: row.uuid
    }));
  }
}
loadBatchOptions();

function openPublish() {
  publishBatchUuid.value = null;
  publishVisible.value = true;
}

async function handlePublish() {
  if (!publishBatchUuid.value) {
    window.$message?.warning($t('page.hrm.salary.payslip.chooseBatchFirst'));
    return;
  }
  publishLoading.value = true;
  const { data, error } = await PublishSalaryPayslip({ batchUuid: publishBatchUuid.value });
  publishLoading.value = false;
  if (error) return;
  window.$message?.success(
    `${$t('page.hrm.salary.payslip.publishDone')} ${data?.count ?? 0}`
  );
  publishVisible.value = false;
  getData();
}

// ===== 列表 =====
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
  apiFn: GetSalaryPayslipList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    batchUuid: null,
    employeeName: null,
    yearMonth: null,
    readStatus: null
  },
  columns: () => [
    {
      key: 'index',
      title: $t('common.index'),
      align: 'center',
      width: 64
    },
    {
      key: 'title',
      title: $t('page.hrm.salary.payslip.title'),
      align: 'center',
      minWidth: 180
    },
    {
      key: 'employeeNo',
      title: $t('page.hrm.salary.payslip.employeeNo'),
      align: 'center',
      minWidth: 90
    },
    {
      key: 'employeeName',
      title: $t('page.hrm.salary.payslip.employeeName'),
      align: 'center',
      minWidth: 90
    },
    {
      key: 'yearMonth',
      title: $t('page.hrm.salary.payslip.yearMonth'),
      align: 'center',
      width: 90
    },
    {
      key: 'earningTotal',
      title: $t('page.hrm.salary.payslip.earningTotal'),
      align: 'center',
      width: 100
    },
    {
      key: 'deductionTotal',
      title: $t('page.hrm.salary.payslip.deductionTotal'),
      align: 'center',
      width: 100
    },
    {
      key: 'netTotal',
      title: $t('page.hrm.salary.payslip.netTotal'),
      align: 'center',
      width: 100
    },
    {
      key: 'published',
      title: $t('page.hrm.salary.payslip.published'),
      align: 'center',
      width: 80,
      render: row => (
        <NTag type={row.published === 1 ? 'success' : 'default'}>
          {row.published === 1 ? $t('common.yesOrNo.yes') : $t('common.yesOrNo.no')}
        </NTag>
      )
    },
    {
      key: 'readStatus',
      title: $t('page.hrm.salary.payslip.readStatus'),
      align: 'center',
      width: 80,
      render: row => (
        <NTag type={row.readStatus === 1 ? 'info' : 'warning'}>
          {row.readStatus === 1 ? $t('common.yesOrNo.yes') : $t('common.yesOrNo.no')}
        </NTag>
      )
    },
    {
      key: 'readTime',
      title: $t('page.hrm.salary.payslip.readTime'),
      align: 'center',
      width: 150,
      render: row => row.readTime || '-'
    },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      width: 100,
      render: row => (
        <NButton quaternary type="primary" size="small" onClick={() => showDetail(row)}>
          {$t('page.hrm.salary.payslip.detail')}
        </NButton>
      )
    }
  ]
});

// ===== 明细弹窗（按薪资项目 JSON 展开渲染） =====
const detailVisible = ref(false);
const detailRow = ref<Api.Hrm.SalaryPayslip | null>(null);
const detailItems = computed<Api.Hrm.SalaryPayslipItem[]>(() => detailRow.value?.items ?? []);

const detailColumns = computed(() => [
  { key: 'itemName', title: $t('page.hrm.salary.payslip.itemName'), align: 'center' as const },
  {
    key: 'itemClassText',
    title: $t('page.hrm.salary.payslip.itemClass'),
    align: 'center' as const,
    width: 90,
    render: (row: Api.Hrm.SalaryPayslipItem) => classTextOf(row)
  },
  { key: 'amount', title: $t('page.hrm.salary.payslip.amount'), align: 'center' as const, width: 110 },
  {
    key: 'sourceDetail',
    title: $t('page.hrm.salary.payslip.sourceDetail'),
    align: 'center' as const,
    render: (row: Api.Hrm.SalaryPayslipItem) => row.sourceDetail || '-'
  }
]);

const classTextRecord: Record<string, string> = {
  earning: $t('page.hrm.salary.item.itemClass.earning'),
  deduction: $t('page.hrm.salary.item.itemClass.deduction'),
  cost: $t('page.hrm.salary.item.itemClass.cost')
};

function detailRowKey(row: Api.Hrm.SalaryPayslipItem) {
  return row.itemCode;
}

function classTextOf(row: Api.Hrm.SalaryPayslipItem) {
  return classTextRecord[row.itemClass] || row.itemClass;
}

function showDetail(row: Api.Hrm.SalaryPayslip) {
  detailRow.value = row;
  detailVisible.value = true;
}
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <PayslipSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <NCard
      :title="$t('page.hrm.salary.payslip.pageTitle')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <NSpace align="center" :wrap="false">
          <NButton v-if="canPublish" size="small" ghost type="primary" @click="openPublish">
            <template #icon>
              <icon-ic-round-plus class="text-icon" />
            </template>
            {{ $t('page.hrm.salary.payslip.publish') }}
          </NButton>
          <TableHeaderOperation v-model:columns="columnChecks" :loading="loading" @refresh="getData" />
        </NSpace>
      </template>
      <NDataTable
        :columns="columns"
        :data="data"
        size="small"
        :flex-height="!appStore.isMobile"
        :scroll-x="1400"
        :loading="loading"
        remote
        :row-key="row => row.uuid"
        :pagination="mobilePagination"
        class="sm:h-full"
      />
    </NCard>

    <NModal v-model:show="publishVisible" preset="card" :title="$t('page.hrm.salary.payslip.publish')" class="w-520px">
      <NFormItem :label="$t('page.hrm.salary.payslip.chooseBatch')" label-placement="left" :label-width="90">
        <NSelect
          v-model:value="publishBatchUuid"
          :options="batchOptions"
          :placeholder="$t('page.hrm.salary.payslip.chooseBatchPlaceholder')"
          filterable
          clearable
        />
      </NFormItem>
      <div class="flex justify-end">
        <NButton type="primary" :loading="publishLoading" @click="handlePublish">
          {{ $t('page.hrm.salary.payslip.publish') }}
        </NButton>
      </div>
    </NModal>

    <NModal v-model:show="detailVisible" preset="card" :title="detailRow?.title" class="w-680px">
      <NDescriptions :column="2" size="small" label-placement="left" bordered>
        <NDescriptionsItem :label="$t('page.hrm.salary.payslip.employeeName')">
          {{ detailRow?.employeeName }}（{{ detailRow?.employeeNo }}）
        </NDescriptionsItem>
        <NDescriptionsItem :label="$t('page.hrm.salary.payslip.yearMonth')">{{ detailRow?.yearMonth }}</NDescriptionsItem>
        <NDescriptionsItem :label="$t('page.hrm.salary.payslip.earningTotal')">
          {{ detailRow?.earningTotal }}
        </NDescriptionsItem>
        <NDescriptionsItem :label="$t('page.hrm.salary.payslip.deductionTotal')">
          {{ detailRow?.deductionTotal }}
        </NDescriptionsItem>
        <NDescriptionsItem :label="$t('page.hrm.salary.payslip.netTotal')" :span="2">
          {{ detailRow?.netTotal }}
        </NDescriptionsItem>
      </NDescriptions>
      <NDataTable
        class="mt-12px"
        size="small"
        :columns="detailColumns"
        :data="detailItems"
        :row-key="detailRowKey"
        :pagination="false"
      />
    </NModal>
  </div>
</template>

<style scoped></style>
