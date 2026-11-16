<script setup lang="tsx">
import { reactive, ref } from 'vue';
import { NButton, NDropdown, NPopconfirm, NTag } from 'naive-ui';
import { AdvancePerfPeriod, DeletePerfPeriods, GetPerfPeriodList, PublishPerfPeriod } from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { useAuth } from '@/hooks/business/auth';
import {
  perfPeriodStatusRecord,
  perfPeriodStatusTag,
  perfPeriodTypeRecord
} from '@/constants/hrm-performance';
import { useTable } from '@/hooks/common/table';
import PeriodOperateModal from './modules/period-operate-modal.vue';
import PeriodSearch from './modules/period-search.vue';

defineOptions({
  name: 'PerfPeriodList'
});

const appStore = useAppStore();
const { hasAuth } = useAuth();

type LoadingStatus = Record<string, boolean>;
const deleteLoadingStatus = reactive<LoadingStatus>({});
const publishLoadingStatus = reactive<LoadingStatus>({});

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
  apiFn: GetPerfPeriodList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    name: null,
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
      key: 'periodNo',
      title: $t('page.hrm.performance.period.periodNo'),
      align: 'center',
      width: 140
    },
    {
      key: 'name',
      title: $t('page.hrm.performance.period.name'),
      align: 'center',
      minWidth: 160
    },
    {
      key: 'periodType',
      title: $t('page.hrm.performance.period.periodType'),
      align: 'center',
      width: 80,
      render: row => $t(perfPeriodTypeRecord[row.periodType])
    },
    {
      key: 'schemeName',
      title: $t('page.hrm.performance.period.schemeName'),
      align: 'center',
      minWidth: 120
    },
    {
      key: 'periodEnd',
      title: $t('page.hrm.performance.period.periodRange'),
      align: 'center',
      width: 180,
      render: row => `${row.periodStart} ~ ${row.periodEnd}`
    },
    {
      key: 'resultPublishDate',
      title: $t('page.hrm.performance.period.resultPublishDate'),
      align: 'center',
      width: 110
    },
    {
      key: 'participantCount',
      title: $t('page.hrm.performance.period.participantCount'),
      align: 'center',
      width: 90
    },
    {
      key: 'status',
      title: $t('page.hrm.performance.period.status'),
      align: 'center',
      width: 100,
      render: row => {
        const tag = perfPeriodStatusTag(row.status);
        if (!row.status || tag === null) {
          return null;
        }
        const label = $t(perfPeriodStatusRecord[row.status]);
        return <NTag type={tag}>{label}</NTag>;
      }
    },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      width: 200,
      render: row => {
        const isDraft = row.status === '0';
        const canPublish = isDraft && hasAuth('hrm:perf_period:publish');
        const moreOptions = [
          { label: $t('page.hrm.performance.period.advance'), key: 'advance' }
        ];
        return (
          <div class="flex-center gap-8px">
            <NButton
              v-auth={['hrm:perf_period:edit']}
              type="primary"
              ghost
              size="small"
              disabled={!isDraft}
              onClick={() => edit(row)}
            >
              {$t('common.edit')}
            </NButton>
            <NPopconfirm onPositiveClick={() => handlePublish(row.uuid)}>
              {{
                default: () => $t('page.hrm.performance.period.publishConfirm'),
                trigger: () => (
                  <NButton
                    loading={publishLoadingStatus[row.uuid]}
                    type="info"
                    ghost
                    size="small"
                    disabled={!canPublish}
                  >
                    {$t('page.hrm.performance.period.publish')}
                  </NButton>
                )
              }}
            </NPopconfirm>
            <NDropdown
              options={moreOptions}
              on-select={(key: string) => handleMore(key, row)}
              v-auth={['hrm:perf_period:edit']}
            >
              <NButton size="small">{$t('page.hrm.performance.period.more')}</NButton>
            </NDropdown>
          </div>
        );
      }
    }
  ]
});

// HRM 页面约定：手动管理弹窗状态（status '0'-'5' 不满足 useTableOperate 的泛型约束）
const operateVisible = ref(false);
const operateType = ref<NaiveUI.TableOperateType>('add');
const editingData = ref<Api.Hrm.Performance.PerfPeriod | null>(null);

function handleAdd() {
  operateType.value = 'add';
  editingData.value = null;
  operateVisible.value = true;
}

function edit(row: Api.Hrm.Performance.PerfPeriod) {
  operateType.value = 'edit';
  editingData.value = row;
  operateVisible.value = true;
}

async function handlePublish(uuid: string) {
  publishLoadingStatus[uuid] = true;
  const { error } = await PublishPerfPeriod(uuid);
  publishLoadingStatus[uuid] = false;
  if (!error) {
    window.$message?.success($t('page.hrm.performance.period.publishSuccess'));
    getData();
  }
}

async function handleMore(key: string, row: Api.Hrm.Performance.PerfPeriod) {
  if (key === 'advance') {
    const { error } = await AdvancePerfPeriod(row.uuid);
    if (!error) {
      window.$message?.success($t('common.updateSuccess'));
      getData();
    }
  }
}

async function handleDelete(uuid: string) {
  deleteLoadingStatus[uuid] = true;
  const { error } = await DeletePerfPeriods([uuid]);
  deleteLoadingStatus[uuid] = false;
  if (!error) {
    getData();
  }
}
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <PeriodSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <NCard
      :title="$t('page.hrm.performance.period.title')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <TableHeaderOperation
          v-model:columns="columnChecks"
          :loading="loading"
          add-auth="hrm:perf_period:add"
          @add="handleAdd"
          @refresh="getData"
        />
      </template>
      <NDataTable
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
      <PeriodOperateModal
        v-model:visible="operateVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="getDataByPage"
      />
    </NCard>
  </div>
</template>

<style scoped></style>
