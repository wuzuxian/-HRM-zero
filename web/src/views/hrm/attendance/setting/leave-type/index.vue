<script setup lang="tsx">
import { computed, reactive, ref } from 'vue';
import { NButton, NPopconfirm, NTag } from 'naive-ui';
import { DeleteLeaveType, GetLeaveTypeList } from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { hrmAttendanceStatusRecord } from '@/constants/business';
import { useTable } from '@/hooks/common/table';
import LeaveTypeOperateModal from './modules/leave-type-operate-modal.vue';
import LeaveTypeSearch from './modules/leave-type-search.vue';

const appStore = useAppStore();

type LoadingStatus = Record<string, boolean>;
const deleteLoadingStatus = reactive<LoadingStatus>({});

function yesNoTag(value: number) {
  if (value === 1) {
    return (
      <NTag type="success" size="small">
        {$t('page.hrm.attendance.yes')}
      </NTag>
    );
  }
  return <span>{$t('page.hrm.attendance.no')}</span>;
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
  apiFn: GetLeaveTypeList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    typeName: null,
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
      key: 'typeName',
      title: $t('page.hrm.attendance.leaveType.typeName'),
      align: 'center',
      minWidth: 110
    },
    {
      key: 'typeCode',
      title: $t('page.hrm.attendance.leaveType.typeCode'),
      align: 'center',
      minWidth: 150
    },
    {
      key: 'typeValue',
      title: $t('page.hrm.attendance.leaveType.typeValue'),
      align: 'center',
      width: 80
    },
    {
      key: 'balanceEnabled',
      title: $t('page.hrm.attendance.leaveType.balanceEnabled'),
      align: 'center',
      width: 100,
      render: row => yesNoTag(row.balanceEnabled)
    },
    {
      key: 'attachmentRequired',
      title: $t('page.hrm.attendance.leaveType.attachmentRequired'),
      align: 'center',
      width: 100,
      render: row => yesNoTag(row.attachmentRequired)
    },
    {
      key: 'defaultDays',
      title: $t('page.hrm.attendance.leaveType.defaultDays'),
      align: 'center',
      width: 100,
      render: row => `${row.defaultDays}`
    },
    {
      key: 'orderNum',
      title: $t('page.hrm.attendance.leaveType.orderNum'),
      align: 'center',
      width: 80
    },
    {
      key: 'status',
      title: $t('page.hrm.attendance.leaveType.status'),
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

// modal state (manual, same pattern as 组织管理 page)
const operateVisible = ref(false);
const operateType = ref<NaiveUI.TableOperateType>('add');
const editingData = ref<Api.Hrm.LeaveType | null>(null);

function handleAdd() {
  operateType.value = 'add';
  editingData.value = null;
  operateVisible.value = true;
}

function handleEdit(row: Api.Hrm.LeaveType) {
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
  const { error } = await DeleteLeaveType(uuids);

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
  const { error } = await DeleteLeaveType(uuids);
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
    <LeaveTypeSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <NCard
      :title="$t('page.hrm.attendance.leaveType.title')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <TableHeaderOperation
          v-model:columns="columnChecks"
          :disabled-delete="!hasCheckedRows"
          :loading="loading"
          add-auth="hrm:leavetype:add"
          delete-auth="hrm:leavetype:delete"
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
        :scroll-x="1100"
        :loading="loading"
        remote
        :row-key="row => row.uuid"
        :pagination="mobilePagination"
        class="sm:h-full"
      />
      <LeaveTypeOperateModal
        v-model:visible="operateVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="handleSubmitted"
      />
    </NCard>
  </div>
</template>

<style scoped></style>
