<script setup lang="tsx">
import { computed, reactive, ref } from 'vue';
import { NButton, NTag } from 'naive-ui';
import { GetLeaveBalanceList, GetOrgTree, InitLeaveBalance, PreviewLeaveBalanceInit } from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { useTable } from '@/hooks/common/table';
import LeaveBalanceSearch from './modules/leave-balance-search.vue';
import LeaveBalanceOperateModal from './modules/leave-balance-operate-modal.vue';

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
  apiFn: GetLeaveBalanceList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    year: null,
    name: null,
    orgUuid: null,
    leaveTypeCode: null
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
      title: $t('page.hrm.attendance.leaveBalance.employeeNo'),
      align: 'center',
      width: 100
    },
    {
      key: 'employeeName',
      title: $t('page.hrm.attendance.leaveBalance.employeeName'),
      align: 'center',
      width: 100
    },
    {
      key: 'orgName',
      title: $t('page.hrm.attendance.leaveBalance.orgName'),
      align: 'center',
      minWidth: 120,
      render: row => row.orgName || '-'
    },
    {
      key: 'hireDate',
      title: $t('page.hrm.attendance.leaveBalance.hireDate'),
      align: 'center',
      width: 110
    },
    {
      key: 'year',
      title: $t('page.hrm.attendance.leaveBalance.year'),
      align: 'center',
      width: 80
    },
    {
      key: 'leaveTypeName',
      title: $t('page.hrm.attendance.leaveBalance.leaveTypeName'),
      align: 'center',
      width: 100
    },
    {
      key: 'available',
      title: $t('page.hrm.attendance.leaveBalance.available'),
      align: 'center',
      width: 90,
      render: row => (
        <NTag type={row.available > 0 ? 'success' : 'warning'} size="small">
          {row.available}
        </NTag>
      )
    },
    {
      key: 'granted',
      title: $t('page.hrm.attendance.leaveBalance.granted'),
      align: 'center',
      width: 90
    },
    {
      key: 'used',
      title: $t('page.hrm.attendance.leaveBalance.used'),
      align: 'center',
      width: 80
    },
    {
      key: 'frozen',
      title: $t('page.hrm.attendance.leaveBalance.frozen'),
      align: 'center',
      width: 80,
      render: row => row.frozen || 0
    },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      width: 150,
      render: row => (
        <div class="flex-center gap-8px">
          <NButton
            v-auth={['hrm:leavebalance:adjust']}
            type="primary"
            ghost
            size="small"
            onClick={() => handleAdjust(row)}
          >
            {$t('page.hrm.attendance.leaveBalance.adjustTitle')}
          </NButton>
          <NButton type="default" ghost size="small" onClick={() => handleViewLog(row)}>
            {$t('page.hrm.attendance.leaveBalance.logTitle')}
          </NButton>
        </div>
      )
    }
  ]
});

// modal state (manual, HRM 页面不用 useTableOperate)
const operateVisible = ref(false);
const operateType = ref<'adjust' | 'log'>('adjust');
const editingRow = ref<Api.Hrm.LeaveBalance | null>(null);

function handleAdjust(row: Api.Hrm.LeaveBalance) {
  operateType.value = 'adjust';
  editingRow.value = { ...row };
  operateVisible.value = true;
}

function handleViewLog(row: Api.Hrm.LeaveBalance) {
  operateType.value = 'log';
  editingRow.value = { ...row };
  operateVisible.value = true;
}

function handleSubmitted() {
  getData();
}

// ---- 初始化本年度额度 modal ----
const initVisible = ref(false);
const initSubmitting = ref(false);

const initModel = reactive({
  year: new Date().getFullYear() as number | null,
  employeeUuids: [] as string[]
});

const initYearOptions = computed(() => {
  const current = new Date().getFullYear();
  const options: { label: string; value: number }[] = [];
  for (let y = current + 1; y >= current - 5; y -= 1) {
    options.push({ label: `${y}`, value: y });
  }
  return options;
});

interface PreviewItem {
  employeeUuid: string;
  employeeNo: string;
  employeeName: string;
  leaveTypeName: string;
  defaultDays: number;
  skipped: number;
}

const previewRows = ref<PreviewItem[]>([]);
const previewLoading = ref(false);

async function handleOpenInit() {
  initVisible.value = true;
  previewRows.value = [];
  await loadPreview();
}

async function loadPreview() {
  if (!initModel.year) return;
  previewLoading.value = true;
  const { data: previewData, error } = await PreviewLeaveBalanceInit({
    year: initModel.year,
    employeeUuids: initModel.employeeUuids.length > 0 ? initModel.employeeUuids : undefined
  });
  previewLoading.value = false;
  if (error) return;
  previewRows.value = previewData?.records ?? [];
}

