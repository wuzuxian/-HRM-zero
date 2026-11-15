<script setup lang="tsx">
import { computed, reactive, ref } from 'vue';
import dayjs from 'dayjs';
import { NButton, NPopconfirm, NTag } from 'naive-ui';
import { DeleteEmployeeContract, GetEmployeeContractList, GetOrgTree, RenewEmployeeContract } from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { useTable } from '@/hooks/common/table';
import ContractOperateModal from './modules/contract-operate-modal.vue';
import ContractSearch from './modules/contract-search.vue';

defineOptions({
  name: 'EmployeeContractExpiring'
});

const appStore = useAppStore();

type LoadingStatus = Record<string, boolean>;
const deleteLoadingStatus = reactive<LoadingStatus>({});
const renewLoadingStatus = reactive<LoadingStatus>({});

function expiryStateTag(row: Api.Hrm.HrmEmployeeContract) {
  const tagMap: Record<string, NaiveUI.ThemeColor> = {
    normal: 'default',
    warning: 'warning',
    expired: 'error',
    terminated: 'info'
  };
  const labelMap: Record<string, App.I18n.I18nKey> = {
    normal: 'page.hrm.employee.contractExpiring.stateNormal',
    warning: 'page.hrm.employee.contractExpiring.stateWarning',
    expired: 'page.hrm.employee.contractExpiring.stateExpired',
    terminated: 'page.hrm.employee.contractExpiring.stateTerminated'
  };
  const key = labelMap[row.expiryState || 'normal'];
  return <NTag type={tagMap[row.expiryState || 'normal']}>{$t(key)}</NTag>;
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
  apiFn: GetEmployeeContractList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    employeeName: null,
    employeeNo: null,
    startDate: null,
    endDate: null,
    expiryState: null
  },
  columns: () => [
    {
      key: 'index',
      title: $t('common.index'),
      align: 'center',
      width: 64
    },
    {
      key: 'employeeNo',
      title: $t('page.hrm.employee.contractExpiring.employeeNo'),
      align: 'center',
      width: 100
    },
    {
      key: 'employeeName',
      title: $t('page.hrm.employee.contractExpiring.employeeName'),
      align: 'center',
      width: 90
    },
    {
      key: 'orgUuid',
      title: $t('page.hrm.employee.contractExpiring.orgName'),
      align: 'center',
      width: 130,
      render: row => orgNameByUuid(row.orgUuid) || '-'
    },
    {
      key: 'startDate',
      title: $t('page.hrm.employee.contractExpiring.startDate'),
      align: 'center',
      width: 110,
      render: row => row.startDate || '-'
    },
    {
      key: 'endDate',
      title: $t('page.hrm.employee.contractExpiring.endDate'),
      align: 'center',
      width: 110,
      render: row => row.endDate || '-'
    },
    {
      key: 'daysToExpiry',
      title: $t('page.hrm.employee.contractExpiring.daysToExpiry'),
      align: 'center',
      width: 110,
      render: row => (row.expiryState === 'terminated' ? '-' : String(row.daysToExpiry))
    },
    {
      key: 'renewCount',
      title: $t('page.hrm.employee.contractExpiring.renewCount'),
      align: 'center',
      width: 90
    },
    {
      key: 'expiryState',
      title: $t('page.hrm.employee.contractExpiring.status'),
      align: 'center',
      width: 100,
      render: row => expiryStateTag(row)
    },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      width: 200,
      render: row => renderOperate(row)
    }
  ]
});

function renderOperate(row: Api.Hrm.HrmEmployeeContract) {
  return (
    <div class="flex-center flex-wrap gap-4px">
      {row.expiryState !== 'terminated' && (
        <NPopconfirm onPositiveClick={() => handleRenew(row)}>
          {{
            default: () =>
              $t('page.hrm.employee.contractExpiring.confirmRenew').replace(
                '{endDate}',
                dayjs(row.endDate).add(1, 'year').format('YYYY-MM-DD')
              ),
            trigger: () => (
              <NButton loading={renewLoadingStatus[row.uuid]} type="info" ghost size="small">
                {$t('page.hrm.employee.contractExpiring.renew')}
              </NButton>
            )
          }}
        </NPopconfirm>
      )}
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
  );
}

// ---- org tree (部门名称解析, 同 员工档案 页) ----
const orgOptions = ref<Api.Hrm.Org[]>([]);

async function loadOrgTree() {
  const { error, data: tree } = await GetOrgTree();
  if (!error && tree) {
    orgOptions.value = tree.records;
  }
}

const orgFlatMap = computed<Map<string, string>>(() => {
  const map = new Map<string, string>();
  const walk = (nodes: Api.Hrm.Org[]) => {
    for (const node of nodes) {
      map.set(node.uuid, node.orgName);
      if (node.children?.length) walk(node.children);
    }
  };
  walk(orgOptions.value);
  return map;
});

function orgNameByUuid(uuid: string) {
  return orgFlatMap.value.get(uuid) ?? '';
}

loadOrgTree();

// modal state (manual, same pattern as 组织管理 page)
const operateVisible = ref(false);
const operateType = ref<NaiveUI.TableOperateType>('add');
const editingData = ref<Api.Hrm.HrmEmployeeContract | null>(null);

function handleAdd() {
  operateType.value = 'add';
  editingData.value = null;
  operateVisible.value = true;
}

function handleEdit(row: Api.Hrm.HrmEmployeeContract) {
  operateType.value = 'edit';
  editingData.value = { ...row };
  operateVisible.value = true;
}

function handleSubmitted() {
  getDataByPage();
}

async function handleRenew(row: Api.Hrm.HrmEmployeeContract) {
  renewLoadingStatus[row.uuid] = true;
  const { error } = await RenewEmployeeContract({ uuid: row.uuid });
  renewLoadingStatus[row.uuid] = false;
  if (error) return;
  window.$message?.success($t('page.hrm.employee.contractExpiring.renewSuccess'));
  getData();
}

async function handleBatchDelete() {
  const uuids: string[] = [...checkedRowKeys.value];
  uuids.forEach(uuid => {
    deleteLoadingStatus[uuid] = true;
  });
  const { error } = await DeleteEmployeeContract(uuids);
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
  const { error } = await DeleteEmployeeContract(uuids);
  deleteLoadingStatus[uuid] = false;
  if (error) return;
  window.$message?.success($t('common.deleteSuccess'));
  getData();
}

const checkedRowKeys = ref<string[]>([]);
const hasCheckedRows = computed(() => checkedRowKeys.value.length === 0);
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <ContractSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <NCard
      :title="$t('page.hrm.employee.contractExpiring.title')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <TableHeaderOperation
          v-model:columns="columnChecks"
          :disabled-delete="!hasCheckedRows"
          :loading="loading"
          add-auth="hrm:contract:add"
          delete-auth="hrm:contract:delete"
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
        :scroll-x="1200"
        :loading="loading"
        remote
        :row-key="row => row.uuid"
        :pagination="mobilePagination"
        class="sm:h-full"
      />
      <ContractOperateModal
        v-model:visible="operateVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="handleSubmitted"
      />
    </NCard>
  </div>
</template>

<style scoped></style>
