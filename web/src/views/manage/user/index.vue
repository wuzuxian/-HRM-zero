<script setup lang="tsx">
import { reactive, ref, watch } from 'vue';
import { NButton, NPopconfirm, NTag } from 'naive-ui';
import { DeleteUser, GetLinkedEmployees, GetUserList } from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { enableStatusRecord, userGenderRecord } from '@/constants/business';
import { useTable, useTableOperate } from '@/hooks/common/table';
import UserOperateDrawer from './modules/user-operate-modal.vue';
import UserSearch from './modules/user-search.vue';

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
  apiFn: GetUserList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    // if you want to use the searchParams in Form, you need to define the following properties, and the value is null
    // the value can not be undefined, otherwise the property in Form will not be reactive
    status: null,
    username: null,
    userGender: null,
    nickName: null,
    userPhone: null,
    userEmail: null
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
      key: 'username',
      title: $t('page.manage.user.username'),
      align: 'center',
      minWidth: 100
    },
    {
      key: 'userGender',
      title: $t('page.manage.user.userGender'),
      align: 'center',
      width: 100,
      render: row => {
        // e2e 造数可能写入空字符串 gender，未知的枚举值直接不渲染 tag，避免 $t(undefined) 崩掉整个表格 patch
        const genderKey =
          row.userGender !== null ? userGenderRecord[row.userGender as Api.Manage.UserGender] : undefined;

        if (!genderKey) {
          return null;
        }

        const tagMap: Record<Api.Manage.UserGender, NaiveUI.ThemeColor> = {
          1: 'primary',
          2: 'error'
        };

        return <NTag type={tagMap[row.userGender as Api.Manage.UserGender]}>{$t(genderKey)}</NTag>;
      }
    },
    {
      key: 'nickName',
      title: $t('page.manage.user.nickName'),
      align: 'center',
      minWidth: 100
    },
    {
      key: 'userPhone',
      title: $t('page.manage.user.userPhone'),
      align: 'center',
      width: 120
    },
    {
      key: 'userEmail',
      title: $t('page.manage.user.userEmail'),
      align: 'center',
      minWidth: 200
    },
    {
      // 只读「关联员工」列 (HRM-51): 工号 + 姓名; 未关联账号显示 '-'
      // (关联唯一入口在员工侧, 用户管理侧不做写入, ADR 0007)
      key: 'linkedEmployee',
      title: $t('page.manage.user.linkedEmployee'),
      align: 'center',
      width: 120,
      render: row => linkedEmployeeText.value.get(row.uuid) || '-'
    },
    {
      key: 'status',
      title: $t('page.manage.user.userStatus'),
      align: 'center',
      width: 100,
      render: row => {
        const statusKey = row.status !== null ? enableStatusRecord[row.status as Api.Common.EnableStatus] : undefined;

        if (!statusKey) {
          return null;
        }

        const tagMap: Record<Api.Common.EnableStatus, NaiveUI.ThemeColor> = {
          1: 'success',
          2: 'warning'
        };

        return <NTag type={tagMap[row.status as Api.Common.EnableStatus]}>{$t(statusKey)}</NTag>;
      }
    },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      width: 130,
      render: row => (
        <div class="flex-center gap-8px">
          <NButton type="primary" ghost size="small" onClick={() => edit(row.uuid)}>
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

const {
  drawerVisible,
  operateType,
  editingData,
  handleAdd,
  handleEdit,
  checkedRowKeys,
  onBatchDeleted,
  onDeleted
  // closeDrawer
} = useTableOperate(data, getData);

// ---- 关联员工列 (HRM-51): 按当前页账号 uuid 批量反查工号+姓名 ----
const linkedEmployeeText = ref<Map<string, string>>(new Map());

async function loadLinkedEmployees() {
  const uuids = (data.value ?? []).map(row => row.uuid).filter(Boolean);
  if (uuids.length === 0) {
    linkedEmployeeText.value = new Map();
    return;
  }
  const { error, data: resp } = await GetLinkedEmployees(uuids);
  if (error || !resp) {
    return;
  }
  const next = new Map<string, string>();
  for (const item of resp.records ?? []) {
    next.set(item.userUuid, `${item.employeeNo} ${item.employeeName}`);
  }
  linkedEmployeeText.value = next;
}

watch(data, () => {
  loadLinkedEmployees();
});

async function handleBatchDelete() {
  // request

  const uuids: string[] = checkedRowKeys.value.map(uuid => uuid);

  uuids.forEach(uuid => {
    deleteLoadingStatus[uuid] = true;
  });
  const { error } = await DeleteUser(uuids);

  uuids.forEach(uuid => {
    deleteLoadingStatus[uuid] = false;
  });

  if (error) return;

  onBatchDeleted();
}

async function handleDelete(uuid: string) {
  // request
  const uuids: string[] = [uuid];
  deleteLoadingStatus[uuid] = true;
  const { error } = await DeleteUser(uuids);
  deleteLoadingStatus[uuid] = false;
  if (error) {
    return;
  }
  onDeleted();
}

function edit(uuid: string) {
  handleEdit(uuid);
}
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <UserSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <NCard :title="$t('page.manage.user.title')" :bordered="false" size="small" class="sm:flex-1-hidden card-wrapper">
      <template #header-extra>
        <TableHeaderOperation
          v-model:columns="columnChecks"
          :disabled-delete="checkedRowKeys.length === 0"
          :loading="loading"
          add-auth="manage:user:add"
          delete-auth="manage:user:delete"
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
        :scroll-x="962"
        :loading="loading"
        remote
        :row-key="row => row.uuid"
        :pagination="mobilePagination"
        class="sm:h-full"
      />
      <UserOperateDrawer
        v-model:visible="drawerVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="getDataByPage"
      />
    </NCard>
  </div>
</template>

<style scoped></style>
