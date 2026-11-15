<script setup lang="tsx">
import { computed, reactive, ref } from 'vue';
import { NButton, NPopconfirm, NTag } from 'naive-ui';
import { DeleteEmployee, ExportEmployee, GenerateEmployeeUser, GetEmployeeDetail, GetEmployeeList, GetOrgTree } from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import {
  hrmEmployeeGenderRecord,
  hrmEmployeePersonStatusRecord,
  hrmEmployeeUserStatusRecord
} from '@/constants/business';
import { useTable } from '@/hooks/common/table';
import { useAuth } from '@/hooks/business/auth';
import EmployeeOperateModal from './modules/employee-operate-modal.vue';
import EmployeeImportModal from './modules/employee-import-modal.vue';
import EmployeeSearch from './modules/employee-search.vue';

const appStore = useAppStore();
const { hasAuth } = useAuth();

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
  apiFn: GetEmployeeList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    employeeNo: null,
    name: null,
    orgUuid: null,
    personStatus: null
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
      key: 'employeeNo',
      title: $t('page.hrm.employee.employeeNo'),
      align: 'center',
      width: 110
    },
    {
      key: 'name',
      title: $t('page.hrm.employee.name'),
      align: 'center',
      width: 100
    },
    {
      key: 'gender',
      title: $t('page.hrm.employee.genderLabel'),
      align: 'center',
      width: 70,
      render: row => (row.gender ? $t(hrmEmployeeGenderRecord[row.gender as Api.Hrm.Gender]) : '-')
    },
    {
      key: 'orgUuid',
      title: $t('page.hrm.employee.orgName'),
      align: 'center',
      minWidth: 120,
      render: row => orgNameByUuid(row.orgUuid) || '-'
    },
    {
      key: 'position',
      title: $t('page.hrm.employee.position'),
      align: 'center',
      width: 110,
      render: row => row.position || '-'
    },
    {
      key: 'post',
      title: $t('page.hrm.employee.post'),
      align: 'center',
      width: 110,
      render: row => row.post || '-'
    },
    {
      key: 'personStatus',
      title: $t('page.hrm.employee.personStatusLabel'),
      align: 'center',
      width: 100,
      render: row => {
        if (!row.personStatus) return null;
        const tagMap: Record<Api.Hrm.PersonStatus, NaiveUI.ThemeColor> = {
          '1': 'success',
          '2': 'info',
          '3': 'info',
          '4': 'warning',
          '5': 'error',
          '6': 'default'
        };
        return (
          <NTag type={tagMap[row.personStatus as Api.Hrm.PersonStatus]} size="small">
            {$t(hrmEmployeePersonStatusRecord[row.personStatus as Api.Hrm.PersonStatus])}
          </NTag>
        );
      }
    },
    {
      key: 'hireDate',
      title: $t('page.hrm.employee.hireDate'),
      align: 'center',
      width: 110,
      render: row => row.hireDate || '-'
    },
    {
      key: 'formalDate',
      title: $t('page.hrm.employee.formalDate'),
      align: 'center',
      width: 110,
      render: row => row.formalDate || '-'
    },
    {
      key: 'companyUuid',
      title: $t('page.hrm.employee.companyName'),
      align: 'center',
      minWidth: 110,
      render: row => orgNameByUuid(row.companyUuid) || '-'
    },
    {
      key: 'phone',
      title: $t('page.hrm.employee.phone'),
      align: 'center',
      width: 130,
      render: row => row.phone || '-'
    },
    {
      key: 'email',
      title: $t('page.hrm.employee.email'),
      align: 'center',
      minWidth: 140,
      render: row => row.email || '-'
    },
    {
      key: 'userStatus',
      title: $t('page.hrm.employee.userStatusLabel'),
      align: 'center',
      width: 100,
      render: row =>
        row.userStatus ? (
          <NTag type={row.userStatus === '1' ? 'success' : 'warning'} size="small">
            {$t(hrmEmployeeUserStatusRecord[row.userStatus as Api.Hrm.UserStatus])}
          </NTag>
        ) : null
    },
    {
      key: 'createTime',
      title: $t('page.hrm.employee.createTime'),
      align: 'center',
      width: 160
    },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      width: 160,
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

// ---- org tree (部门/单位名称解析) ----
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

function loadOrgOptions() {
  return orgOptions.value.flatMap(node => flattenOrg(node));
}

