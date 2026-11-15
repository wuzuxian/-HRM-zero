<script setup lang="tsx">
import { reactive, ref } from 'vue';
import { NButton, NPopconfirm, NTag } from 'naive-ui';
import {
  AddCardReplacement,
  ApproveBill,
  DeleteCardReplacement,
  EditCardReplacement,
  GetApprovalRecordList,
  GetCardReplacementList,
  RejectBill,
  RevokeBill,
  SubmitBill
} from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { hrmBillStatusRecord } from '@/constants/business';
import { useTable } from '@/hooks/common/table';
import CardReplacementSearch from './modules/card-replacement-search.vue';
import CardReplacementOperateModal from './modules/card-replacement-operate-modal.vue';

const appStore = useAppStore();

type LoadingStatus = Record<string, boolean>;
const deleteLoadingStatus = reactive<LoadingStatus>({});
const actionLoadingStatus = reactive<LoadingStatus>({});

/** 单据状态可修改/删除/提交的状态集（未提交/审批不通过/已撤销，演示站规则） */
const EDITABLE_STATUSES: Api.Hrm.BillStatus[] = ['unsubmitted', 'rejected', 'revoked'];

const HR206 = '206';

function isEditable(row: Api.Hrm.HrmCardReplacement) {
  return row.status !== null && EDITABLE_STATUSES.includes(row.status);
}

