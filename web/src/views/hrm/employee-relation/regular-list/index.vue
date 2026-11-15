<script setup lang="tsx">
import { computed, reactive, ref } from 'vue';
import { NButton, NPopconfirm, NTag } from 'naive-ui';
import {
  ApproveBill,
  DeleteRegularization,
  GetApprovalRecordList,
  GetRegularizationList,
  RejectBill,
  RevokeBill,
  SubmitBill
} from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { hrmBillStatusRecord, hrmEmployeePersonStatusRecord } from '@/constants/business';
import { useTable } from '@/hooks/common/table';
import RegularizationOperateModal from './modules/regularization-operate-modal.vue';
import RegularizationSearch from './modules/regularization-search.vue';

const appStore = useAppStore();

type LoadingStatus = Record<string, boolean>;
const deleteLoadingStatus = reactive<LoadingStatus>({});
const actionLoadingStatus = reactive<LoadingStatus>({});

/** 单据状态可修改/删除/提交的状态集（未提交/审批不通过/已撤销，演示站规则） */
const EDITABLE_STATUSES: Api.Hrm.BillStatus[] = ['unsubmitted', 'rejected', 'revoked'];

const HR202 = '202';

function isEditable(row: Api.Hrm.HrmRegularization) {
  return row.status !== null && EDITABLE_STATUSES.includes(row.status);
}

function statusTag(row: Api.Hrm.HrmRegularization) {
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

function personStatusText(status: string) {
  const key = (hrmEmployeePersonStatusRecord as Record<string, App.I18n.I18nKey>)[status];
  return key ? $t(key) : status || '-';
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
  apiFn: GetRegularizationList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    billNo: null,
    status: null,
    employeeNo: null
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
      title: $t('page.hrm.employeeRelation.regularization.billNo'),
      align: 'center',
      width: 170
    },
    {
      key: 'status',
      title: $t('page.hrm.employeeRelation.billStatus.title'),
      align: 'center',
      width: 100,
      render: row => statusTag(row)
    },
    {
      key: 'employeeNo',
      title: $t('page.hrm.employeeRelation.regularization.employeeNo'),
      align: 'center',
      width: 100
    },
    {
      key: 'employeeName',
      title: $t('page.hrm.employeeRelation.regularization.employeeName'),
      align: 'center',
      width: 90
    },
    {
      key: 'orgUuid',
      title: $t('page.hrm.employeeRelation.regularization.orgUuid'),
      align: 'center',
      width: 130,
      render: row => row.orgUuid || '-'
    },
    {
      key: 'companyUuid',
      title: $t('page.hrm.employeeRelation.regularization.companyUuid'),
      align: 'center',
      width: 130,
      render: row => row.companyUuid || '-'
    },
    {
      key: 'position',
      title: $t('page.hrm.employeeRelation.regularization.position'),
      align: 'center',
      width: 110,
      render: row => row.position || '-'
    },
    {
      key: 'post',
      title: $t('page.hrm.employeeRelation.regularization.post'),
      align: 'center',
      width: 110,
      render: row => row.post || '-'
    },
    {
      key: 'personStatus',
      title: $t('page.hrm.employeeRelation.regularization.personStatus'),
      align: 'center',
      width: 90,
      render: row => personStatusText(row.personStatus || '')
    },
    {
      key: 'hireDate',
      title: $t('page.hrm.employeeRelation.regularization.hireDate'),
      align: 'center',
      width: 110,
      render: row => row.hireDate || '-'
    },
    {
      key: 'formalDate',
      title: $t('page.hrm.employeeRelation.regularization.formalDate'),
      align: 'center',
      width: 110,
      render: row => row.formalDate || '-'
    },
    {
      key: 'expectedFormalDate',
      title: $t('page.hrm.employeeRelation.regularization.expectedFormalDate'),
      align: 'center',
      width: 120,
      render: row => row.expectedFormalDate || '-'
    },
    {
      key: 'creatorName',
      title: $t('page.hrm.employeeRelation.regularization.creatorName'),
      align: 'center',
      width: 90
    },
    {
      key: 'createTime',
      title: $t('page.hrm.employeeRelation.regularization.createTime'),
      align: 'center',
      width: 160
    },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      width: 300,
      render: row => renderOperate(row)
    }
  ]
});

