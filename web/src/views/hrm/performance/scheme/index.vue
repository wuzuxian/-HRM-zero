<script setup lang="tsx">
import { reactive } from 'vue';
import { NButton, NPopconfirm, NTag } from 'naive-ui';
import { DeletePerfSchemes, GetPerfSchemeList } from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import {
  perfSchemeAssessmentModeRecord,
  perfSchemePayLinkModeRecord,
  perfSchemePeriodTypeRecord,
  perfSchemeScoreModeRecord,
  perfStatusRecord,
  perfStatusTag
} from '@/constants/hrm-performance';
import { useTable, useTableOperate } from '@/hooks/common/table';
import SchemeOperateModal from './modules/scheme-operate-modal.vue';
import SchemeSearch from './modules/scheme-search.vue';

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
  apiFn: GetPerfSchemeList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    name: null,
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
      key: 'name',
      title: $t('page.hrm.performance.scheme.name'),
      align: 'center',
      minWidth: 160
    },
    {
      key: 'templateName',
      title: $t('page.hrm.performance.scheme.templateName'),
      align: 'center',
      minWidth: 140
    },
    {
      key: 'periodType',
      title: $t('page.hrm.performance.scheme.periodType'),
      align: 'center',
      width: 90,
      render: row => $t(perfSchemePeriodTypeRecord[row.periodType])
    },
    {
      key: 'assessmentMode',
      title: $t('page.hrm.performance.scheme.assessmentMode'),
      align: 'center',
      width: 130,
      render: row => $t(perfSchemeAssessmentModeRecord[row.assessmentMode])
    },
    {
      key: 'scoreMode',
      title: $t('page.hrm.performance.scheme.scoreMode'),
      align: 'center',
      width: 90,
      render: row => $t(perfSchemeScoreModeRecord[row.scoreMode])
    },
    {
      key: 'payLinkMode',
      title: $t('page.hrm.performance.scheme.payLinkMode'),
      align: 'center',
      width: 90,
      render: row => $t(perfSchemePayLinkModeRecord[row.payLinkMode])
    },
    {
      key: 'status',
      title: $t('page.hrm.performance.scheme.status'),
      align: 'center',
      width: 90,
      render: row => {
        const tag = perfStatusTag(row.status);
        if (!row.status || tag === null) {
          return null;
        }
        const label = $t(perfStatusRecord[row.status]);
        return <NTag type={tag}>{label}</NTag>;
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

const { drawerVisible, operateType, editingData, handleAdd, handleEdit, checkedRowKeys, onBatchDeleted, onDeleted } =
  useTableOperate(data, getData);

async function handleBatchDelete() {
  const uuids: string[] = checkedRowKeys.value.map(uuid => uuid);

  uuids.forEach(uuid => {
    deleteLoadingStatus[uuid] = true;
  });
  const { error } = await DeletePerfSchemes(uuids);

  uuids.forEach(uuid => {
    deleteLoadingStatus[uuid] = false;
  });

  if (error) return;

  onBatchDeleted();
}

async function handleDelete(uuid: string) {
  const uuids: string[] = [uuid];
  deleteLoadingStatus[uuid] = true;
  const { error } = await DeletePerfSchemes(uuids);
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
    <SchemeSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <NCard
      :title="$t('page.hrm.performance.scheme.title')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <TableHeaderOperation
          v-model:columns="columnChecks"
          :disabled-delete="checkedRowKeys.length === 0"
          :loading="loading"
          add-auth="hrm:perf_scheme:add"
          delete-auth="hrm:perf_scheme:delete"
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
      <SchemeOperateModal
        v-model:visible="drawerVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="getDataByPage"
      />
    </NCard>
  </div>
</template>

<style scoped></style>
