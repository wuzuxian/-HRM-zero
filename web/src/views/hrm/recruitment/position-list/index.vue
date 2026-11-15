<script setup lang="tsx">
import { computed, reactive, ref } from 'vue';
import { NButton, NPopconfirm, NTag } from 'naive-ui';
import { DeletePosition, GetPositionList, UpdatePositionStatus } from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import {
  hrmPositionPostCategoryRecord,
  hrmPositionStatusRecord
} from '@/constants/business';
import { useTable } from '@/hooks/common/table';
import PositionOperateModal from './modules/position-operate-modal.vue';
import PositionSearch from './modules/position-search.vue';

const appStore = useAppStore();

type LoadingStatus = Record<string, boolean>;
const deleteLoadingStatus = reactive<LoadingStatus>({});
const actionLoadingStatus = reactive<LoadingStatus>({});

function statusTag(row: Api.Hrm.HrmPosition) {
  if (row.status === null) return null;
  const tagMap: Record<Api.Hrm.PositionStatus, NaiveUI.ThemeColor> = {
    open: 'info',
    full: 'warning',
    closed: 'default'
  };
  return <NTag type={tagMap[row.status]}>{$t(hrmPositionStatusRecord[row.status])}</NTag>;
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
  apiFn: GetPositionList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    postName: null,
    requirementUuid: null,
    postCategory: null,
    status: null
  },
  columns: () => [
    {
      key: 'index',
      title: $t('common.index'),
      align: 'center',
      width: 64
    },
    {
      key: 'postName',
      title: $t('page.hrm.position.postName'),
      align: 'center',
      width: 140
    },
    {
      key: 'requirementBillNo',
      title: $t('page.hrm.position.requirementBillNo'),
      align: 'center',
      width: 170,
      render: row => row.requirementBillNo || '-'
    },
    {
      key: 'postCategory',
      title: $t('page.hrm.position.postCategory.title'),
      align: 'center',
      width: 100,
      render: row =>
        row.postCategory === null ? null : (
          <NTag size="small">{$t(hrmPositionPostCategoryRecord[row.postCategory])}</NTag>
        )
    },
    {
      key: 'ownerName',
      title: $t('page.hrm.position.ownerName'),
      align: 'center',
      width: 110,
      render: row => row.ownerName || '-'
    },
    {
      key: 'status',
      title: $t('page.hrm.position.status.title'),
      align: 'center',
      width: 100,
      render: row => statusTag(row)
    },
    {
      key: 'headcount',
      title: $t('page.hrm.position.headcount'),
      align: 'center',
      width: 90
    },
    {
      key: 'hiredCount',
      title: $t('page.hrm.position.hiredCount'),
      align: 'center',
      width: 100,
      render: row => `${row.hiredCount}/${row.headcount}`
    },
    {
      key: 'publishDate',
      title: $t('page.hrm.position.publishDate'),
      align: 'center',
      width: 110,
      render: row => row.publishDate || '-'
    },
    {
      key: 'createTime',
      title: $t('page.hrm.position.createTime'),
      align: 'center',
      width: 160
    },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      width: 240,
      fixed: 'right',
      render: row => renderOperate(row)
    }
  ]
});

function renderOperate(row: Api.Hrm.HrmPosition) {
  const actions = new Set<string>();
  if (row.status !== 'closed') {
    actions.add('edit');
    actions.add('close');
  }
  actions.add('delete');

  return (
    <div class="flex-center flex-wrap gap-4px">
      {actions.has('edit') && (
        <NButton type="primary" ghost size="small" onClick={() => handleEdit(row)}>
          {$t('common.edit')}
        </NButton>
      )}
      {actions.has('close') && (
        <NPopconfirm onPositiveClick={() => handleClose(row.uuid)}>
          {{
            default: () => $t('page.hrm.position.confirmClose'),
            trigger: () => (
              <NButton loading={actionLoadingStatus[row.uuid]} type="warning" ghost size="small">
                {$t('page.hrm.position.close')}
              </NButton>
            )
          }}
        </NPopconfirm>
      )}
      {actions.has('delete') && (
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
      )}
    </div>
  );
}

const checkedRowKeys = ref<string[]>([]);
const hasCheckedRows = computed(() => checkedRowKeys.value.length === 0);

// modal state (manual, same pattern as 组织管理 page)
const operateVisible = ref(false);
const operateType = ref<NaiveUI.TableOperateType>('add');
const editingData = ref<Api.Hrm.HrmPosition | null>(null);

function handleAdd() {
  operateType.value = 'add';
  editingData.value = null;
  operateVisible.value = true;
}

function handleEdit(row: Api.Hrm.HrmPosition) {
  operateType.value = 'edit';
  editingData.value = { ...row };
  operateVisible.value = true;
}

function handleSubmitted() {
  getDataByPage();
}

async function handleClose(uuid: string) {
  actionLoadingStatus[uuid] = true;
  const { error } = await UpdatePositionStatus({ uuid, status: 'closed' });
  actionLoadingStatus[uuid] = false;
  if (error) return;
  window.$message?.success($t('common.updateSuccess'));
  getData();
}

async function handleBatchDelete() {
  const uuids: string[] = [...checkedRowKeys.value];

  uuids.forEach(uuid => {
    deleteLoadingStatus[uuid] = true;
  });
  const { error } = await DeletePosition(uuids);

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
  const { error } = await DeletePosition(uuids);
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
    <PositionSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <NCard
      :title="$t('page.hrm.position.title')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <TableHeaderOperation
          v-model:columns="columnChecks"
          :disabled-delete="!hasCheckedRows"
          :loading="loading"
          add-auth="hrm:position:add"
          delete-auth="hrm:position:delete"
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
        :scroll-x="1500"
        :loading="loading"
        remote
        :row-key="row => row.uuid"
        :pagination="mobilePagination"
        class="sm:h-full"
      />
      <PositionOperateModal
        v-model:visible="operateVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="handleSubmitted"
      />
    </NCard>
  </div>
</template>

<style scoped></style>
