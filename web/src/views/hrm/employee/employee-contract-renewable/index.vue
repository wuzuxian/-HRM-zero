<script setup lang="tsx">
import { computed, reactive, ref } from 'vue';
import { NButton, NPopconfirm, NTag } from 'naive-ui';
import { GetEmployeeContractList, GetOrgTree, RenewEmployeeContract } from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { useTable } from '@/hooks/common/table';
import ContractSearch from './modules/contract-search.vue';

defineOptions({
  name: 'EmployeeContractRenewable'
});

const appStore = useAppStore();

type RenewLoadingStatus = Record<string, boolean>;
const renewLoadingStatus = reactive<RenewLoadingStatus>({});

function expiryStateTag(row: Api.Hrm.HrmEmployeeContract) {
  const tagMap: Record<string, NaiveUI.ThemeColor> = {
    normal: 'default',
    warning: 'warning',
    expired: 'error',
    terminated: 'info'
  };
  const labelMap: Record<string, App.I18n.I18nKey> = {
    normal: 'page.hrm.employee.contractRenewable.stateNormal',
    warning: 'page.hrm.employee.contractRenewable.stateWarning',
    expired: 'page.hrm.employee.contractRenewable.stateExpired',
    terminated: 'page.hrm.employee.contractRenewable.stateTerminated'
  };
  return <NTag type={tagMap[row.expiryState || 'normal']}>{$t(labelMap[row.expiryState || 'normal'])}</NTag>;
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
  // 待续签口径（spec 用户故事 15 / demo-inventory §2.3）：默认只看即将到期(warning)，
  // 用户可切换到 已到期(expired) 或已解除(terminated) 口径复查。
  apiParams: {
    current: 1,
    size: 10,
    employeeName: null,
    employeeNo: null,
    startDate: null,
    endDate: null,
    expiryState: 'warning'
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
      title: $t('page.hrm.employee.contractRenewable.employeeNo'),
      align: 'center',
      width: 100
    },
    {
      key: 'employeeName',
      title: $t('page.hrm.employee.contractRenewable.employeeName'),
      align: 'center',
      width: 90
    },
    {
      key: 'orgUuid',
      title: $t('page.hrm.employee.contractRenewable.orgName'),
      align: 'center',
      width: 130,
      render: row => orgNameByUuid(row.orgUuid) || '-'
    },
    {
      key: 'startDate',
      title: $t('page.hrm.employee.contractRenewable.startDate'),
      align: 'center',
      width: 110,
      render: row => row.startDate || '-'
    },
    {
      key: 'endDate',
      title: $t('page.hrm.employee.contractRenewable.endDate'),
      align: 'center',
      width: 110,
      render: row => row.endDate || '-'
    },
    {
      key: 'daysToExpiry',
      title: $t('page.hrm.employee.contractRenewable.daysToExpiry'),
      align: 'center',
      width: 110,
      render: row => (row.expiryState === 'terminated' ? '-' : String(row.daysToExpiry))
    },
    {
      key: 'renewCount',
      title: $t('page.hrm.employee.contractRenewable.renewCount'),
      align: 'center',
      width: 90
    },
    {
      key: 'expiryState',
      title: $t('page.hrm.employee.contractRenewable.status'),
      align: 'center',
      width: 100,
      render: row => expiryStateTag(row)
    },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      width: 120,
      render: row => renderOperate(row)
    }
  ]
});

function renderOperate(row: Api.Hrm.HrmEmployeeContract) {
  if (row.expiryState === 'terminated') return null;
  return (
    <NPopconfirm onPositiveClick={() => handleRenew(row)}>
      {{
        default: () => $t('page.hrm.employee.contractRenewable.confirmRenew'),
        trigger: () => (
          <NButton loading={renewLoadingStatus[row.uuid]} type="info" ghost size="small">
            {$t('page.hrm.employee.contractRenewable.renew')}
          </NButton>
        )
      }}
    </NPopconfirm>
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

async function handleRenew(row: Api.Hrm.HrmEmployeeContract) {
  renewLoadingStatus[row.uuid] = true;
  const { error } = await RenewEmployeeContract({ uuid: row.uuid });
  renewLoadingStatus[row.uuid] = false;
  if (error) return;
  window.$message?.success($t('page.hrm.employee.contractRenewable.renewSuccess'));
  getData();
}
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <ContractSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <NCard
      :title="$t('page.hrm.employee.contractRenewable.title')"
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
        :scroll-x="1100"
        :loading="loading"
        remote
        :row-key="row => row.uuid"
        :pagination="mobilePagination"
        class="sm:h-full"
      />
    </NCard>
  </div>
</template>

<style scoped></style>
