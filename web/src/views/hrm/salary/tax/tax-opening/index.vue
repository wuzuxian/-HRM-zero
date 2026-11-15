<script setup lang="tsx">
import { computed, reactive, ref } from 'vue';
import { NButton, NPopconfirm } from 'naive-ui';
import { DeleteTaxOpening, ExportTaxOpening, GetTaxOpeningList } from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { useTable } from '@/hooks/common/table';
import TaxOpeningOperateModal from './modules/tax-opening-operate-modal.vue';
import TaxOpeningImportModal from './modules/tax-opening-import-modal.vue';
import TaxOpeningSearch from './modules/tax-opening-search.vue';

const appStore = useAppStore();

type LoadingStatus = Record<string, boolean>;
const deleteLoadingStatus = reactive<LoadingStatus>({});

function money(v: number): string {
  return v == null ? '' : v.toFixed(2);
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
  apiFn: GetTaxOpeningList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    employeeName: null,
    subjectUuid: null,
    taxYear: null
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
      title: $t('page.hrm.salary.tax.opening.employeeName'),
      align: 'center',
      minWidth: 100
    },
    {
      key: 'idCard',
      title: $t('page.hrm.salary.tax.opening.idCard'),
      align: 'center',
      minWidth: 160
    },
    {
      key: 'subjectName',
      title: $t('page.hrm.salary.tax.opening.subjectName'),
      align: 'center',
      minWidth: 130,
      render: row => row.subjectName || $t('page.hrm.salary.tax.opening.form.subjectName').replace(/请选择/, '')
    },
    {
      key: 'taxYear',
      title: $t('page.hrm.salary.tax.opening.taxYear'),
      align: 'center',
      width: 80
    },
    {
      key: 'untilMonth',
      title: $t('page.hrm.salary.tax.opening.untilMonth'),
      align: 'center',
      width: 80
    },
    {
      key: 'incomeTotal',
      title: $t('page.hrm.salary.tax.opening.incomeTotal'),
      align: 'center',
      width: 120,
      render: row => money(row.incomeTotal)
    },
    {
      key: 'specialDeductionTotal',
      title: $t('page.hrm.salary.tax.opening.specialDeductionTotal'),
      align: 'center',
      width: 130,
      render: row => money(row.specialDeductionTotal)
    },
    {
      key: 'specialExtraTotal',
      title: $t('page.hrm.salary.tax.opening.specialExtraTotal'),
      align: 'center',
      width: 130,
      render: row => money(row.specialExtraTotal)
    },
    {
      key: 'exemptTotal',
      title: $t('page.hrm.salary.tax.opening.exemptTotal'),
      align: 'center',
      width: 110,
      render: row => money(row.exemptTotal)
    },
    {
      key: 'otherDeductionTotal',
      title: $t('page.hrm.salary.tax.opening.otherDeductionTotal'),
      align: 'center',
      width: 130,
      render: row => money(row.otherDeductionTotal)
    },
    {
      key: 'taxableIncomeTotal',
      title: $t('page.hrm.salary.tax.opening.taxableIncomeTotal'),
      align: 'center',
      width: 140,
      render: row => money(row.taxableIncomeTotal)
    },
    {
      key: 'taxPaidTotal',
      title: $t('page.hrm.salary.tax.opening.taxPaidTotal'),
      align: 'center',
      width: 130,
      render: row => money(row.taxPaidTotal)
    },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      width: 130,
      fixed: 'right',
      render: row => (
        <div class="flex-center gap-8px">
          <NButton type="primary" ghost size="small" onClick={() => handleEdit(row)}>
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

// modal state (manual)
const operateVisible = ref(false);
const operateType = ref<NaiveUI.TableOperateType>('add');
const editingData = ref<Api.Hrm.TaxOpening | null>(null);
const importVisible = ref(false);

function handleAdd() {
  operateType.value = 'add';
  editingData.value = null;
  operateVisible.value = true;
}

function handleEdit(row: Api.Hrm.TaxOpening) {
  operateType.value = 'edit';
  editingData.value = { ...row };
  operateVisible.value = true;
}

function handleSubmitted() {
  getDataByPage();
}

async function handleBatchDelete() {
  const uuids: string[] = [...checkedRowKeys.value];

  uuids.forEach(uuid => {
    deleteLoadingStatus[uuid] = true;
  });
  const { error } = await DeleteTaxOpening(uuids);

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
  const { error } = await DeleteTaxOpening(uuids);
  deleteLoadingStatus[uuid] = false;
  if (error) {
    return;
  }
  window.$message?.success($t('common.deleteSuccess'));
  getData();
}

// ---- export ----
async function handleExport() {
  const { error, data: exported } = await ExportTaxOpening();
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
    <TaxOpeningSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <NCard
      :title="$t('page.hrm.salary.tax.opening.title')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <TableHeaderOperation
          v-model:columns="columnChecks"
          :disabled-delete="!hasCheckedRows"
          :loading="loading"
          add-auth="hrm:tax_opening:add"
          delete-auth="hrm:tax_opening:delete"
          @add="handleAdd"
          @delete="handleBatchDelete"
          @refresh="getData"
        >
          <NButton v-auth="'hrm:tax_opening:import'" size="small" class="ml-8px" @click="importVisible = true">
            <icon-mdi-file-upload-outline class="mr-4px text-icon" />
            {{ $t('page.hrm.salary.tax.opening.import') }}
          </NButton>
          <NButton v-auth="'hrm:tax_opening:export'" size="small" class="ml-8px" @click="handleExport">
            <icon-mdi-file-export-outline class="mr-4px text-icon" />
            {{ $t('button.hrm.tax_opening.export') }}
          </NButton>
        </TableHeaderOperation>
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
      <TaxOpeningOperateModal
        v-model:visible="operateVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="handleSubmitted"
      />
      <TaxOpeningImportModal v-model:visible="importVisible" @submitted="handleSubmitted" />
    </NCard>
  </div>
</template>

<style scoped></style>
