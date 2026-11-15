<script setup lang="tsx">
import { computed, reactive, ref } from 'vue';
import { NButton, NPopconfirm, NTag } from 'naive-ui';
import { DeleteCandidate, GetCandidateList, UpdateCandidateStatus } from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { useAuth } from '@/hooks/business/auth';
import {
  hrmCandidateEducationRecord,
  hrmCandidateSourceChannelRecord,
  hrmCandidateStatusRecord
} from '@/constants/business';
import { useTable } from '@/hooks/common/table';
import CandidateImportModal from './modules/candidate-import-modal.vue';
import CandidateOperateModal from './modules/candidate-operate-modal.vue';
import CandidateSearch from './modules/candidate-search.vue';

const appStore = useAppStore();
const { hasAuth } = useAuth();

type LoadingStatus = Record<string, boolean>;
const deleteLoadingStatus = reactive<LoadingStatus>({});
const statusLoadingStatus = reactive<LoadingStatus>({});

/** 候选人状态流转顺序 (新线索→已投递→面试中→已发Offer→录用通过; 不可跳级/不可回退) */
const STATUS_ORDER: Api.Hrm.CandidateStatus[] = ['new', 'delivered', 'interviewing', 'offer', 'hired'];

/** 候选人状态流转 tag 色 */
const statusTagMap: Record<Api.Hrm.CandidateStatus, NaiveUI.ThemeColor> = {
  new: 'default',
  delivered: 'info',
  interviewing: 'warning',
  offer: 'success',
  hired: 'success',
  rejected: 'error'
};

/** 下一个流转阶段 (hired/rejected 无) */
function nextStatus(row: Api.Hrm.CandidateStatus | null): Api.Hrm.CandidateStatus | null {
  if (row === null) return null;
  const idx = STATUS_ORDER.indexOf(row);
  if (idx < 0 || idx === STATUS_ORDER.length - 1) return null;
  return STATUS_ORDER[idx + 1];
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
  apiFn: GetCandidateList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    name: null,
    phone: null,
    sourceChannel: null,
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
      key: 'name',
      title: $t('page.hrm.candidate.name'),
      align: 'center',
      width: 100
    },
    {
      key: 'phone',
      title: $t('page.hrm.candidate.phone'),
      align: 'center',
      width: 130,
      render: row => row.phone || '-'
    },
    {
      key: 'email',
      title: $t('page.hrm.candidate.email'),
      align: 'center',
      width: 180,
      render: row => row.email || '-'
    },
    {
      key: 'sourceChannel',
      title: $t('page.hrm.candidate.sourceChannel.title'),
      align: 'center',
      width: 100,
      render: row =>
        row.sourceChannel === null ? null : (
          <NTag size="small">{$t(hrmCandidateSourceChannelRecord[row.sourceChannel])}</NTag>
        )
    },
    {
      key: 'education',
      title: $t('page.hrm.candidate.education.title'),
      align: 'center',
      width: 90,
      render: row =>
        row.education === null ? null : <NTag size="small">{$t(hrmCandidateEducationRecord[row.education])}</NTag>
    },
    {
      key: 'currentCompany',
      title: $t('page.hrm.candidate.currentCompany'),
      align: 'center',
      width: 130,
      render: row => row.currentCompany || '-'
    },
    {
      key: 'currentPosition',
      title: $t('page.hrm.candidate.currentPosition'),
      align: 'center',
      width: 130,
      render: row => row.currentPosition || '-'
    },
    {
      key: 'expectSalary',
      title: $t('page.hrm.candidate.expectSalary'),
      align: 'center',
      width: 110,
      render: row => row.expectSalary || '-'
    },
    {
      key: 'status',
      title: $t('page.hrm.candidate.status.title'),
      align: 'center',
      width: 100,
      render: row =>
        row.status === null ? null : (
          <NTag type={statusTagMap[row.status]} size="small">
            {$t(hrmCandidateStatusRecord[row.status])}
          </NTag>
        )
    },
    {
      key: 'createTime',
      title: $t('page.hrm.candidate.createTime'),
      align: 'center',
      width: 160
    },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      width: 260,
      fixed: 'right',
      render: row => renderOperate(row)
    }
  ]
});

