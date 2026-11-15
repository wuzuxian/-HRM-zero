<script setup lang="tsx">
import { computed, reactive, ref } from 'vue';
import { NButton, NPopconfirm, NTag } from 'naive-ui';
import {
  ApproveBill,
  DeleteOvertimeBill,
  GetApprovalRecordList,
  GetOvertimeBillList,
  RejectBill,
  RevokeBill,
  SubmitOvertimeBill
} from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { hrmBillStatusRecord, hrmOvertimeDayTypeRecord } from '@/constants/business';
import { useTable } from '@/hooks/common/table';
import OvertimeBillSearch from './modules/overtime-bill-search.vue';
import OvertimeBillOperateModal from './modules/overtime-bill-operate-modal.vue';
import OvertimeBillConfirmModal from './modules/overtime-bill-confirm-modal.vue';

const appStore = useAppStore();

type LoadingStatus = Record<string, boolean>;
const deleteLoadingStatus = reactive<LoadingStatus>({});
const actionLoadingStatus = reactive<LoadingStatus>({});

/** 单据状态可修改/删除/提交的状态集（未提交/审批不通过/已撤销，演示站规则） */
const EDITABLE_STATUSES: Api.Hrm.BillStatus[] = ['unsubmitted', 'rejected', 'revoked'];

const HR217 = '217';

function isEditable(row: Api.Hrm.HrmOvertimeBill) {
  return row.status !== null && EDITABLE_STATUSES.includes(row.status);
}

function statusTag(row: Api.Hrm.HrmOvertimeBill) {
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

function dayTypeTag(row: Api.Hrm.HrmOvertimeBill) {
  if (!row.dayType) return '-';
  const key = hrmOvertimeDayTypeRecord[row.dayType];
  return key ? $t(key) : row.dayType;
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
  apiFn: GetOvertimeBillList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    billNo: null,
    status: null,
    dayType: null
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
      title: $t('page.hrm.attendance.overtimeBill.billNo'),
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
      key: 'overtimeDate',
      title: $t('page.hrm.attendance.overtimeBill.overtimeDate'),
      align: 'center',
      width: 110
    },
    {
      key: 'dayType',
      title: $t('page.hrm.attendance.overtimeBill.dayType'),
      align: 'center',
      width: 100,
      render: row => dayTypeTag(row)
    },
    {
      key: 'startTime',
      title: $t('page.hrm.attendance.overtimeBill.startTime'),
      align: 'center',
      width: 150
    },
    {
      key: 'endTime',
      title: $t('page.hrm.attendance.overtimeBill.endTime'),
      align: 'center',
      width: 150
    },
    {
      key: 'applyHours',
      title: $t('page.hrm.attendance.overtimeBill.applyHours'),
      align: 'center',
      width: 90
    },
    {
      key: 'confirmHours',
      title: $t('page.hrm.attendance.overtimeBill.confirmHours'),
      align: 'center',
      width: 90,
      render: row => (row.confirmHours ? row.confirmHours : '-')
    },
    {
      key: 'compType',
      title: $t('page.hrm.attendance.overtimeBill.compType'),
      align: 'center',
      width: 90,
      render: row => compTypeLabel(row)
    },
    {
      key: 'source',
      title: $t('page.hrm.attendance.overtimeBill.source'),
      align: 'center',
      width: 90,
      render: row => (row.source === 'employee' ? $t('page.hrm.attendance.overtimeBill.sourceEmployee') : row.source)
    },
    {
      key: 'employeeName',
      title: $t('page.hrm.attendance.overtimeBill.employeeName'),
      align: 'center',
      width: 90
    },
    {
      key: 'orgName',
      title: $t('page.hrm.attendance.overtimeBill.orgName'),
      align: 'center',
      width: 110,
      render: row => row.orgName || '-'
    },
    {
      key: 'createTime',
      title: $t('page.hrm.attendance.overtimeBill.createTime'),
      align: 'center',
      width: 160
    },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      width: 340,
      fixed: 'right',
      render: row => renderOperate(row)
    }
  ]
});

function compTypeLabel(row: Api.Hrm.HrmOvertimeBill) {
  if (row.compType === 'comp_leave') return $t('page.hrm.attendance.overtimeBill.compTypeLeave');
  if (row.compType === 'other') return $t('page.hrm.attendance.overtimeBill.compTypeOther');
  return '-';
}

