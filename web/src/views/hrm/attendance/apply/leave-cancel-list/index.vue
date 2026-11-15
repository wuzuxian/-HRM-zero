<script setup lang="tsx">
import { computed, reactive, ref } from 'vue';
import { NButton, NPopconfirm, NTag } from 'naive-ui';
import {
  AddLeave,
  ApproveBill,
  CancelLeave,
  DeleteLeave,
  EditLeave,
  GetApprovalRecordList,
  GetLeaveList,
  RejectBill,
  RevokeBill,
  SubmitBill
} from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { hrmBillStatusRecord } from '@/constants/business';
import { useTable } from '@/hooks/common/table';
import LeaveCancelSearch from './modules/leave-cancel-search.vue';
import LeaveCancelOperateModal from './modules/leave-cancel-operate-modal.vue';
import LeaveCancelModal from './modules/leave-cancel-modal.vue';

const appStore = useAppStore();

type LoadingStatus = Record<string, boolean>;
const deleteLoadingStatus = reactive<LoadingStatus>({});
const actionLoadingStatus = reactive<LoadingStatus>({});

/** 单据状态可修改/删除/提交的状态集（未提交/审批不通过/已撤销，演示站规则） */
const EDITABLE_STATUSES: Api.Hrm.BillStatus[] = ['unsubmitted', 'rejected', 'revoked'];

const HR205 = '205';

function isEditable(row: Api.Hrm.HrmLeave) {
  return row.status !== null && EDITABLE_STATUSES.includes(row.status);
}

