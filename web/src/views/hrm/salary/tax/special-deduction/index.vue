<script setup lang="tsx">
import { computed, reactive, ref } from 'vue';
import { NButton, NPopconfirm, NTag } from 'naive-ui';
import { DeleteTaxSpecialDeduction, ExportTaxSpecialDeduction, GetTaxSpecialDeductionList } from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { hrmTaxDeductTypeRecord, hrmTaxDeductionSourceRecord } from '@/constants/business';
import { useTable } from '@/hooks/common/table';
import TaxDeductionMaintainModal from './modules/tax-deduction-maintain-modal.vue';
import TaxDeductionMaintainYearModal from './modules/tax-deduction-maintain-year-modal.vue';
import TaxDeductionImportModal from './modules/tax-deduction-import-modal.vue';
import TaxDeductionSearch from './modules/tax-deduction-search.vue';

const appStore = useAppStore();

type LoadingStatus = Record<string, boolean>;
const deleteLoadingStatus = reactive<LoadingStatus>({});

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
  apiFn: GetTaxSpecialDeductionList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    employeeName: null,
    deductYear: null,
    deductType: null,
    source: null
  },
  columns: () => [
    {
      type: 'selection',
      align: 'center',
      width: 48
    },
    {
      key: 'index',
      title: $t('common.index'),
      align: 'center',
      width: 64
    },
    {
      key: 'employeeName',
      title: $t('page.hrm.salary.tax.deduction.employeeName'),
      align: 'center',
      minWidth: 100
    },
    {
      key: 'idCard',
      title: $t('page.hrm.salary.tax.deduction.idCard'),
      align: 'center',
      minWidth: 160
    },
    {
      key: 'deductYear',
      title: $t('page.hrm.salary.tax.deduction.deductYear'),
      align: 'center',
      width: 90
    },
    {
      key: 'deductMonth',
      title: $t('page.hrm.salary.tax.deduction.deductMonth'),
      align: 'center',
      width: 70
    },
    {
      key: 'deductType',
      title: $t('page.hrm.salary.tax.deduction.deductTypeTitle'),
      align: 'center',
      minWidth: 130,
      render: row => $t(hrmTaxDeductTypeRecord[row.deductType])
    },
    {
      key: 'amount',
      title: $t('page.hrm.salary.tax.deduction.amount'),
      align: 'center',
      width: 110,
      render: row => row.amount.toFixed(2)
    },
    {
      key: 'source',
      title: $t('page.hrm.salary.tax.deduction.sourceTitle'),
      align: 'center',
      width: 100,
      render: row => (
        <NTag type={row.source === 'import' ? 'info' : 'default'}>{$t(hrmTaxDeductionSourceRecord[row.source])}</NTag>
      )
    },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      width: 140,
      render: row => (
        <div class="flex-center gap-8px">
          <NButton type="primary" ghost size="small" onClick={() => handleMaintain(row)}>
            {$t('page.hrm.salary.tax.deduction.maintain')}
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

// ---- maintain modal (single employee + year; 12-month grid) ----
const maintainVisible = ref(false);
const maintainEmployeeUuid = ref('');

function handleMaintain(row: Api.Hrm.TaxSpecialDeductionRow) {
  maintainEmployeeUuid.value = row.employeeUuid;
  maintainVisible.value = true;
}

function handleAddMaintain() {
  maintainEmployeeUuid.value = '';
  maintainVisible.value = true;
}

// ---- maintain year modal ----
const maintainYearVisible = ref(false);

// ---- import modal ----
const importVisible = ref(false);

function handleSubmitted() {
  getDataByPage();
}

async function handleBatchDelete() {
  const uuids: string[] = [...checkedRowKeys.value];

  uuids.forEach(uuid => {
    deleteLoadingStatus[uuid] = true;
  });
  const { error } = await DeleteTaxSpecialDeduction(uuids);

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
  const { error } = await DeleteTaxSpecialDeduction(uuids);
  deleteLoadingStatus[uuid] = false;
  if (error) {
    return;
  }
  window.$message?.success($t('common.deleteSuccess'));
  getData();
}

// ---- export ----
async function handleExport() {
  const { error, data: exported } = await ExportTaxSpecialDeduction();
  if (error || !exported) return;
  const bin = atob(exported.content);
  const bytes = new Uint8Array(bin.length);
  for (let i = 0; i < bin.length; i += 1) {
    bytes[i] = bin.charCodeAt(i);
  }
  const blob = new Blob([bytes], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' });
  const anchor = document.createElement('a');
  anchor.href = URL.createObjectURL(blob);
  anchor.download = exported.fileName;
  anchor.click();
  URL.revokeObjectURL(anchor.href);
}
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <TaxDeductionSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <NCard
      :title="$t('page.hrm.salary.tax.deduction.title')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <TableHeaderOperation
          v-model:columns="columnChecks"
          :disabled-delete="!hasCheckedRows"
          :loading="loading"
          add-auth="hrm:tax_special_deduction:maintain"
          delete-auth="hrm:tax_special_deduction:delete"
          @add="handleAddMaintain"
          @delete="handleBatchDelete"
          @refresh="getData"
        >
          <NButton
            v-auth="'hrm:tax_special_deduction:maintain_year'"
            size="small"
            class="ml-8px"
            @click="maintainYearVisible = true"
          >
            <icon-mdi-calendar-edit class="mr-4px text-icon" />
            {{ $t('page.hrm.salary.tax.deduction.maintainYear') }}
          </NButton>
          <NButton
            v-auth="'hrm:tax_special_deduction:import'"
            size="small"
            class="ml-8px"
            @click="importVisible = true"
          >
            <icon-mdi-file-upload-outline class="mr-4px text-icon" />
            {{ $t('page.hrm.salary.tax.deduction.import') }}
          </NButton>
          <NButton v-auth="'hrm:tax_special_deduction:export'" size="small" class="ml-8px" @click="handleExport">
            <icon-mdi-file-export-outline class="mr-4px text-icon" />
            {{ $t('button.hrm.tax_special_deduction.export') }}
          </NButton>
        </TableHeaderOperation>
      </template>
      <NDataTable
        v-model:checked-row-keys="checkedRowKeys"
        :columns="columns"
        :data="data"
        size="small"
        :flex-height="!appStore.isMobile"
        :scroll-x="1300"
        :loading="loading"
        remote
        :row-key="row => row.uuid"
        :pagination="mobilePagination"
        class="sm:h-full"
      />
      <TaxDeductionMaintainModal
        v-model:visible="maintainVisible"
        :employee-uuid="maintainEmployeeUuid"
        @submitted="handleSubmitted"
      />
      <TaxDeductionMaintainYearModal v-model:visible="maintainYearVisible" @submitted="handleSubmitted" />
      <TaxDeductionImportModal v-model:visible="importVisible" @submitted="handleSubmitted" />
    </NCard>
  </div>
</template>

<style scoped></style>
