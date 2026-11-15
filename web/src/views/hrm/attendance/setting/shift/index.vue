<script setup lang="tsx">
import { computed, reactive, ref } from 'vue';
import { NButton, NPopconfirm, NTag } from 'naive-ui';
import { DeleteShift, GetShiftList } from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { hrmAttendanceStatusRecord } from '@/constants/business';
import { useTable } from '@/hooks/common/table';
import ShiftOperateModal from './modules/shift-operate-modal.vue';
import ShiftSearch from './modules/shift-search.vue';

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
  apiFn: GetShiftList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    shiftName: null,
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
      key: 'shiftName',
      title: $t('page.hrm.attendance.shift.shiftName'),
      align: 'center',
      minWidth: 120
    },
    {
      key: 'startTime',
      title: $t('page.hrm.attendance.shift.workTime'),
      align: 'center',
      minWidth: 160,
      render: row => (
        <div class="flex-center gap-8px">
          <span>
            {row.startTime} - {row.crossDay === 1 ? $t('page.hrm.attendance.shift.nextDay') : ''}
            {row.endTime}
          </span>
          {row.crossDay === 1 ? (
            <NTag type="warning" size="small">
              {$t('page.hrm.attendance.shift.crossDay')}
            </NTag>
          ) : null}
        </div>
      )
    },
    {
      key: 'restStart',
      title: $t('page.hrm.attendance.shift.restTime'),
      align: 'center',
      width: 130,
      render: row =>
        row.restStart && row.restEnd ? (
          <span>
            {row.restStart} - {row.restEnd}
          </span>
        ) : (
          $t('page.hrm.attendance.shift.noRest')
        )
    },
    {
      key: 'standardHours',
      title: $t('page.hrm.attendance.shift.standardHours'),
      align: 'center',
      width: 100,
      render: row => `${row.standardHours} h`
    },
    {
      key: 'flexibleMinutes',
      title: $t('page.hrm.attendance.shift.flexibleMinutes'),
      align: 'center',
      width: 110,
      render: row =>
        row.flexibleMinutes === 0
          ? $t('page.hrm.attendance.shift.notFlexible')
          : $t('page.hrm.attendance.shift.flexibleMinutesValue', { n: row.flexibleMinutes })
    },
    {
      key: 'status',
      title: $t('page.hrm.attendance.shift.status'),
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
      key: 'remark',
      title: $t('page.hrm.attendance.shift.remark'),
      align: 'center',
      minWidth: 120,
      render: row => row.remark || '-'
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
const editingData = ref<Api.Hrm.Shift | null>(null);

function handleAdd() {
  operateType.value = 'add';
  editingData.value = null;
  operateVisible.value = true;
}

function handleEdit(row: Api.Hrm.Shift) {
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
  const { error } = await DeleteShift(uuids);

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
  const { error } = await DeleteShift(uuids);
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
    <ShiftSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <NCard
      :title="$t('page.hrm.attendance.shift.title')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <TableHeaderOperation
          v-model:columns="columnChecks"
          :disabled-delete="!hasCheckedRows"
          :loading="loading"
          add-auth="hrm:shift:add"
          delete-auth="hrm:shift:delete"
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
      <ShiftOperateModal
        v-model:visible="operateVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="handleSubmitted"
      />
    </NCard>
  </div>
</template>

<style scoped></style>
