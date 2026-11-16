<script setup lang="tsx">
import { computed, reactive, ref } from 'vue';
import { NButton, NPopconfirm, NTag } from 'naive-ui';
import { DeleteSalaryBatch, GetSalaryBatchList, TrialSalaryBatch } from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { hrmSalaryStatusRecord } from '@/constants/business';
import { useTable } from '@/hooks/common/table';
import SalaryBatchOperateModal from './modules/salary-batch-operate-modal.vue';
import SalaryBatchResultModal from './modules/salary-batch-result-modal.vue';
import SalaryBatchSearch from './modules/salary-batch-search.vue';

const appStore = useAppStore();

type LoadingStatus = Record<string, boolean>;
const deleteLoadingStatus = reactive<LoadingStatus>({});
const trialLoadingStatus = reactive<LoadingStatus>({});

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
  apiFn: GetSalaryBatchList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    batchName: null,
    subjectUuid: null,
    batchType: null,
    year: null,
    month: null,
    trialStatus: null
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
      title: $t('page.hrm.salary.batch.batchNo'),
      align: 'center',
      minWidth: 140
    },
    {
      key: 'batchName',
      title: $t('page.hrm.salary.batch.batchName'),
      align: 'center',
      minWidth: 120
    },
    {
      key: 'subjectName',
      title: $t('page.hrm.salary.batch.subject'),
      align: 'center',
      minWidth: 110
    },
    {
      key: 'batchYear',
      title: $t('page.hrm.salary.batch.year'),
      align: 'center',
      width: 80
    },
    {
      key: 'batchMonth',
      title: $t('page.hrm.salary.batch.month'),
      align: 'center',
      width: 80
    },
    {
      key: 'batchType',
      title: $t('page.hrm.salary.batch.batchType'),
      align: 'center',
      width: 100,
      render: row => $t(`page.hrm.salary.batch.batchTypeEnum.${row.batchType}` as App.I18n.I18nKey)
    },
    {
      key: 'headcount',
      title: $t('page.hrm.salary.batch.headcount'),
      align: 'center',
      width: 80
    },
    {
      key: 'earningTotal',
      title: $t('page.hrm.salary.batch.earningTotal'),
      align: 'center',
      width: 110
    },
    {
      key: 'deductionTotal',
      title: $t('page.hrm.salary.batch.deductionTotal'),
      align: 'center',
      width: 110
    },
    {
      key: 'netTotal',
      title: $t('page.hrm.salary.batch.netTotal'),
      align: 'center',
      width: 110
    },
    {
      key: 'trialStatus',
      title: $t('page.hrm.salary.batch.progress'),
      align: 'center',
      width: 180,
      render: row => {
        const tags: { done: boolean; label: string }[] = [
          { done: row.trialStatus === '1', label: $t('page.hrm.salary.batch.progressEnum.trial') },
          { done: row.confirmStatus === '1', label: $t('page.hrm.salary.batch.progressEnum.confirm') },
          { done: row.payStatus === '1', label: $t('page.hrm.salary.batch.progressEnum.pay') },
          { done: row.payslipStatus === '1', label: $t('page.hrm.salary.batch.progressEnum.payslip') }
        ];
        return (
          <div class="flex-center gap-4px">
            {tags.map(item => (
              <NTag size="small" type={item.done ? 'success' : 'default'} bordered={false}>
                {item.label}
                {item.done ? '✓' : ''}
              </NTag>
            ))}
          </div>
        );
      }
    },
    {
      key: 'trialTime',
      title: $t('page.hrm.salary.batch.trialTime'),
      align: 'center',
      width: 150,
      render: row => row.trialTime || '-'
    },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      width: 220,
      render: row => (
        <div class="flex-center gap-8px">
          <NButton
            quaternary
            type="primary"
            size="small"
            disabled={row.trialStatus !== '1'}
            onClick={() => handleViewResult(row)}
          >
            {$t('page.hrm.salary.batch.viewResult')}
          </NButton>
          <NButton
            quaternary
            type="warning"
            size="small"
            loading={trialLoadingStatus[row.uuid]}
            onClick={() => handleTrial(row)}
          >
            {$t('page.hrm.salary.batch.trial')}
          </NButton>
          <NButton quaternary type="primary" size="small" onClick={() => handleEdit(row)}>
            {$t('common.edit')}
          </NButton>
          <NPopconfirm onPositiveClick={() => handleDelete(row.uuid)}>
            {{
              default: () => $t('common.confirmDelete'),
              trigger: () => (
                <NButton loading={deleteLoadingStatus[row.uuid]} type="error" ghost size="small">
                  {$t('common.delete')}
                </NButton>
              )
            }}
          </NPopconfirm>
        </div>
      )
    }
  ]
});