function statusTag(row: Api.Hrm.HrmLeave) {
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
  apiFn: GetLeaveList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    billNo: null,
    status: null,
    leaveTypeCode: null
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
      title: $t('page.hrm.attendance.leaveCancel.billNo'),
      align: 'center',
      width: 170
    },
    {
      key: 'status',
      title: $t('page.hrm.attendance.billStatus.title'),
      align: 'center',
      width: 100,
      render: row => statusTag(row)
    },
    {
      key: 'leaveTypeName',
      title: $t('page.hrm.attendance.leaveCancel.leaveTypeName'),
      align: 'center',
      width: 90,
      render: row => (row.leaveTypeName ? <NTag size="small">{row.leaveTypeName}</NTag> : '-')
    },
    {
      key: 'planStartTime',
      title: $t('page.hrm.attendance.leaveCancel.planStartTime'),
      align: 'center',
      width: 150
    },
    {
      key: 'planEndTime',
      title: $t('page.hrm.attendance.leaveCancel.planEndTime'),
      align: 'center',
      width: 150
    },
    {
      key: 'planDays',
      title: $t('page.hrm.attendance.leaveCancel.planDays'),
      align: 'center',
      width: 80
    },
    {
      key: 'actualStartTime',
      title: $t('page.hrm.attendance.leaveCancel.actualStartTime'),
      align: 'center',
      width: 150,
      render: row => row.actualStartTime || '-'
    },
    {
      key: 'actualEndTime',
      title: $t('page.hrm.attendance.leaveCancel.actualEndTime'),
      align: 'center',
      width: 150,
      render: row => row.actualEndTime || '-'
    },
    {
      key: 'actualDays',
      title: $t('page.hrm.attendance.leaveCancel.actualDays'),
      align: 'center',
      width: 80,
      render: row => (row.settled === 1 ? row.actualDays : '-')
    },
    {
      key: 'employeeName',
      title: $t('page.hrm.attendance.leaveCancel.employeeName'),
      align: 'center',
      width: 90
    },
    {
      key: 'orgName',
      title: $t('page.hrm.attendance.leaveCancel.orgName'),
      align: 'center',
      width: 110,
      render: row => row.orgName || '-'
    },
    {
      key: 'createTime',
      title: $t('page.hrm.attendance.leaveCancel.createTime'),
      align: 'center',
      width: 160
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

function renderOperate(row: Api.Hrm.HrmLeave) {
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
  if (row.status === 'approved' && row.settled === 0) {
    actions.add('cancel');
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
            default: () => $t('page.hrm.attendance.leaveCancel.confirmSubmit'),
            trigger: () => (
              <NButton loading={actionLoadingStatus[row.uuid]} type="info" ghost size="small">
                {$t('page.hrm.attendance.leaveCancel.submit')}
              </NButton>
            )
          }}
        </NPopconfirm>
      )}
      {actions.has('approve') && (
        <NPopconfirm onPositiveClick={() => handleBillAction(row, ApproveBill)}>
          {{
            default: () => $t('page.hrm.attendance.leaveCancel.confirmApprove'),
            trigger: () => (
              <NButton loading={actionLoadingStatus[row.uuid]} type="success" ghost size="small">
                {$t('page.hrm.attendance.leaveCancel.approve')}
              </NButton>
            )
          }}
        </NPopconfirm>
      )}
      {actions.has('reject') && (
        <NPopconfirm onPositiveClick={() => handleBillAction(row, RejectBill)}>
          {{
            default: () => $t('page.hrm.attendance.leaveCancel.confirmReject'),
            trigger: () => (
              <NButton loading={actionLoadingStatus[row.uuid]} type="error" ghost size="small">
                {$t('page.hrm.attendance.leaveCancel.reject')}
              </NButton>
            )
          }}
        </NPopconfirm>
      )}
      {actions.has('revoke') && (
        <NPopconfirm onPositiveClick={() => handleBillAction(row, RevokeBill)}>
          {{
            default: () => $t('page.hrm.attendance.leaveCancel.confirmRevoke'),
            trigger: () => (
              <NButton loading={actionLoadingStatus[row.uuid]} type="warning" ghost size="small">
                {$t('page.hrm.attendance.leaveCancel.revoke')}
              </NButton>
            )
          }}
        </NPopconfirm>
      )}
      {actions.has('cancel') && (
        <NButton
          v-auth={['hrm:leave:cancel']}
          loading={actionLoadingStatus[row.uuid]}
          type="primary"
          ghost
          size="small"
          onClick={() => handleCancel(row)}
        >
          {$t('page.hrm.attendance.leaveCancel.cancelLeave')}
        </NButton>
      )}
      {actions.has('records') && (
        <NButton size="small" onClick={() => handleShowRecords(row)}>
          {$t('page.hrm.attendance.leaveCancel.records')}
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

// modal state (manual, HRM 页面不用 useTableOperate)
const operateVisible = ref(false);
const operateType = ref<NaiveUI.TableOperateType>('add');
const editingData = ref<Api.Hrm.HrmLeave | null>(null);

function handleAdd() {
  operateType.value = 'add';
  editingData.value = null;
  operateVisible.value = true;
}

function handleEdit(row: Api.Hrm.HrmLeave) {
  operateType.value = 'edit';
  editingData.value = { ...row };
  operateVisible.value = true;
}

function handleSubmitted() {
  getDataByPage();
}

type BillActionFn = (req: Api.Hrm.BillActionRequest) => Promise<{ data: unknown; error: unknown }>;

async function handleBillAction(row: Api.Hrm.HrmLeave, action: BillActionFn) {
  actionLoadingStatus[row.uuid] = true;
  const { error } = await action({ billType: HR205, billUuid: row.uuid });
  actionLoadingStatus[row.uuid] = false;
  if (error) return;
  window.$message?.success($t('common.updateSuccess'));
  getData();
}

async function handleDelete(uuid: string) {
  deleteLoadingStatus[uuid] = true;
  const { error } = await DeleteLeave([uuid]);
  deleteLoadingStatus[uuid] = false;
  if (error) return;
  window.$message?.success($t('common.deleteSuccess'));
  getData();
}

// ---- 销假 modal ----
const cancelVisible = ref(false);
const cancelRow = ref<Api.Hrm.HrmLeave | null>(null);

function handleCancel(row: Api.Hrm.HrmLeave) {
  cancelRow.value = { ...row };
  cancelVisible.value = true;
}

// ---- 审批记录 modal ----
const recordsVisible = ref(false);
const records = ref<Api.Hrm.ApprovalRecord[]>([]);
const recordsBillNo = ref('');

async function handleShowRecords(row: Api.Hrm.HrmLeave) {
  recordsBillNo.value = row.billNo;
  const { data: recordResp, error } = await GetApprovalRecordList({ billType: HR205, billUuid: row.uuid });
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
    submit: 'page.hrm.attendance.leaveCancel.submit',
    approve: 'page.hrm.attendance.leaveCancel.approve',
    reject: 'page.hrm.attendance.leaveCancel.reject',
    revoke: 'page.hrm.attendance.leaveCancel.revoke'
  };
  const key = keys[action];
  return key ? $t(key as App.I18n.I18nKey) : action;
}

const dayPrecision = computed(() => $t('page.hrm.attendance.leaveCancel.dayPrecision'));
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <LeaveCancelSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <NCard
      :title="$t('page.hrm.attendance.leaveCancel.title')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <div class="flex items-center gap-8px">
          <span class="text-12px text-gray-400">{{ dayPrecision }}</span>
          <TableHeaderOperation
            v-model:columns="columnChecks"
            :loading="loading"
            add-auth="hrm:leave:add"
            delete-auth="hrm:leave:delete"
            @add="handleAdd"
            @refresh="getData"
          />
        </div>
      </template>
      <NDataTable
        :columns="columns"
        :data="data"
        size="small"
        :flex-height="!appStore.isMobile"
        :scroll-x="1800"
        :loading="loading"
        remote
        :pagination="mobilePagination"
        class="sm:h-full"
      />
      <LeaveCancelOperateModal
        v-model:visible="operateVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="handleSubmitted"
      />
      <LeaveCancelModal v-model:visible="cancelVisible" :row-data="cancelRow" @submitted="getData" />
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
        <NEmpty v-else :description="$t('page.hrm.attendance.leaveCancel.noRecords')" />
      </NModal>
    </NCard>
  </div>
</template>

<style scoped></style>
