<script setup lang="tsx">
import { computed, reactive, ref } from 'vue';
import { NButton, NPopconfirm, NTag } from 'naive-ui';
import {
  ApproveBill,
  DeleteTrainingPlan,
  GetApprovalRecordList,
  GetTrainingPlanList,
  RejectBill,
  RevokeBill,
  SubmitBill
} from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import {
  hrmBillStatusRecord,
  hrmTrainingPlanPublishedRecord,
  hrmTrainingPlanTypeRecord
} from '@/constants/business';
import { useTable } from '@/hooks/common/table';
import PlanOperateModal from './modules/plan-operate-modal.vue';
import PlanSearch from './modules/plan-search.vue';

const appStore = useAppStore();

type LoadingStatus = Record<string, boolean>;
const deleteLoadingStatus = reactive<LoadingStatus>({});
const actionLoadingStatus = reactive<LoadingStatus>({});

/** 单据状态可修改/删除/提交的状态集（未提交/审批不通过/已撤销，演示站规则） */
const EDITABLE_STATUSES: Api.Hrm.BillStatus[] = ['unsubmitted', 'rejected', 'revoked'];

const HR216 = '216';

function isEditable(row: Api.Hrm.TrainingPlan) {
  return (
    row.status !== null &&
    EDITABLE_STATUSES.includes(row.status) &&
    row.published === 0
  );
}

function statusTag(row: Api.Hrm.TrainingPlan) {
  if (row.status === null) return null;
  const tagMap: Record<Api.Hrm.BillStatus, NaiveUI.ThemeColor> = {
    unsubmitted: 'default',
    approving: 'info',
    approved: 'success',
    rejected: 'error',
    revoked: 'warning'
  };
  return <NTag type={tagMap[row.status]}>{$t(hrmBillStatusRecord[row.status])}</NTag>;
}

function publishedTag(row: Api.Hrm.TrainingPlan) {
  if (row.published === 1) {
    return (
      <NTag type="success" size="small">
        {$t(hrmTrainingPlanPublishedRecord['1'])}
      </NTag>
    );
  }
  return <NTag size="small">{$t(hrmTrainingPlanPublishedRecord['0'])}</NTag>;
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
  apiFn: GetTrainingPlanList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    planName: null,
    planType: null,
    status: null,
    published: null
  },
  columns: () => [
    {
      key: 'index',
      title: $t('common.index'),
      align: 'center',
      width: 64
    },
    {
      key: 'billNo',
      title: $t('page.hrm.training.plan.billNo'),
      align: 'center',
      width: 170
    },
    {
      key: 'planName',
      title: $t('page.hrm.training.plan.planName'),
      align: 'center',
      minWidth: 160
    },
    {
      key: 'theme',
      title: $t('page.hrm.training.plan.theme'),
      align: 'center',
      minWidth: 120,
      render: row => row.theme || '-'
    },
    {
      key: 'planYear',
      title: $t('page.hrm.training.plan.planYear'),
      align: 'center',
      width: 80
    },
    {
      key: 'planType',
      title: $t('page.hrm.training.plan.planType'),
      align: 'center',
      width: 100,
      render: row =>
        row.planType === null ? null : <NTag size="small">{$t(hrmTrainingPlanTypeRecord[row.planType])}</NTag>
    },
    {
      key: 'startDate',
      title: $t('page.hrm.training.plan.startDate'),
      align: 'center',
      width: 110
    },
    {
      key: 'endDate',
      title: $t('page.hrm.training.plan.endDate'),
      align: 'center',
      width: 110
    },
    {
      key: 'status',
      title: $t('page.hrm.training.plan.billStatusLabel'),
      align: 'center',
      width: 100,
      render: row => statusTag(row)
    },
    {
      key: 'published',
      title: $t('page.hrm.training.plan.published'),
      align: 'center',
      width: 100,
      render: row => publishedTag(row)
    },
    {
      key: 'creatorName',
      title: $t('page.hrm.training.plan.creatorName'),
      align: 'center',
      width: 90
    },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      width: 300,
      fixed: 'right',
      render: row => renderOperate(row)
    }
  ]
});