function renderOperate(row: Api.Hrm.HrmOvertimeBill) {
  const actions = new Set<string>();

  if (isEditable(row)) {
    actions.add('edit');
    actions.add('delete');
    actions.add('submit');
  }
  if (row.status === 'approving') {
    actions.add('confirm');
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
        <NPopconfirm onPositiveClick={() => handleSubmitBill(row)}>
          {{
            default: () => $t('page.hrm.attendance.overtimeBill.confirmSubmit'),
            trigger: () => (
              <NButton loading={actionLoadingStatus[row.uuid]} type="info" ghost size="small">
                {$t('page.hrm.attendance.overtimeBill.submit')}
              </NButton>
            )
          }}
        </NPopconfirm>
      )}
      {actions.has('confirm') && (
        <NButton
          v-auth={['hrm:overtimebill:confirm']}
          loading={actionLoadingStatus[row.uuid]}
          type="primary"
          ghost
          size="small"
          onClick={() => handleConfirm(row)}
        >
          {$t('page.hrm.attendance.overtimeBill.confirm')}
        </NButton>
      )}
      {actions.has('approve') && (
        <NPopconfirm onPositiveClick={() => handleBillAction(row, ApproveBill)}>
          {{
            default: () => $t('page.hrm.attendance.overtimeBill.confirmApprove'),
            trigger: () => (
              <NButton loading={actionLoadingStatus[row.uuid]} type="success" ghost size="small">
                {$t('page.hrm.attendance.overtimeBill.approve')}
              </NButton>
            )
          }}
        </NPopconfirm>
      )}
      {actions.has('reject') && (
        <NPopconfirm onPositiveClick={() => handleBillAction(row, RejectBill)}>
          {{
            default: () => $t('page.hrm.attendance.overtimeBill.confirmReject'),
            trigger: () => (
              <NButton loading={actionLoadingStatus[row.uuid]} type="error" ghost size="small">
                {$t('page.hrm.attendance.overtimeBill.reject')}
              </NButton>
            )
          }}
        </NPopconfirm>
      )}
      {actions.has('revoke') && (
        <NPopconfirm onPositiveClick={() => handleBillAction(row, RevokeBill)}>
          {{
            default: () => $t('page.hrm.attendance.overtimeBill.confirmRevoke'),
            trigger: () => (
              <NButton loading={actionLoadingStatus[row.uuid]} type="warning" ghost size="small">
                {$t('page.hrm.attendance.overtimeBill.revoke')}
              </NButton>
            )
          }}
        </NPopconfirm>
      )}
      {actions.has('records') && (
        <NButton size="small" onClick={() => handleShowRecords(row)}>
          {$t('page.hrm.attendance.overtimeBill.records')}
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
const editingData = ref<Api.Hrm.HrmOvertimeBill | null>(null);

function handleAdd() {
  operateType.value = 'add';
  editingData.value = null;
  operateVisible.value = true;
}

function handleEdit(row: Api.Hrm.HrmOvertimeBill) {
  operateType.value = 'edit';
  editingData.value = { ...row };
  operateVisible.value = true;
}

function handleSubmitted() {
  getDataByPage();
}

type BillActionFn = (req: Api.Hrm.BillActionRequest) => Promise<{ data: unknown; error: unknown }>;

/**
 * 提交走 SubmitOvertimeBill 判定口径的前端简化：先 Add 不需要，提交用共享底座； 免审批自动通过由后端 SubmitOvertimeBill 完成（前端只调共享 submit）— 此处直接调共享
 * SubmitBill；后端 overtimebill submit 会先判规则， 免审批单在提交后立即变 审批通过。
 */
async function handleSubmitBill(row: Api.Hrm.HrmOvertimeBill) {
  actionLoadingStatus[row.uuid] = true;
  const { error } = await SubmitOvertimeBill({ uuid: row.uuid });
  actionLoadingStatus[row.uuid] = false;
  if (error) return;
  window.$message?.success($t('common.updateSuccess'));
  getData();
}

async function handleBillAction(row: Api.Hrm.HrmOvertimeBill, action: BillActionFn) {
  actionLoadingStatus[row.uuid] = true;
  const { error } = await action({ billType: HR217, billUuid: row.uuid });
  actionLoadingStatus[row.uuid] = false;
  if (error) return;
  window.$message?.success($t('common.updateSuccess'));
  getData();
}

async function handleDelete(uuid: string) {
  deleteLoadingStatus[uuid] = true;
  const { error } = await DeleteOvertimeBill([uuid]);
  deleteLoadingStatus[uuid] = false;
  if (error) return;
  window.$message?.success($t('common.deleteSuccess'));
  getData();
}

// ---- 核定 modal ----
const confirmVisible = ref(false);
const confirmRow = ref<Api.Hrm.HrmOvertimeBill | null>(null);

function handleConfirm(row: Api.Hrm.HrmOvertimeBill) {
  confirmRow.value = { ...row };
  confirmVisible.value = true;
}

async function handleConfirmSubmitted() {
  await getData();
}

// ---- 审批记录 modal ----
const recordsVisible = ref(false);
const records = ref<Api.Hrm.ApprovalRecord[]>([]);
const recordsBillNo = ref('');

async function handleShowRecords(row: Api.Hrm.HrmOvertimeBill) {
  recordsBillNo.value = row.billNo;
  const { data: recordResp, error } = await GetApprovalRecordList({ billType: HR217, billUuid: row.uuid });
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
    submit: 'page.hrm.attendance.overtimeBill.submit',
    approve: 'page.hrm.attendance.overtimeBill.approve',
    reject: 'page.hrm.attendance.overtimeBill.reject',
    revoke: 'page.hrm.attendance.overtimeBill.revoke'
  };
  const key = keys[action];
  return key ? $t(key as App.I18n.I18nKey) : action;
}

const hoursPrecision = computed(() => $t('page.hrm.attendance.overtimeBill.hoursPrecision'));
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <OvertimeBillSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <NCard
      :title="$t('page.hrm.attendance.overtimeBill.title')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <div class="flex items-center gap-8px">
          <span class="text-12px text-gray-400">{{ hoursPrecision }}</span>
          <TableHeaderOperation
            v-model:columns="columnChecks"
            :loading="loading"
            add-auth="hrm:overtimebill:add"
            delete-auth="hrm:overtimebill:delete"
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
        :scroll-x="2000"
        :loading="loading"
        remote
        :pagination="mobilePagination"
        class="sm:h-full"
      />
      <OvertimeBillOperateModal
        v-model:visible="operateVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="handleSubmitted"
      />
      <OvertimeBillConfirmModal
        v-model:visible="confirmVisible"
        :row-data="confirmRow"
        @submitted="handleConfirmSubmitted"
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
        <NEmpty v-else :description="$t('page.hrm.attendance.overtimeBill.noRecords')" />
      </NModal>
    </NCard>
  </div>
</template>

<style scoped></style>
