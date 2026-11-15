<script setup lang="tsx">
import { computed, reactive, ref } from 'vue';
import { NButton, NPopconfirm, NTag } from 'naive-ui';
import { DeleteTaxProfile, ExportTaxProfile, GetTaxProfileList } from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { hrmSalaryStatusRecord, hrmTaxDeductTypeRecord, hrmTaxResidentTypeRecord } from '@/constants/business';
import { useTable } from '@/hooks/common/table';
import TaxProfileOperateModal from './modules/tax-profile-operate-modal.vue';
import TaxProfileSearch from './modules/tax-profile-search.vue';

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
  apiFn: GetTaxProfileList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    employeeName: null,
    residentType: null,
    status: null
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
      title: $t('page.hrm.salary.tax.profile.employeeName'),
      align: 'center',
      minWidth: 110
    },
    {
      key: 'idCard',
      title: $t('page.hrm.salary.tax.profile.idCard'),
      align: 'center',
      minWidth: 160
    },
    {
      key: 'residentType',
      title: $t('page.hrm.salary.tax.profile.residentType'),
      align: 'center',
      width: 110,
      render: row => $t(hrmTaxResidentTypeRecord[row.residentType])
    },
    {
      key: 'taxCity',
      title: $t('page.hrm.salary.tax.profile.taxCity'),
      align: 'center',
      minWidth: 100
    },
    {
      key: 'defaultSpecialExtra',
      title: $t('page.hrm.salary.tax.profile.defaultSpecialExtra'),
      align: 'center',
      minWidth: 180,
      render: row => {
        if (!row.defaultSpecialExtra) return '';
        return row.defaultSpecialExtra
          .split(',')
          .filter(Boolean)
          .map(code => $t(hrmTaxDeductTypeRecord[code as Api.Hrm.TaxDeductionType]))
          .join('、');
      }
    },
    {
      key: 'status',
      title: $t('page.hrm.salary.tax.profile.status'),
      align: 'center',
      width: 90,
      render: row => {
        if (row.status === null) {
          return null;
        }
        const tagMap: Record<Api.Hrm.SalarySubjectStatus, NaiveUI.ThemeColor> = {
          '1': 'success',
          '0': 'warning'
        };
        return <NTag type={tagMap[row.status]}>{$t(hrmSalaryStatusRecord[row.status])}</NTag>;
      }
    },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      width: 130,
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

// modal state (manual, same pattern as 薪酬主体 page)
const operateVisible = ref(false);
const operateType = ref<NaiveUI.TableOperateType>('add');
const editingData = ref<Api.Hrm.TaxProfile | null>(null);

function handleAdd() {
  operateType.value = 'add';
  editingData.value = null;
  operateVisible.value = true;
}

function handleEdit(row: Api.Hrm.TaxProfile) {
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
  const { error } = await DeleteTaxProfile(uuids);

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
  const { error } = await DeleteTaxProfile(uuids);
  deleteLoadingStatus[uuid] = false;
  if (error) {
    return;
  }
  window.$message?.success($t('common.deleteSuccess'));
  getData();
}

// ---- export ----
async function handleExport() {
  const { error, data: exported } = await ExportTaxProfile();
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
    <TaxProfileSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <NCard
      :title="$t('page.hrm.salary.tax.profile.title')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <TableHeaderOperation
          v-model:columns="columnChecks"
          :disabled-delete="!hasCheckedRows"
          :loading="loading"
          add-auth="hrm:tax_profile:add"
          delete-auth="hrm:tax_profile:delete"
          @add="handleAdd"
          @delete="handleBatchDelete"
          @refresh="getData"
        >
          <NButton
            v-auth="'hrm:tax_profile:export'"
            size="small"
            class="ml-8px"
            @click="handleExport"
          >
            <icon-mdi-file-export-outline class="text-icon mr-4px" />
            {{ $t('button.hrm.tax_profile.export') }}
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
      <TaxProfileOperateModal
        v-model:visible="operateVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="handleSubmitted"
      />
    </NCard>
  </div>
</template>

<style scoped></style>
