<script setup lang="tsx">
import { computed, reactive, ref } from 'vue';
import { NButton, NPopconfirm, NTag } from 'naive-ui';
import { DeleteSalaryItem, GetSalaryItemList } from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import {
  hrmSalaryItemClassRecord,
  hrmSalaryItemTypeRecord,
  hrmSalaryItemValueModeRecord,
  hrmSalaryStatusRecord
} from '@/constants/business';
import { useTable } from '@/hooks/common/table';
import SalaryItemOperateModal from './modules/salary-item-operate-modal.vue';
import SalaryItemSearch from './modules/salary-item-search.vue';

const appStore = useAppStore();

type LoadingStatus = Record<string, boolean>;
const deleteLoadingStatus = reactive<LoadingStatus>({});

function enumTag<T extends string>(value: T, record: Record<T, App.I18n.I18nKey>) {
  return <NTag size="small">{$t(record[value])}</NTag>;
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
  apiFn: GetSalaryItemList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    itemName: null,
    itemClass: null,
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
      key: 'itemName',
      title: $t('page.hrm.salary.item.itemName'),
      align: 'center',
      minWidth: 110
    },
    {
      key: 'itemCode',
      title: $t('page.hrm.salary.item.itemCode'),
      align: 'center',
      minWidth: 150
    },
    {
      key: 'itemType',
      title: $t('page.hrm.salary.item.itemType.title'),
      align: 'center',
      width: 90,
      render: row => enumTag(row.itemType, hrmSalaryItemTypeRecord)
    },
    {
      key: 'itemClass',
      title: $t('page.hrm.salary.item.itemClass.title'),
      align: 'center',
      width: 80,
      render: row => enumTag(row.itemClass, hrmSalaryItemClassRecord)
    },
    {
      key: 'valueMode',
      title: $t('page.hrm.salary.item.valueMode.title'),
      align: 'center',
      width: 100,
      render: row => enumTag(row.valueMode, hrmSalaryItemValueModeRecord)
    },
    {
      key: 'calcOrder',
      title: $t('page.hrm.salary.item.calcOrder'),
      align: 'center',
      width: 90
    },
    {
      key: 'orderNum',
      title: $t('page.hrm.salary.item.orderNum'),
      align: 'center',
      width: 90
    },
    {
      key: 'isPreset',
      title: $t('page.hrm.salary.item.isPreset'),
      align: 'center',
      width: 90,
      render: row =>
        row.isPreset === 1 ? (
          <NTag type="info" size="small">
            {$t('page.hrm.salary.item.presetYes')}
          </NTag>
        ) : (
          <span>{$t('page.hrm.salary.item.presetNo')}</span>
        )
    },
    {
      key: 'status',
      title: $t('page.hrm.salary.item.status'),
      align: 'center',
      width: 90,
      render: row => {
        if (row.status === null) {
          return null;
        }
        const tagMap: Record<Api.Hrm.SalarySubjectStatus, NaiveUI.ThemeColor> = {
          '1': 'success',
          '0': 'warning'
        };
        return <NTag type={tagMap[row.status]}>{$t(hrmSalaryStatusRecord[row.status])}</NTag>;
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

// modal state (manual, same pattern as 假期类型 page)
const operateVisible = ref(false);
const operateType = ref<NaiveUI.TableOperateType>('add');
const editingData = ref<Api.Hrm.SalaryItem | null>(null);

function handleAdd() {
  operateType.value = 'add';
  editingData.value = null;
  operateVisible.value = true;
}

function handleEdit(row: Api.Hrm.SalaryItem) {
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
  const { error } = await DeleteSalaryItem(uuids);

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
  const { error } = await DeleteSalaryItem(uuids);
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
    <SalaryItemSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <NCard
      :title="$t('page.hrm.salary.item.title')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <TableHeaderOperation
          v-model:columns="columnChecks"
          :disabled-delete="!hasCheckedRows"
          :loading="loading"
          add-auth="hrm:salary_item:add"
          delete-auth="hrm:salary_item:delete"
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
      <SalaryItemOperateModal
        v-model:visible="operateVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="handleSubmitted"
      />
    </NCard>
  </div>
</template>

<style scoped></style>
