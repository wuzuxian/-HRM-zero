<script setup lang="tsx">
import { computed, reactive, ref } from 'vue';
import { NButton, NPopconfirm, NTag } from 'naive-ui';
import { DeleteScheduleCycle, GetScheduleCycleList, ToggleScheduleCycleStatus } from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { hrmAttendanceStatusRecord } from '@/constants/business';
import { useTable } from '@/hooks/common/table';
import ScheduleCycleOperateModal from './modules/schedule-cycle-operate-modal.vue';
import ScheduleCycleSearch from './modules/schedule-cycle-search.vue';

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
  apiFn: GetScheduleCycleList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    cycleName: null,
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
      key: 'cycleName',
      title: $t('page.hrm.attendance.scheduleCycle.cycleName'),
      align: 'center',
      minWidth: 140
    },
    {
      key: 'shiftUuids',
      title: $t('page.hrm.attendance.scheduleCycle.shiftSequence'),
      align: 'center',
      minWidth: 200,
      render: row =>
        row.shiftUuids.length > 0
          ? $t('page.hrm.attendance.scheduleCycle.sequenceSummary', { n: row.shiftUuids.length })
          : '-'
    },
    {
      key: 'uuid',
      title: $t('page.hrm.attendance.scheduleCycle.cycleDays'),
      align: 'center',
      width: 100,
      render: row => (row.shiftUuids.length > 0 ? `${row.shiftUuids.length}` : '-')
    },
    {
      key: 'groupUuids',
      title: $t('page.hrm.attendance.scheduleCycle.usedGroups'),
      align: 'center',
      width: 120,
      render: row =>
        row.groupUuids.length > 0
          ? $t('page.hrm.attendance.scheduleCycle.boundGroupCount', { n: row.groupUuids.length })
          : $t('page.hrm.attendance.scheduleCycle.noBoundGroup')
    },
    {
      key: 'remark',
      title: $t('page.hrm.attendance.scheduleCycle.remark'),
      align: 'center',
      minWidth: 120,
      render: row => row.remark || '-'
    },
    {
      key: 'status',
      title: $t('page.hrm.attendance.scheduleCycle.status'),
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
              ? $t('page.hrm.attendance.scheduleCycle.disable')
              : $t('page.hrm.attendance.scheduleCycle.enable')}
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

// modal state (manual, same pattern as 组织管理 page)
const operateVisible = ref(false);
const operateType = ref<NaiveUI.TableOperateType>('add');
const editingData = ref<Api.Hrm.ScheduleCycle | null>(null);

function handleAdd() {
  operateType.value = 'add';
  editingData.value = null;
  operateVisible.value = true;
}

function handleEdit(row: Api.Hrm.ScheduleCycle) {
  operateType.value = 'edit';
  editingData.value = { ...row };
  operateVisible.value = true;
}

async function handleToggle(row: Api.Hrm.ScheduleCycle) {
  const next: Api.Hrm.AttendanceStatus = row.status === '1' ? '0' : '1';
  const { error } = await ToggleScheduleCycleStatus({ uuid: row.uuid, status: next });
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
  const { error } = await DeleteScheduleCycle(uuids);

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
  const { error } = await DeleteScheduleCycle(uuids);
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
    <ScheduleCycleSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <NCard
      :title="$t('page.hrm.attendance.scheduleCycle.title')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <TableHeaderOperation
          v-model:columns="columnChecks"
          :disabled-delete="!hasCheckedRows"
          :loading="loading"
          add-auth="hrm:schedulecycle:add"
          delete-auth="hrm:schedulecycle:delete"
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
      <ScheduleCycleOperateModal
        v-model:visible="operateVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="handleSubmitted"
      />
    </NCard>
  </div>
</template>

<style scoped></style>
