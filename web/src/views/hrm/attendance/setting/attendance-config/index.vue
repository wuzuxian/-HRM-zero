<script setup lang="tsx">
import { computed, reactive, ref } from 'vue';
import { NButton, NDrawer, NDrawerContent, NPopconfirm, NTag } from 'naive-ui';
import {
  DeleteAttendanceConfig,
  GetAttendanceConfigList,
  GetEffectiveAttendanceRule,
  GetEmployeeList,
  GetOrgTree
} from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { useTable } from '@/hooks/common/table';
import AttendanceConfigOperateModal from './modules/config-operate-modal.vue';
import AttendanceConfigSearch from './modules/config-search.vue';

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
  apiFn: GetAttendanceConfigList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    orgName: null
  },
  columns: () => [
    {
      key: 'index',
      title: $t('common.index'),
      align: 'center',
      width: 64
    },
    {
      key: 'orgName',
      title: $t('page.hrm.attendance.config.orgName'),
      align: 'center',
      minWidth: 140
    },
    {
      key: 'startTime',
      title: $t('page.hrm.attendance.group.workTime'),
      align: 'center',
      width: 140,
      render: row => (
        <span>
          {row.startTime} - {row.endTime}
        </span>
      )
    },
    {
      key: 'workDays',
      title: $t('page.hrm.attendance.group.workDays'),
      align: 'center',
      width: 150,
      render: row =>
        row.workDays?.length
          ? row.workDays.map(d => $t(`page.hrm.attendance.group.weekdayShort.${d}` as App.I18n.I18nKey)).join(' ')
          : '-'
    },
    {
      key: 'remark',
      title: $t('page.hrm.attendance.group.remark'),
      align: 'center',
      minWidth: 120,
      render: row => row.remark || '-'
    },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      width: 140,
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

// modal state
const operateVisible = ref(false);
const editingData = ref<Api.Hrm.AttendanceConfig | null>(null);

function handleQuickStart() {
  editingData.value = null;
  operateVisible.value = true;
}

function handleEdit(row: Api.Hrm.AttendanceConfig) {
  editingData.value = { ...row };
  operateVisible.value = true;
}

function handleSubmitted() {
  getDataByPage();
}

async function handleDelete(uuid: string) {
  const uuids: string[] = [uuid];
  deleteLoadingStatus[uuid] = true;
  const { error } = await DeleteAttendanceConfig(uuids);
  deleteLoadingStatus[uuid] = false;
  if (error) {
    return;
  }
  window.$message?.success($t('common.deleteSuccess'));
  getData();
}

// ---- 重新体检 (三件套: 逐员工生效规则查询, 发现无规则/走回退的员工) ----
type HealthRow = {
  uuid: string;
  employeeNo: string;
  name: string;
  orgName: string;
  found: boolean;
  source: string;
  sourceName: string;
  rule: string;
};

const healthVisible = ref(false);
const healthLoading = ref(false);
const healthRows = ref<HealthRow[]>([]);

const healthStats = computed(() => {
  const groupHit = healthRows.value.filter(r => r.source === 'group').length;
  const orgHit = healthRows.value.filter(r => r.found && r.source === 'org').length;
  const noRule = healthRows.value.filter(r => !r.found).length;
  return { total: healthRows.value.length, groupHit, orgHit, noRule };
});

const healthColumns = computed(() => [
  {
    key: 'name',
    title: $t('page.hrm.attendance.config.healthEmployee'),
    minWidth: 120,
    render: (row: HealthRow) => (
      <div>
        <div>{row.name}</div>
        <div class="text-12px text-gray-400">{row.employeeNo}</div>
      </div>
    )
  },
  {
    key: 'orgName',
    title: $t('page.hrm.attendance.config.orgName'),
    minWidth: 100,
    render: (row: HealthRow) => row.orgName || '-'
  },
  {
    key: 'source',
    title: $t('page.hrm.attendance.config.healthSource'),
    width: 120,
    render: (row: HealthRow) => {
      if (!row.found) {
        return (
          <NTag type="error" size="small">
            {$t('page.hrm.attendance.config.noRule')}
          </NTag>
        );
      }
      const label =
        row.source === 'group'
          ? $t('page.hrm.attendance.config.sourceGroup')
          : $t('page.hrm.attendance.config.sourceOrg');
      return (
        <NTag type={sourceTagType(row.source)} size="small">
          {label}
          {row.sourceName && row.sourceName !== '-' ? `: ${row.sourceName}` : ''}
        </NTag>
      );
    }
  },
  {
    key: 'rule',
    title: $t('page.hrm.attendance.group.workTime'),
    width: 130,
    render: (row: HealthRow) => row.rule
  }
]);

function sourceTagType(source: string) {
  if (source === 'group') return 'success' as const;
  if (source === 'org') return 'info' as const;
  return 'error' as const;
}

async function handleHealthCheck() {
  healthVisible.value = true;
  healthLoading.value = true;
  healthRows.value = [];

  const { error, data: employeeData } = await GetEmployeeList({ current: 1, size: 500 });
  if (error || !employeeData) {
    healthLoading.value = false;
    return;
  }

  // 逐员工查询生效规则 (体检为显式按钮触发的低频操作, 循环内 await 可接受)
  // eslint-disable-next-line no-await-in-loop
  const rules = await Promise.all(
    employeeData.records.map(async emp => {
      const { data: rule } = await GetEffectiveAttendanceRule({ employeeUuid: emp.uuid });
      return { emp, rule };
    })
  );
  for (const { emp, rule } of rules) {
    healthRows.value.push({
      uuid: emp.uuid,
      employeeNo: emp.employeeNo,
      name: emp.name,
      orgName: orgNameOf(emp.orgUuid),
      found: rule?.found ?? false,
      source: rule?.source ?? '',
      sourceName: rule?.sourceName || '-',
      rule: rule?.found ? `${rule.startTime} - ${rule.endTime}` : $t('page.hrm.attendance.config.noRule')
    });
  }
  healthLoading.value = false;
}

// org name resolution (employee list rows carry orgUuid)
const orgNameMap = new Map<string, string>();

function orgNameOf(orgUuid: string) {
  return orgNameMap.get(orgUuid) ?? '';
}

async function loadOrgNames() {
  const { error, data: tree } = await GetOrgTree();
  if (!error && tree) {
    const walk = (nodes: Api.Hrm.Org[]) => {
      for (const node of nodes) {
        orgNameMap.set(node.uuid, node.orgName);
        if (node.children?.length) walk(node.children);
      }
    };
    walk(tree.records);
  }
}

loadOrgNames();
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <AttendanceConfigSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <NCard
      :title="$t('page.hrm.attendance.config.title')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <NSpace>
          <NButton size="small" type="warning" ghost @click="handleHealthCheck">
            {{ $t('page.hrm.attendance.config.recheck') }}
          </NButton>
          <TableHeaderOperation
            v-model:columns="columnChecks"
            :disabled-delete="true"
            :loading="loading"
            add-auth="hrm:attendanceconfig:save"
            @add="handleQuickStart"
            @refresh="getData"
          />
        </NSpace>
      </template>
      <NAlert type="info" class="mb-12px" :show-icon="true">
        {{ $t('page.hrm.attendance.config.fallbackHint') }}
      </NAlert>
      <NDataTable
        :columns="columns"
        :data="data"
        size="small"
        :flex-height="!appStore.isMobile"
        :scroll-x="900"
        :loading="loading"
        remote
        :row-key="row => row.uuid"
        :pagination="mobilePagination"
        class="sm:h-full"
      />
      <AttendanceConfigOperateModal
        v-model:visible="operateVisible"
        :row-data="editingData"
        @submitted="handleSubmitted"
      />
      <NDrawer v-model:show="healthVisible" :width="560">
        <NDrawerContent :title="$t('page.hrm.attendance.config.recheck')" closable>
          <NSpin :show="healthLoading">
            <NSpace class="mb-12px">
              <NTag type="default">{{ $t('page.hrm.attendance.config.healthyTotal', { n: healthStats.total }) }}</NTag>
              <NTag type="success">
                {{ $t('page.hrm.attendance.config.healthyGroup', { n: healthStats.groupHit }) }}
              </NTag>
              <NTag type="info">{{ $t('page.hrm.attendance.config.healthyOrg', { n: healthStats.orgHit }) }}</NTag>
              <NTag type="error">{{ $t('page.hrm.attendance.config.healthyNoRule', { n: healthStats.noRule }) }}</NTag>
            </NSpace>
            <NDataTable
              :columns="healthColumns"
              :data="healthRows"
              size="small"
              :row-key="(row: HealthRow) => row.uuid"
              :max-height="520"
            />
          </NSpin>
        </NDrawerContent>
      </NDrawer>
    </NCard>
  </div>
</template>

<style scoped></style>