function renderOperate(row: Api.Hrm.TrainingPlan) {
  const actions = new Set<string>();

  if (isEditable(row)) {
    actions.add('edit');
    actions.add('delete');
    actions.add('submit');
  }
  if (row.status === 'approving') {
    actions.add('approve');
    actions.add('reject');
    actions.add('revoke');
  }
  actions.add('records');

  return (
    <div class="flex-center flex-wrap gap-4px">
      {actions.has('edit') && (
        <NButton type="primary" ghost size="small" onClick={() => handleEdit(row)}>
          {$t('common.edit')}
        </NButton>
      )}
      {actions.has('submit') && (
        <NPopconfirm onPositiveClick={() => handleBillAction(row, SubmitBill)}>
          {{
            default: () => $t('page.hrm.training.plan.confirmSubmit'),
            trigger: () => (
              <NButton loading={actionLoadingStatus[row.uuid]} type="info" ghost size="small">
                {$t('page.hrm.training.plan.submit')}
              </NButton>
            )
          }}
        </NPopconfirm>
      )}
      {actions.has('approve') && (
        <NPopconfirm onPositiveClick={() => handleBillAction(row, ApproveBill)}>
          {{
            default: () => $t('page.hrm.training.plan.confirmApprove'),
            trigger: () => (
              <NButton loading={actionLoadingStatus[row.uuid]} type="success" ghost size="small">
                {$t('page.hrm.training.plan.approve')}
              </NButton>
            )
          }}
        </NPopconfirm>
      )}
      {actions.has('reject') && (
        <NPopconfirm onPositiveClick={() => handleBillAction(row, RejectBill)}>
          {{
            default: () => $t('page.hrm.training.plan.confirmReject'),
            trigger: () => (
              <NButton loading={actionLoadingStatus[row.uuid]} type="error" ghost size="small">
                {$t('page.hrm.training.plan.reject')}
              </NButton>
            )
          }}
        </NPopconfirm>
      )}
      {actions.has('revoke') && (
        <NPopconfirm onPositiveClick={() => handleBillAction(row, RevokeBill)}>
          {{
            default: () => $t('page.hrm.training.plan.confirmRevoke'),
            trigger: () => (
              <NButton loading={actionLoadingStatus[row.uuid]} type="warning" ghost size="small">
                {$t('page.hrm.training.plan.revoke')}
              </NButton>
            )
          }}
        </NPopconfirm>
      )}
      {actions.has('records') && (
        <NButton size="small" onClick={() => handleShowRecords(row)}>
          {$t('page.hrm.training.plan.records')}
        </NButton>
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
const editingData = ref<Api.Hrm.TrainingPlan | null>(null);

function handleAdd() {
  operateType.value = 'add';
  editingData.value = null;
  operateVisible.value = true;
}

function handleEdit(row: Api.Hrm.TrainingPlan) {
  operateType.value = 'edit';
  editingData.value = { ...row };
  operateVisible.value = true;
}

function handleSubmitted() {
  getDataByPage();
}

type BillActionFn = (req: Api.Hrm.BillActionRequest) => Promise<{ data: unknown; error: unknown }>;

async function handleBillAction(row: Api.Hrm.TrainingPlan, action: BillActionFn) {
  actionLoadingStatus[row.uuid] = true;
  const { error } = await action({ billType: HR216, billUuid: row.uuid });
  actionLoadingStatus[row.uuid] = false;
  if (error) return;
  window.$message?.success($t('common.updateSuccess'));
  getData();
}

async function handleBatchDelete() {
  const uuids: string[] = [...checkedRowKeys.value];

  uuids.forEach(uuid => {
    deleteLoadingStatus[uuid] = true;
  });
  const { error } = await DeleteTrainingPlan(uuids);

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
  const { error } = await DeleteTrainingPlan(uuids);
  deleteLoadingStatus[uuid] = false;
  if (error) {
    return;
  }
  window.$message?.success($t('common.deleteSuccess'));
  getData();
}

// ---- 审批记录 modal ----
const recordsVisible = ref(false);
const records = ref<Api.Hrm.ApprovalRecord[]>([]);
const recordsBillNo = ref('');

async function handleShowRecords(row: Api.Hrm.TrainingPlan) {
  recordsBillNo.value = row.billNo;
  const { data: recordResp, error } = await GetApprovalRecordList({ billType: HR216, billUuid: row.uuid });
  if (error) return;
  records.value = recordResp?.records ?? [];
  recordsVisible.value = true;
}

const recordActionTagMap: Record<string, NaiveUI.ThemeColor> = {
  submit: 'info',
  approve: 'success',
  reject: 'error',
  revoke: 'warning'
};

function recordStatusLabel(status: string) {
  const key = (hrmBillStatusRecord as Record<string, App.I18n.I18nKey>)[status];
  return key ? $t(key) : status;
}

function recordActionLabel(action: string) {
  const keys: Record<string, string> = {
    submit: 'page.hrm.training.plan.submit',
    approve: 'page.hrm.training.plan.approve',
    reject: 'page.hrm.training.plan.reject',
    revoke: 'page.hrm.training.plan.revoke'
  };
  const key = keys[action];
  return key ? $t(key as App.I18n.I18nKey) : action;
}
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <PlanSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <NCard
      :title="$t('page.hrm.training.plan.title')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <TableHeaderOperation
          v-model:columns="columnChecks"
          :disabled-delete="!hasCheckedRows"
          :loading="loading"
          add-auth="hrm:training_plan:add"
          delete-auth="hrm:training_plan:delete"
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
      <PlanOperateModal
        v-model:visible="operateVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="handleSubmitted"
      />
      <NModal v-model:show="recordsVisible" preset="card" class="w-560px" :title="recordsBillNo">
        <NList v-if="records.length > 0">
          <NListItem v-for="(rec, idx) in records" :key="idx">
            <div class="flex flex-col gap-4px">
              <div class="flex items-center gap-8px">
                <NTag :type="recordActionTagMap[rec.action] ?? 'default'" size="small">
                  {{ recordActionLabel(rec.action) }}
                </NTag>
                <span class="text-12px text-gray-500">{{ rec.createTime }}</span>
                <span class="text-12px">{{ rec.operatorName }}</span>
              </div>
              <div class="text-12px text-gray-500">
                {{ recordStatusLabel(rec.fromStatus) }} → {{ recordStatusLabel(rec.toStatus) }}
              </div>
              <div v-if="rec.opinion" class="text-12px">{{ rec.opinion }}</div>
            </div>
          </NListItem>
        </NList>
        <NEmpty v-else :description="$t('page.hrm.employeeRelation.transfer.noRecords')" />
      </NModal>
    </NCard>
  </div>
</template>

<style scoped></style>