function flattenOrg(node: Api.Hrm.Org): Api.Hrm.Org[] {
  return [node, ...(node.children?.length ? node.children.flatMap(flattenOrg) : [])];
}

loadOrgTree();

// ---- modal state (manual, same pattern as 班次管理 page) ----
const operateVisible = ref(false);
const operateType = ref<NaiveUI.TableOperateType>('add');
const editingData = ref<Api.Hrm.Employee | null>(null);

function handleAdd() {
  operateType.value = 'add';
  editingData.value = null;
  operateVisible.value = true;
}

async function handleEdit(row: Api.Hrm.Employee) {
  // 详情接口带三个从表, 编辑前拉全量
  const { error, data: detail } = await GetEmployeeDetail(row.uuid);
  operateType.value = 'edit';
  editingData.value = error || !detail ? { ...row } : detail.employee;
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
  const { error } = await DeleteEmployee(uuids);

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
  const { error } = await DeleteEmployee(uuids);
  deleteLoadingStatus[uuid] = false;
  if (error) {
    return;
  }
  window.$message?.success($t('common.deleteSuccess'));
  getData();
}

// ---- import modal ----
const importVisible = ref(false);

// ---- 生成系统账号 (HRM-04): 对勾选的员工批量生成, 未勾选则提示 ----
const generateUserLoading = ref(false);

async function handleGenerateUser() {
  const uuids: string[] = [...checkedRowKeys.value];
  if (uuids.length === 0) {
    window.$message?.warning($t('page.hrm.employee.generateUserSelectTip'));
    return;
  }
  generateUserLoading.value = true;
  const { error, data } = await GenerateEmployeeUser({ uuids });
  generateUserLoading.value = false;
  if (error) return;
  if (data && data.generated > 0) {
    window.$message?.success($t('page.hrm.employee.generateUserSuccess', { generated: data.generated, skipped: data.skipped }));
  } else {
    window.$message?.info($t('page.hrm.employee.generateUserAllSkipped', { skipped: data?.skipped ?? 0 }));
  }
  checkedRowKeys.value = [];
  getData();
}

// ---- export ----
async function handleExport() {
  const { error, data: exported } = await ExportEmployee();
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
    <EmployeeSearch
      v-model:model="searchParams"
      :org-options="loadOrgOptions()"
      @reset="resetSearchParams"
      @search="getDataByPage"
    />
    <NCard :title="$t('page.hrm.employee.title')" :bordered="false" size="small" class="sm:flex-1-hidden card-wrapper">
      <template #header-extra>
        <NSpace :size="12" class="mr-12px" align="center">
          <NButton v-if="hasAuth('hrm:employee:add')" size="small" type="primary" @click="handleAdd">
            <template #icon>
              <icon-ic-round-plus class="text-icon" />
            </template>
            {{ $t('page.hrm.employee.addEmployee') }}
          </NButton>
          <NButton
            v-if="hasAuth('hrm:employee:import')"
            size="small"
            ghost
            type="primary"
            @click="importVisible = true"
          >
            {{ $t('page.hrm.employee.importEmployee') }}
          </NButton>
          <NButton v-if="hasAuth('hrm:employee:export')" size="small" ghost type="primary" @click="handleExport">
            {{ $t('page.hrm.employee.exportEmployee') }}
          </NButton>
          <NButton
            v-if="hasAuth('hrm:employee:generateuser')"
            size="small"
            ghost
            type="warning"
            :loading="generateUserLoading"
            @click="handleGenerateUser"
          >
            {{ $t('page.hrm.employee.generateUser') }}
          </NButton>
        </NSpace>
        <TableHeaderOperation
          v-model:columns="columnChecks"
          :disabled-delete="checkedRowKeys.length === 0"
          :loading="loading"
          delete-auth="hrm:employee:delete"
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
        :scroll-x="1800"
        :loading="loading"
        remote
        :row-key="row => row.uuid"
        :pagination="mobilePagination"
        class="sm:h-full"
      />
      <EmployeeOperateModal
        v-model:visible="operateVisible"
        :operate-type="operateType"
        :row-data="editingData"
        :org-options="loadOrgOptions()"
        @submitted="handleSubmitted"
      />
      <EmployeeImportModal v-model:visible="importVisible" @submitted="getDataByPage" />
    </NCard>
  </div>
</template>

<style scoped></style>