function statusTag(row: Api.Hrm.HrmCardReplacement) {
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

function cardTypeTag(row: Api.Hrm.HrmCardReplacement) {
  const type = row.cardType === 'clockIn' ? 'info' : 'warning';
  return <NTag type={type} size="small">{$t(`page.hrm.attendance.cardReplacement.${row.cardType}` as App.I18n.I18nKey)}</NTag>;
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
  apiFn: GetCardReplacementList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    billNo: null,
    status: null,
    employeeNo: null,
    cardType: null
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
      title: $t('page.hrm.attendance.cardReplacement.billNo'),
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
      key: 'employeeNo',
      title: $t('page.hrm.attendance.cardReplacement.employeeNo'),
      align: 'center',
      width: 90
    },
    {
      key: 'employeeName',
      title: $t('page.hrm.attendance.cardReplacement.employeeName'),
      align: 'center',
      width: 90
    },
    {
      key: 'orgName',
      title: $t('page.hrm.attendance.cardReplacement.orgName'),
      align: 'center',
      width: 110,
      render: row => row.orgName || '-'
    },
    {
      key: 'punchDate',
      title: $t('page.hrm.attendance.cardReplacement.punchDate'),
      align: 'center',
      width: 110
    },
    {
      key: 'cardType',
      title: $t('page.hrm.attendance.cardReplacement.cardType'),
      align: 'center',
      width: 90,
      render: row => cardTypeTag(row)
    },
    {
      key: 'punchTime',
      title: $t('page.hrm.attendance.cardReplacement.punchTime'),
      align: 'center',
      width: 160
    },
    {
      key: 'origClockInTime',
      title: $t('page.hrm.attendance.cardReplacement.origClockInTime'),
      align: 'center',
      width: 160,
      render: row => row.origClockInTime || '-'
    },
    {
      key: 'origClockOutTime',
      title: $t('page.hrm.attendance.cardReplacement.origClockOutTime'),
      align: 'center',
      width: 160,
      render: row => row.origClockOutTime || '-'
    },
    {
      key: 'reason',
      title: $t('page.hrm.attendance.cardReplacement.reason'),
      align: 'center',
      width: 140,
      render: row => row.reason || '-'
    },
    {
      key: 'creatorName',
      title: $t('page.hrm.attendance.cardReplacement.creatorName'),
      align: 'center',
      width: 90
    },
    {
      key: 'createTime',
      title: $t('page.hrm.attendance.cardReplacement.createTime'),
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

function renderOperate(row: Api.Hrm.HrmCardReplacement) {
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
            default: () => $t('page.hrm.attendance.cardReplacement.confirmSubmit'),
            trigger: () => (
              <NButton loading={actionLoadingStatus[row.uuid]} type="info" ghost size="small">
                {$t('page.hrm.attendance.cardReplacement.submit')}
              </NButton>
            )
          }}
        </NPopconfirm>
      )}
      {actions.has('approve') && (
        <NPopconfirm onPositiveClick={() => handleBillAction(row, ApproveBill)}>
          {{
            default: () => $t('page.hrm.attendance.cardReplacement.confirmApprove'),
            trigger: () => (
              <NButton loading={actionLoadingStatus[row.uuid]} type="success" ghost size="small">
                {$t('page.hrm.attendance.cardReplacement.approve')}
              </NButton>
            )
          }}
        </NPopconfirm>
      )}
      {actions.has('reject') && (
        <NPopconfirm onPositiveClick={() => handleBillAction(row, RejectBill)}>
          {{
            default: () => $t('page.hrm.attendance.cardReplacement.confirmReject'),
            trigger: () => (
              <NButton loading={actionLoadingStatus[row.uuid]} type="error" ghost size="small">
                {$t('page.hrm.attendance.cardReplacement.reject')}
              </NButton>
            )
          }}
        </NPopconfirm>
      )}
      {actions.has('revoke') && (
        <NPopconfirm onPositiveClick={() => handleBillAction(row, RevokeBill)}>
          {{
            default: () => $t('page.hrm.attendance.cardReplacement.confirmRevoke'),
            trigger: () => (
              <NButton loading={actionLoadingStatus[row.uuid]} type="warning" ghost size="small">
                {$t('page.hrm.attendance.cardReplacement.revoke')}
              </NButton>
            )
          }}
        </NPopconfirm>
      )}
      {actions.has('records') && (
        <NButton size="small" onClick={() => handleShowRecords(row)}>
          {$t('page.hrm.attendance.cardReplacement.records')}
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
const editingData = ref<Api.Hrm.HrmCardReplacement | null>(null);

function handleAdd() {
  operateType.value = 'add';
  editingData.value = null;
  operateVisible.value = true;
}

function handleEdit(row: Api.Hrm.HrmCardReplacement) {
  operateType.value = 'edit';
  editingData.value = { ...row };
  operateVisible.value = true;
}

function handleSubmitted() {
  getDataByPage();
}

type BillActionFn = (req: Api.Hrm.BillActionRequest) => Promise<{ data: unknown; error: unknown }>;

async function handleBillAction(row: Api.Hrm.HrmCardReplacement, action: BillActionFn) {
  actionLoadingStatus[row.uuid] = true;
  const { error } = await action({ billType: HR206, billUuid: row.uuid });
  actionLoadingStatus[row.uuid] = false;
  if (error) return;
  window.$message?.success($t('common.updateSuccess'));
  getData();
}

async function handleDelete(uuid: string) {
  deleteLoadingStatus[uuid] = true;
  const { error } = await DeleteCardReplacement([uuid]);
  deleteLoadingStatus[uuid] = false;
  if (error) return;
  window.$message?.success($t('common.deleteSuccess'));
  getData();
}

// ---- 审批记录 modal ----
const recordsVisible = ref(false);
const records = ref<Api.Hrm.ApprovalRecord[]>([]);
const recordsBillNo = ref('');

async function handleShowRecords(row: Api.Hrm.HrmCardReplacement) {
  recordsBillNo.value = row.billNo;
  const { data: recordResp, error } = await GetApprovalRecordList({ billType: HR206, billUuid: row.uuid });
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
    submit: 'page.hrm.attendance.cardReplacement.submit',
    approve: 'page.hrm.attendance.cardReplacement.approve',
    reject: 'page.hrm.attendance.cardReplacement.reject',
    revoke: 'page.hrm.attendance.cardReplacement.revoke'
  };
  const key = keys[action];
  return key ? $t(key as App.I18n.I18nKey) : action;
}
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <CardReplacementSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <NCard
      :title="$t('page.hrm.attendance.cardReplacement.title')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <TableHeaderOperation
          v-model:columns="columnChecks"
          :loading="loading"
          add-auth="hrm:cardreplacement:add"
          delete-auth="hrm:cardreplacement:delete"
          @add="handleAdd"
          @refresh="getData"
        />
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
      <CardReplacementOperateModal
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
        <NEmpty v-else :description="$t('page.hrm.attendance.cardReplacement.noRecords')" />
      </NModal>
    </NCard>
  </div>
</template>

<style scoped></style>