function renderOperate(row: Api.Hrm.HrmRegularization) {
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
            default: () => $t('page.hrm.employeeRelation.regularization.confirmSubmit'),
            trigger: () => (
              <NButton loading={actionLoadingStatus[row.uuid]} type="info" ghost size="small">
                {$t('page.hrm.employeeRelation.regularization.submit')}
              </NButton>
            )
          }}
        </NPopconfirm>
      )}
      {actions.has('approve') && (
        <NPopconfirm onPositiveClick={() => handleBillAction(row, ApproveBill)}>
          {{
            default: () => $t('page.hrm.employeeRelation.regularization.confirmApprove'),
            trigger: () => (
              <NButton loading={actionLoadingStatus[row.uuid]} type="success" ghost size="small">
                {$t('page.hrm.employeeRelation.regularization.approve')}
              </NButton>
            )
          }}
        </NPopconfirm>
      )}
      {actions.has('reject') && (
        <NPopconfirm onPositiveClick={() => handleBillAction(row, RejectBill)}>
          {{
            default: () => $t('page.hrm.employeeRelation.regularization.confirmReject'),
            trigger: () => (
              <NButton loading={actionLoadingStatus[row.uuid]} type="error" ghost size="small">
                {$t('page.hrm.employeeRelation.regularization.reject')}
              </NButton>
            )
          }}
        </NPopconfirm>
      )}
      {actions.has('revoke') && (
        <NPopconfirm onPositiveClick={() => handleBillAction(row, RevokeBill)}>
          {{
            default: () => $t('page.hrm.employeeRelation.regularization.confirmRevoke'),
            trigger: () => (
              <NButton loading={actionLoadingStatus[row.uuid]} type="warning" ghost size="small">
                {$t('page.hrm.employeeRelation.regularization.revoke')}
              </NButton>
            )
          }}
        </NPopconfirm>
      )}
      {actions.has('records') && (
        <NButton size="small" onClick={() => handleShowRecords(row)}>
          {$t('page.hrm.employeeRelation.regularization.records')}
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
const editingData = ref<Api.Hrm.HrmRegularization | null>(null);

function handleAdd() {
  operateType.value = 'add';
  editingData.value = null;
  operateVisible.value = true;
}

function handleEdit(row: Api.Hrm.HrmRegularization) {
  operateType.value = 'edit';
  editingData.value = { ...row };
  operateVisible.value = true;
}

function handleSubmitted() {
  getDataByPage();
}

type BillActionFn = (req: Api.Hrm.BillActionRequest) => Promise<{ data: unknown; error: unknown }>;

async function handleBillAction(row: Api.Hrm.HrmRegularization, action: BillActionFn) {
  actionLoadingStatus[row.uuid] = true;
  const { error } = await action({ billType: HR202, billUuid: row.uuid });
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
  const { error } = await DeleteRegularization(uuids);

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
  const { error } = await DeleteRegularization(uuids);
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

async function handleShowRecords(row: Api.Hrm.HrmRegularization) {
  recordsBillNo.value = row.billNo;
  const { data: recordResp, error } = await GetApprovalRecordList({ billType: HR202, billUuid: row.uuid });
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
    submit: 'page.hrm.employeeRelation.regularization.submit',
    approve: 'page.hrm.employeeRelation.regularization.approve',
    reject: 'page.hrm.employeeRelation.regularization.reject',
    revoke: 'page.hrm.employeeRelation.regularization.revoke'
  };
  const key = keys[action];
  return key ? $t(key as App.I18n.I18nKey) : action;
}
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <RegularizationSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <NCard
      :title="$t('page.hrm.employeeRelation.regularization.title')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <TableHeaderOperation
          v-model:columns="columnChecks"
          :disabled-delete="!hasCheckedRows"
          :loading="loading"
          add-auth="hrm:regularization:add"
          delete-auth="hrm:regularization:delete"
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
        :scroll-x="2000"
        :loading="loading"
        remote
        :row-key="row => row.uuid"
        :pagination="mobilePagination"
        class="sm:h-full"
      />
      <RegularizationOperateModal
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
        <NEmpty v-else :description="$t('page.hrm.employeeRelation.regularization.noRecords')" />
      </NModal>
    </NCard>
  </div>
</template>

<style scoped></style>
