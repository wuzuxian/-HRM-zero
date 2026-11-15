<script setup lang="tsx">
import { computed, reactive, ref } from 'vue';
import { NButton, NDrawer, NDrawerContent, NPopconfirm, NTag } from 'naive-ui';
import {
  DeleteAttendanceGroup,
  GetAttendanceGroupList,
  GetAttendanceGroupMembers,
  ToggleAttendanceGroupStatus
} from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { hrmAttendanceStatusRecord } from '@/constants/business';
import { useTable } from '@/hooks/common/table';
import AttendanceGroupOperateModal from './modules/group-operate-modal.vue';
import AttendanceGroupSearch from './modules/group-search.vue';

const appStore = useAppStore();

type LoadingStatus = Record<string, boolean>;
const deleteLoadingStatus = reactive<LoadingStatus>({});
const toggleLoadingStatus = reactive<LoadingStatus>({});

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
  apiFn: GetAttendanceGroupList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    groupName: null,
    groupType: null,
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
      key: 'groupName',
      title: $t('page.hrm.attendance.group.groupName'),
      align: 'center',
      minWidth: 120
    },
    {
      key: 'groupType',
      title: $t('page.hrm.attendance.group.groupType.title'),
      align: 'center',
      width: 100,
      render: row => {
        const tagMap: Record<Api.Hrm.AttendanceGroupType, NaiveUI.ThemeColor> = {
          '1': 'success',
          '2': 'info',
          '3': 'warning'
        };
        const labelMap: Record<Api.Hrm.AttendanceGroupType, string> = {
          '1': $t('page.hrm.attendance.group.groupType.fixed'),
          '2': $t('page.hrm.attendance.group.groupType.scheduling'),
          '3': $t('page.hrm.attendance.group.groupType.flexible')
        };
        return row.groupType ? (
          <NTag type={tagMap[row.groupType]} size="small">
            {labelMap[row.groupType]}
          </NTag>
        ) : null;
      }
    },
    {
      key: 'startTime',
      title: $t('page.hrm.attendance.group.workTime'),
      align: 'center',
      minWidth: 170,
      render: row => {
        if (row.groupType === '3') {
          return $t('page.hrm.attendance.group.flexibleNoTime');
        }
        if (row.groupType === '2') {
          return $t('page.hrm.attendance.group.schedulingFallback');
        }
        if (!row.startTime) {
          return '-';
        }
        return (
          <span>
            {row.startTime} - {row.endTime}
          </span>
        );
      }
    },
    {
      key: 'workDays',
      title: $t('page.hrm.attendance.group.workDays'),
      align: 'center',
      width: 130,
      render: row =>
        row.workDays?.length
          ? row.workDays.map(d => $t(`page.hrm.attendance.group.weekdayShort.${d}` as App.I18n.I18nKey)).join(' ')
          : '-'
    },
    {
      key: 'orgUuids',
      title: $t('page.hrm.attendance.group.scope'),
      align: 'center',
      minWidth: 160,
      render: row => {
        const parts: string[] = [];
        if (row.orgUuids?.length) {
          parts.push($t('page.hrm.attendance.group.scopeDepts', { n: row.orgUuids.length }));
        }
        if (row.memberUuids?.length) {
          parts.push($t('page.hrm.attendance.group.scopeMembers', { n: row.memberUuids.length }));
        }
        if (!parts.length) {
          return '-';
        }
        return (
          <NButton text type="primary" size="small" onClick={() => handleShowMembers(row)}>
            {parts.join(' / ')}
          </NButton>
        );
      }
    },
    {
      key: 'managers',
      title: $t('page.hrm.attendance.group.managers'),
      align: 'center',
      width: 120,
      render: row => (row.managers?.length ? row.managers.join('、') : '-')
    },
    {
      key: 'status',
      title: $t('page.hrm.attendance.group.status'),
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
      width: 210,
      render: row => (
        <div class="flex-center gap-8px">
          <NButton type="primary" ghost size="small" onClick={() => handleEdit(row)}>
            {$t('common.edit')}
          </NButton>
          <NPopconfirm onPositiveClick={() => handleToggle(row)}>
            {{
              default: () =>
                row.status === '1'
                  ? $t('page.hrm.attendance.group.confirmDisable')
                  : $t('page.hrm.attendance.group.confirmEnable'),
              trigger: () => (
                <NButton loading={toggleLoadingStatus[row.uuid]} type="warning" ghost size="small">
                  {row.status === '1'
                    ? $t('page.hrm.attendance.group.disable')
                    : $t('page.hrm.attendance.group.enable')}
                </NButton>
              )
            }}
          </NPopconfirm>
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

// modal state (manual, same pattern as 班次管理 page)
const operateVisible = ref(false);
const operateType = ref<NaiveUI.TableOperateType>('add');
const editingData = ref<Api.Hrm.AttendanceGroup | null>(null);

function handleAdd() {
  operateType.value = 'add';
  editingData.value = null;
  operateVisible.value = true;
}

function handleEdit(row: Api.Hrm.AttendanceGroup) {
  operateType.value = 'edit';
  editingData.value = { ...row };
  operateVisible.value = true;
}

function handleSubmitted() {
  getDataByPage();
}

async function handleToggle(row: Api.Hrm.AttendanceGroup) {
  const nextStatus: Api.Hrm.AttendanceStatus = row.status === '1' ? '0' : '1';
  toggleLoadingStatus[row.uuid] = true;
  const { error } = await ToggleAttendanceGroupStatus({ uuid: row.uuid, status: nextStatus });
  toggleLoadingStatus[row.uuid] = false;
  if (error) return;
  window.$message?.success($t('common.updateSuccess'));
  getData();
}

async function handleBatchDelete() {
  const uuids: string[] = [...checkedRowKeys.value];

  uuids.forEach(uuid => {
    deleteLoadingStatus[uuid] = true;
  });
  const { error } = await DeleteAttendanceGroup(uuids);

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
  const { error } = await DeleteAttendanceGroup(uuids);
  deleteLoadingStatus[uuid] = false;
  if (error) {
    return;
  }
  window.$message?.success($t('common.deleteSuccess'));
  getData();
}

// ---- 生效人数 (生效范围解析: 部门含下级 + 指定人员) ----
const membersVisible = ref(false);
const membersLoading = ref(false);
const members = ref<Api.Hrm.EffectiveGroupMember[]>([]);

async function handleShowMembers(row: Api.Hrm.AttendanceGroup) {
  membersVisible.value = true;
  membersLoading.value = true;
  const { error, data: membersData } = await GetAttendanceGroupMembers(row.uuid);
  membersLoading.value = false;
  if (error) return;
  members.value = membersData?.members ?? [];
}
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <AttendanceGroupSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <NCard
      :title="$t('page.hrm.attendance.group.title')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <TableHeaderOperation
          v-model:columns="columnChecks"
          :disabled-delete="!hasCheckedRows"
          :loading="loading"
          add-auth="hrm:attendancegroup:add"
          delete-auth="hrm:attendancegroup:delete"
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
        :scroll-x="1300"
        :loading="loading"
        remote
        :row-key="row => row.uuid"
        :pagination="mobilePagination"
        class="sm:h-full"
      />
      <AttendanceGroupOperateModal
        v-model:visible="operateVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="handleSubmitted"
      />
      <NDrawer v-model:show="membersVisible" :width="420">
        <NDrawerContent :title="$t('page.hrm.attendance.group.members')" closable>
          <NSpin :show="membersLoading">
            <NList v-if="members.length" bordered>
              <NListItem v-for="member in members" :key="member.uuid">
                <div class="flex items-center justify-between">
                  <div>
                    <div class="font-medium">{{ member.name }}</div>
                    <div class="text-12px text-gray-400">{{ member.employeeNo }}</div>
                  </div>
                  <NTag size="small" type="info">{{ member.orgName || '-' }}</NTag>
                </div>
              </NListItem>
            </NList>
            <NEmpty v-else :description="$t('page.hrm.attendance.group.noMembers')" />
          </NSpin>
        </NDrawerContent>
      </NDrawer>
    </NCard>
  </div>
</template>

<style scoped></style>