function renderOperate(row: Api.Hrm.HrmCandidate) {
  const next = nextStatus(row.status);
  return (
    <div class="flex-center flex-wrap gap-4px">
      <NButton type="primary" ghost size="small" onClick={() => handleEdit(row)}>
        {$t('common.edit')}
      </NButton>
      {next && (
        <NPopconfirm onPositiveClick={() => handleAdvanceStatus(row, next)}>
          {{
            default: () => $t('page.hrm.candidate.confirmAdvance'),
            trigger: () => (
              <NButton loading={statusLoadingStatus[row.uuid]} type="info" ghost size="small">
                {$t('page.hrm.candidate.advanceTo', { status: $t(hrmCandidateStatusRecord[next]) })}
              </NButton>
            )
          }}
        </NPopconfirm>
      )}
      {row.status !== 'hired' && row.status !== 'rejected' && (
        <NPopconfirm onPositiveClick={() => handleReject(row)}>
          {{
            default: () => $t('page.hrm.candidate.confirmReject'),
            trigger: () => (
              <NButton loading={statusLoadingStatus[row.uuid]} type="error" ghost size="small">
                {$t('page.hrm.candidate.reject')}
              </NButton>
            )
          }}
        </NPopconfirm>
      )}
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
  );
}

const checkedRowKeys = ref<string[]>([]);
const hasCheckedRows = computed(() => checkedRowKeys.value.length === 0);

// modal state (manual, same pattern as 招聘需求 page)
const operateVisible = ref(false);
const operateType = ref<NaiveUI.TableOperateType>('add');
const editingData = ref<Api.Hrm.HrmCandidate | null>(null);

function handleAdd() {
  operateType.value = 'add';
  editingData.value = null;
  operateVisible.value = true;
}

function handleEdit(row: Api.Hrm.HrmCandidate) {
  operateType.value = 'edit';
  editingData.value = { ...row };
  operateVisible.value = true;
}

function handleSubmitted() {
  getDataByPage();
}

// ---- import modal (Excel 导入; AI 入库通道按 ADR-0002 裁剪) ----
const importVisible = ref(false);

async function handleAdvanceStatus(row: Api.Hrm.HrmCandidate, target: Api.Hrm.CandidateStatus) {
  statusLoadingStatus[row.uuid] = true;
  const { error } = await UpdateCandidateStatus({ uuid: row.uuid, status: target });
  statusLoadingStatus[row.uuid] = false;
  if (error) return;
  window.$message?.success($t('common.updateSuccess'));
  getData();
}

async function handleReject(row: Api.Hrm.HrmCandidate) {
  statusLoadingStatus[row.uuid] = true;
  const { error } = await UpdateCandidateStatus({ uuid: row.uuid, status: 'rejected' });
  statusLoadingStatus[row.uuid] = false;
  if (error) return;
  window.$message?.success($t('common.updateSuccess'));
  getData();
}

async function handleBatchDelete() {
  const uuids: string[] = [...checkedRowKeys.value];

  uuids.forEach(uuid => {
    deleteLoadingStatus[uuid] = true;
  });
  const { error } = await DeleteCandidate(uuids);

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
  const { error } = await DeleteCandidate(uuids);
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
    <CandidateSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <NCard :title="$t('page.hrm.candidate.title')" :bordered="false" size="small" class="sm:flex-1-hidden card-wrapper">
      <template #header-extra>
        <NSpace :size="12" align="center">
          <NButton v-if="hasAuth('hrm:candidate:import')" type="primary" ghost @click="importVisible = true">
            <template #icon>
              <icon-mdi-file-excel-outline class="text-icon" />
            </template>
            {{ $t('page.hrm.candidate.import') }}
          </NButton>
          <TableHeaderOperation
            v-model:columns="columnChecks"
            :disabled-delete="!hasCheckedRows"
            :loading="loading"
            add-auth="hrm:candidate:add"
            delete-auth="hrm:candidate:delete"
            @add="handleAdd"
            @delete="handleBatchDelete"
            @refresh="getData"
          />
        </NSpace>
      </template>
      <NDataTable
        v-model:checked-row-keys="checkedRowKeys"
        :columns="columns"
        :data="data"
        size="small"
        :flex-height="!appStore.isMobile"
        :scroll-x="1700"
        :loading="loading"
        remote
        :row-key="row => row.uuid"
        :pagination="mobilePagination"
        class="sm:h-full"
      />
      <CandidateOperateModal
        v-model:visible="operateVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="handleSubmitted"
      />
      <CandidateImportModal v-model:visible="importVisible" @submitted="getDataByPage" />
    </NCard>
  </div>
</template>

<style scoped></style>
