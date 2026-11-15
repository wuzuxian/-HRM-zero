<script setup lang="tsx">
import { computed, reactive, ref } from 'vue';
import { NButton, NPopconfirm, NTag } from 'naive-ui';
import { DeleteSalaryArchive, GetSalaryArchiveList } from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { hrmSalaryStatusRecord } from '@/constants/business';
import { useTable } from '@/hooks/common/table';
import SalaryArchiveOperateModal from './modules/salary-archive-operate-modal.vue';
import SalaryArchiveSearch from './modules/salary-archive-search.vue';

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
  apiFn: GetSalaryArchiveList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    employeeName: null,
    employeeNo: null,
    subjectUuid: null,
    schemeUuid: null,
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
      title: $t('page.hrm.salary.archive.employeeName'),
      align: 'center',
      minWidth: 90
    },
    {
      key: 'employeeNo',
      title: $t('page.hrm.salary.archive.employeeNo'),
      align: 'center',
      minWidth: 90
    },
    {
      key: 'companyName',
      title: $t('page.hrm.salary.archive.companyName'),
      align: 'center',
      minWidth: 100
    },
    {
      key: 'orgName',
      title: $t('page.hrm.salary.archive.orgName'),
      align: 'center',
      minWidth: 100
    },
    {
      key: 'subjectName',
      title: $t('page.hrm.salary.archive.subject'),
      align: 'center',
      minWidth: 110
    },
    {
      key: 'schemeName',
      title: $t('page.hrm.salary.archive.scheme'),
      align: 'center',
      minWidth: 110,
      render: row => row.schemeName || '-'
    },
    {
      key: 'status',
      title: $t('page.hrm.salary.archive.status'),
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
      key: 'fixedSalary',
      title: $t('page.hrm.salary.archive.fixedSalary'),
      align: 'center',
      width: 110
    },
    {
      key: 'fixedAllowance',
      title: $t('page.hrm.salary.archive.fixedAllowance'),
      align: 'center',
      width: 110
    },
    {
      key: 'effectiveDate',
      title: $t('page.hrm.salary.archive.effectiveDate'),
      align: 'center',
      width: 110
    },
    {
      key: 'expireDate',
      title: $t('page.hrm.salary.archive.expireDate'),
      align: 'center',
      width: 110,
      render: row => row.expireDate || $t('page.hrm.salary.archive.longTerm')
    },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      width: 150,
      render: row => (
        <div class="flex-center gap-8px">
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

// modal state (manual, same pattern as 薪酬方案 page)
const operateVisible = ref(false);
const operateType = ref<NaiveUI.TableOperateType>('add');
const editingData = ref<Api.Hrm.SalaryArchive | null>(null);

function handleAdd() {
  operateType.value = 'add';
  editingData.value = null;
  operateVisible.value = true;
}

function handleEdit(row: Api.Hrm.SalaryArchive) {
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
  const { error } = await DeleteSalaryArchive(uuids);

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
  const { error } = await DeleteSalaryArchive(uuids);
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
    <SalaryArchiveSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <NCard
      :title="$t('page.hrm.salary.archive.title')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <TableHeaderOperation
          v-model:columns="columnChecks"
          :disabled-delete="!hasCheckedRows"
          :loading="loading"
          add-auth="hrm:salary_archive:add"
          delete-auth="hrm:salary_archive:delete"
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
        :scroll-x="1400"
        :loading="loading"
        remote
        :row-key="row => row.uuid"
        :pagination="mobilePagination"
        class="sm:h-full"
      />
      <SalaryArchiveOperateModal
        v-model:visible="operateVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="handleSubmitted"
      />
    </NCard>
  </div>
</template>

<style scoped></style>