async function loadPreviewOnYearChange() {
  if (initVisible.value) {
    await loadPreview();
  }
}

const previewColumns = computed(() => [
  { key: 'employeeNo', title: $t('page.hrm.attendance.leaveBalance.employeeNo'), align: 'center' as const, width: 100 },
  {
    key: 'employeeName',
    title: $t('page.hrm.attendance.leaveBalance.employeeName'),
    align: 'center' as const,
    width: 100
  },
  {
    key: 'leaveTypeName',
    title: $t('page.hrm.attendance.leaveBalance.leaveTypeName'),
    align: 'center' as const,
    width: 100
  },
  {
    key: 'defaultDays',
    title: $t('page.hrm.attendance.leaveBalance.defaultDays'),
    align: 'center' as const,
    width: 100
  },
  {
    key: 'skipped',
    title: $t('page.hrm.attendance.leaveBalance.skipped'),
    align: 'center' as const,
    width: 100,
    render: (row: PreviewItem) =>
      row.skipped === 1 ? (
        <NTag type="default" size="small">
          {$t('page.hrm.attendance.leaveBalance.skippedYes')}
        </NTag>
      ) : (
        <NTag type="success" size="small">
          {$t('page.hrm.attendance.leaveBalance.skippedNo')}
        </NTag>
      )
  }
]);

async function handleInit() {
  if (!initModel.year) {
    window.$message?.error($t('page.hrm.attendance.leaveBalance.form.year'));
    return;
  }
  initSubmitting.value = true;
  const { error } = await InitLeaveBalance({
    year: initModel.year,
    employeeUuids: initModel.employeeUuids.length > 0 ? initModel.employeeUuids : undefined
  });
  initSubmitting.value = false;
  if (error) return;
  window.$message?.success($t('page.hrm.attendance.leaveBalance.initSuccess'));
  initVisible.value = false;
  getData();
}

// ---- org tree (筛选: 部门含下级) ----
interface OrgOption {
  label: string;
  value: string;
  children?: OrgOption[];
}

const orgOptions = ref<OrgOption[]>([]);

function toOrgOptions(nodes: Api.Hrm.Org[]): OrgOption[] {
  return nodes.map(node => ({
    label: node.orgName,
    value: node.uuid,
    children: node.children && node.children.length > 0 ? toOrgOptions(node.children) : undefined
  }));
}

async function loadOrgOptions() {
  if (orgOptions.value.length > 0) return;
  const { data: orgData, error } = await GetOrgTree();
  if (!error && orgData) {
    orgOptions.value = toOrgOptions(orgData.records ?? []);
  }
}
loadOrgOptions();
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <LeaveBalanceSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <NCard
      :title="$t('page.hrm.attendance.leaveBalance.title')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <div class="flex items-center gap-8px">
          <NButton v-auth="'hrm:leavebalance:init'" type="primary" @click="handleOpenInit">
            {{ $t('page.hrm.attendance.leaveBalance.initTitle') }}
          </NButton>
          <TableHeaderOperation v-model:columns="columnChecks" :loading="loading" @refresh="getData" />
        </div>
      </template>
      <NDataTable
        :columns="columns"
        :data="data"
        size="small"
        :flex-height="!appStore.isMobile"
        :scroll-x="1200"
        :loading="loading"
        remote
        :pagination="mobilePagination"
      />
    </NCard>
    <LeaveBalanceOperateModal
      v-model:visible="operateVisible"
      :operate-type="operateType"
      :row-data="editingRow"
      @submitted="handleSubmitted"
    />
    <NModal
      v-model:show="initVisible"
      :title="$t('page.hrm.attendance.leaveBalance.initTitle')"
      preset="card"
      class="w-860px"
    >
      <NForm label-placement="left" :label-width="100">
        <NFormItem :label="$t('page.hrm.attendance.leaveBalance.year')" path="year">
          <NSelect v-model:value="initModel.year" :options="initYearOptions" @update:value="loadPreviewOnYearChange" />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.leaveBalance.initScope')" path="employeeUuids">
          <span class="text-gray-500">{{ $t('page.hrm.attendance.leaveBalance.initScopeHint') }}</span>
        </NFormItem>
      </NForm>
      <NDataTable
        :columns="previewColumns"
        :data="previewRows"
        :loading="previewLoading"
        size="small"
        :max-height="300"
      />
      <template #footer>
        <NSpace justify="end">
          <NButton @click="initVisible = false">{{ $t('common.cancel') }}</NButton>
          <NButton type="primary" :loading="initSubmitting" @click="handleInit">
            {{ $t('common.confirm') }}
          </NButton>
        </NSpace>
      </template>
    </NModal>
  </div>
</template>