const checkedRowKeys = ref<string[]>([]);
const hasCheckedRows = computed(() => checkedRowKeys.value.length === 0);

// modal state (manual, same pattern as 员工薪资档案 page)
const operateVisible = ref(false);
const operateType = ref<NaiveUI.TableOperateType>('add');
const editingData = ref<Api.Hrm.SalaryBatch | null>(null);

// 查看结果弹窗
const resultVisible = ref(false);
const resultBatchUuid = ref('');

function handleAdd() {
  operateType.value = 'add';
  editingData.value = null;
  operateVisible.value = true;
}

function handleEdit(row: Api.Hrm.SalaryBatch) {
  operateType.value = 'edit';
  editingData.value = { ...row };
  operateVisible.value = true;
}

function handleViewResult(row: Api.Hrm.SalaryBatch) {
  resultBatchUuid.value = row.uuid;
  resultVisible.value = true;
}

async function handleTrial(row: Api.Hrm.SalaryBatch) {
  trialLoadingStatus[row.uuid] = true;
  const { data, error } = await TrialSalaryBatch(row.uuid);
  trialLoadingStatus[row.uuid] = false;
  if (error) return;
  window.$message?.success(
    $t('page.hrm.salary.batch.trialDone').replace('{count}', String(data?.headcount ?? 0))
  );
  if (data?.warnings) {
    window.$message?.warning(data.warnings.split('\n')[0], { duration: 6000 });
  }
  getData();
}

function handleSubmitted() {
  getDataByPage();
}

async function handleBatchDelete() {
  const uuids: string[] = [...checkedRowKeys.value];

  uuids.forEach(uuid => {
    deleteLoadingStatus[uuid] = true;
  });
  const { error } = await DeleteSalaryBatch(uuids);

  uuids.forEach(uuid => {
    deleteLoadingStatus[uuid] = false;
  });

  if (error) return;

  checkedRowKeys.value = [];
  window.$message?.success($t('common.deleteSuccess'));
  getData();
}

async function handleDelete(uuid: string) {
  const uuids: string[] = [uuid];
  deleteLoadingStatus[uuid] = true;
  const { error } = await DeleteSalaryBatch(uuids);
  deleteLoadingStatus[uuid] = false;
  if (error) {
    return;
  }
  window.$message?.success($t('common.deleteSuccess'));
  getData();
}
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <SalaryBatchSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <NCard
      :title="$t('page.hrm.salary.batch.title')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <TableHeaderOperation
          v-model:columns="columnChecks"
          :disabled-delete="!hasCheckedRows"
          :loading="loading"
          add-auth="hrm:salary_batch:add"
          delete-auth="hrm:salary_batch:delete"
          @add="handleAdd"
          @delete="handleBatchDelete"
          @refresh="getData"
        />
      </template>
      <NDataTable
        v-model:checked-row-keys="checkedRowKeys"
        :columns="columns"
        :data="data"
        size="small"
        :flex-height="!appStore.isMobile"
        :scroll-x="1700"
        :loading="loading"
        remote
        :row-key="row => row.uuid"
        :pagination="mobilePagination"
        class="sm:h-full"
      />
      <SalaryBatchOperateModal
        v-model:visible="operateVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="handleSubmitted"
      />
      <SalaryBatchResultModal v-model:visible="resultVisible" :batch-uuid="resultBatchUuid" />
    </NCard>
  </div>
</template>

<style scoped></style>
