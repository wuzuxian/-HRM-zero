<script setup lang="tsx">
import { computed, reactive, ref } from 'vue';
import { NButton, NPopconfirm, NTag } from 'naive-ui';
import { DeleteOvertimeRule, GetOvertimeRuleList, ToggleOvertimeRuleStatus } from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { hrmAttendanceStatusRecord } from '@/constants/business';
import { useTable } from '@/hooks/common/table';
import OvertimeRuleOperateModal from './modules/overtime-rule-operate-modal.vue';
import OvertimeRuleSearch from './modules/overtime-rule-search.vue';

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
  apiFn: GetOvertimeRuleList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    ruleName: null,
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
      key: 'ruleName',
      title: $t('page.hrm.attendance.overtimeRule.ruleName'),
      align: 'center',
      minWidth: 140
    },
    {
      key: 'workday',
      title: $t('page.hrm.attendance.overtimeRule.workday'),
      align: 'center',
      minWidth: 200,
      render: row => renderPolicy(row.workday)
    },
    {
      key: 'restday',
      title: $t('page.hrm.attendance.overtimeRule.restday'),
      align: 'center',
      minWidth: 200,
      render: row => renderPolicy(row.restday)
    },
    {
      key: 'holiday',
      title: $t('page.hrm.attendance.overtimeRule.holiday'),
      align: 'center',
      minWidth: 200,
      render: row => renderPolicy(row.holiday)
    },
    {
      key: 'minUnitMinutes',
      title: $t('page.hrm.attendance.overtimeRule.minUnitMinutes'),
      align: 'center',
      width: 110,
      render: row =>
        row.minUnitMinutes === 60
          ? $t('page.hrm.attendance.overtimeRule.minUnitHour')
          : $t('page.hrm.attendance.overtimeRule.minUnitHalfHour')
    },
    {
      key: 'monthlyLimitHours',
      title: $t('page.hrm.attendance.overtimeRule.monthlyLimitHours'),
      align: 'center',
      width: 110,
      render: row =>
        row.monthlyLimitHours === 0 ? $t('page.hrm.attendance.overtimeRule.noLimit') : `${row.monthlyLimitHours} h`
    },
    {
      key: 'groupUuids',
      title: $t('page.hrm.attendance.overtimeRule.usedGroups'),
      align: 'center',
      width: 120,
      render: row =>
        row.groupUuids.length > 0
          ? $t('page.hrm.attendance.overtimeRule.boundGroupCount', { n: row.groupUuids.length })
          : $t('page.hrm.attendance.overtimeRule.globalFallback')
    },
    {
      key: 'status',
      title: $t('page.hrm.attendance.overtimeRule.status'),
      align: 'center',
      width: 90,
      render: row => {
        if (row.status === null) {
          return null;
        }
        const tagMap: Record<Api.Hrm.AttendanceStatus, NaiveUI.ThemeColor> = {
          '1': 'success',
          '0': 'warning'
        };
        return <NTag type={tagMap[row.status]}>{$t(hrmAttendanceStatusRecord[row.status])}</NTag>;
      }
    },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      width: 180,
      render: row => (
        <div class="flex-center gap-8px">
          <NButton type="primary" ghost size="small" onClick={() => handleEdit(row)}>
            {$t('common.edit')}
          </NButton>
          <NButton
            type={row.status === '1' ? 'warning' : 'success'}
            ghost
            size="small"
            onClick={() => handleToggle(row)}
          >
            {row.status === '1'
              ? $t('page.hrm.attendance.overtimeRule.disable')
              : $t('page.hrm.attendance.overtimeRule.enable')}
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

/** 当天策略列渲染: 需审批/免审批 · 需打卡/免打卡 · 调休×N */
function renderPolicy(policy: Api.Hrm.DayOvertimePolicy) {
  return (
    <span>
      {policy.approvalRequired === 1
        ? $t('page.hrm.attendance.overtimeRule.needApproval')
        : $t('page.hrm.attendance.overtimeRule.noApproval')}
      {' · '}
      {policy.punchRequired === 1
        ? $t('page.hrm.attendance.overtimeRule.needPunch')
        : $t('page.hrm.attendance.overtimeRule.noPunch')}
      {' · '}
      {$t('page.hrm.attendance.overtimeRule.multiplierValue', { n: policy.compMultiplier })}
    </span>
  );
}

const checkedRowKeys = ref<string[]>([]);
const hasCheckedRows = computed(() => checkedRowKeys.value.length === 0);

// modal state (manual, same pattern as 组织管理 page)
const operateVisible = ref(false);
const operateType = ref<NaiveUI.TableOperateType>('add');
const editingData = ref<Api.Hrm.OvertimeRule | null>(null);

function handleAdd() {
  operateType.value = 'add';
  editingData.value = null;
  operateVisible.value = true;
}

function handleEdit(row: Api.Hrm.OvertimeRule) {
  operateType.value = 'edit';
  editingData.value = { ...row };
  operateVisible.value = true;
}

async function handleToggle(row: Api.Hrm.OvertimeRule) {
  const next: Api.Hrm.AttendanceStatus = row.status === '1' ? '0' : '1';
  const { error } = await ToggleOvertimeRuleStatus({ uuid: row.uuid, status: next });
  if (error) return;
  window.$message?.success($t('common.updateSuccess'));
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
  const { error } = await DeleteOvertimeRule(uuids);

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
  const { error } = await DeleteOvertimeRule(uuids);
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
    <OvertimeRuleSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <NCard
      :title="$t('page.hrm.attendance.overtimeRule.title')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <TableHeaderOperation
          v-model:columns="columnChecks"
          :disabled-delete="!hasCheckedRows"
          :loading="loading"
          add-auth="hrm:overtimerule:add"
          delete-auth="hrm:overtimerule:delete"
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
      <OvertimeRuleOperateModal
        v-model:visible="operateVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="handleSubmitted"
      />
    </NCard>
  </div>
</template>

<style scoped></style>
