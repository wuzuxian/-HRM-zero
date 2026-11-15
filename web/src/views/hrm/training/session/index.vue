<script setup lang="tsx">
import { computed, reactive, ref } from 'vue';
import { NButton, NPopconfirm, NTag } from 'naive-ui';
import { DeleteTrainingSession, FinishTrainingSession, GetTrainingSessionList } from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { hrmTrainingSessionStatusRecord } from '@/constants/business';
import { useAuth } from '@/hooks/business/auth';
import { useTable } from '@/hooks/common/table';
import SessionOperateModal from './modules/session-operate-modal.vue';
import SessionSearch from './modules/session-search.vue';
import SessionParticipantsModal from './modules/session-participants-modal.vue';

const appStore = useAppStore();
const { hasAuth } = useAuth();

const canAdd = computed(() => hasAuth('hrm:training_session:add'));
const canEdit = computed(() => hasAuth('hrm:training_session:edit'));
const canParticipants = computed(() => hasAuth('hrm:training_session:setSessionParticipants'));

type LoadingStatus = Record<string, boolean>;
const deleteLoadingStatus = reactive<LoadingStatus>({});
const finishLoadingStatus = reactive<LoadingStatus>({});

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
  apiFn: GetTrainingSessionList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    sessionName: null,
    planUuid: null,
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
      key: 'sessionNo',
      title: $t('page.hrm.training.sessionNo'),
      align: 'center',
      width: 130
    },
    {
      key: 'sessionName',
      title: $t('page.hrm.training.session.sessionName'),
      align: 'center',
      minWidth: 140
    },
    {
      key: 'planName',
      title: $t('page.hrm.training.session.planName'),
      align: 'center',
      minWidth: 140,
      render: row => row.planName || '-'
    },
    {
      key: 'lecturer',
      title: $t('page.hrm.training.lecturer'),
      align: 'center',
      width: 90,
      render: row => row.lecturer || '-'
    },
    {
      key: 'location',
      title: $t('page.hrm.training.location'),
      align: 'center',
      width: 120,
      render: row => row.location || '-'
    },
    {
      key: 'hours',
      title: $t('page.hrm.training.hours'),
      align: 'center',
      width: 80,
      render: row => `${row.hours} h`
    },
    {
      key: 'startTime',
      title: $t('page.hrm.training.startTime'),
      align: 'center',
      width: 150
    },
    {
      key: 'endTime',
      title: $t('page.hrm.training.endTime'),
      align: 'center',
      width: 150
    },
    {
      key: 'status',
      title: $t('page.hrm.training.sessionStatusLabel'),
      align: 'center',
      width: 90,
      render: row => {
        if (row.status === null) {
          return null;
        }
        const tagMap: Record<Api.Hrm.TrainingSessionStatus, NaiveUI.ThemeColor> = {
          '1': 'info',
          '2': 'warning',
          '3': 'success'
        };
        return <NTag type={tagMap[row.status]}>{$t(hrmTrainingSessionStatusRecord[row.status])}</NTag>;
      }
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

function renderOperate(row: Api.Hrm.TrainingSession) {
  const finished = row.status === '3';
  return (
    <div class="flex-center flex-wrap gap-4px">
      {canEdit.value && !finished && (
        <NButton type="primary" ghost size="small" onClick={() => handleEdit(row)}>
          {$t('common.edit')}
        </NButton>
      )}
      {/* 签到 (HRM-29) seam: 行操作占位，后续工单接入课堂动态码/补签 */}
      {!finished && canParticipants.value && (
        <NButton type="info" ghost size="small" onClick={() => handleParticipants(row)}>
          {$t('page.hrm.training.session.participants')}
        </NButton>
      )}
      {!finished && (
        <NPopconfirm onPositiveClick={() => handleFinish(row.uuid)}>
          {{
            default: () => $t('page.hrm.training.session.confirmFinish'),
            trigger: () => (
              <NButton loading={finishLoadingStatus[row.uuid]} type="warning" ghost size="small">
                {$t('page.hrm.training.session.finish')}
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

// modal state (manual, same pattern as 组织管理 page)
const operateVisible = ref(false);
const operateType = ref<NaiveUI.TableOperateType>('add');
const editingData = ref<Api.Hrm.TrainingSession | null>(null);

function handleAdd() {
  operateType.value = 'add';
  editingData.value = null;
  operateVisible.value = true;
}

function handleEdit(row: Api.Hrm.TrainingSession) {
  operateType.value = 'edit';
  editingData.value = { ...row };
  operateVisible.value = true;
}

function handleSubmitted() {
  getDataByPage();
}

// ---- 学员绑定 modal ----
const participantsVisible = ref(false);
const participantsSessionUuid = ref('');
const participantsSessionName = ref('');

function handleParticipants(row: Api.Hrm.TrainingSession) {
  participantsSessionUuid.value = row.uuid;
  participantsSessionName.value = row.sessionName;
  participantsVisible.value = true;
}

async function handleFinish(uuid: string) {
  finishLoadingStatus[uuid] = true;
  const { error } = await FinishTrainingSession(uuid);
  finishLoadingStatus[uuid] = false;
  if (error) {
    return;
  }
  window.$message?.success($t('common.updateSuccess'));
  getData();
}

const checkedRowKeys = ref<string[]>([]);
const hasCheckedRows = computed(() => checkedRowKeys.value.length === 0);

async function handleBatchDelete() {
  const uuids: string[] = [...checkedRowKeys.value];

  uuids.forEach(uuid => {
    deleteLoadingStatus[uuid] = true;
  });
  const { error } = await DeleteTrainingSession(uuids);

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
  const { error } = await DeleteTrainingSession(uuids);
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
    <SessionSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <NCard
      :title="$t('page.hrm.training.session.title')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <TableHeaderOperation
          v-model:columns="columnChecks"
          :disabled-delete="!hasCheckedRows"
          :loading="loading"
          add-auth="hrm:training_session:add"
          delete-auth="hrm:training_session:delete"
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
        :scroll-x="1400"
        :loading="loading"
        remote
        :row-key="row => row.uuid"
        :pagination="mobilePagination"
        class="sm:h-full"
      />
      <SessionOperateModal
        v-model:visible="operateVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="handleSubmitted"
      />
      <SessionParticipantsModal
        v-model:visible="participantsVisible"
        :session-uuid="participantsSessionUuid"
        :session-name="participantsSessionName"
        @submitted="getData"
      />
    </NCard>
  </div>
</template>

<style scoped></style>
