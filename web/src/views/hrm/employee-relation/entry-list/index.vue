<script setup lang="tsx">
import { computed, reactive, ref } from 'vue';
import { NButton, NPopconfirm, NTag } from 'naive-ui';
import {
  ApproveBill,
  DeleteEntry,
  GetApprovalRecordList,
  GetEntryDetail,
  GetEntryList,
  RejectBill,
  RevokeBill,
  SubmitBill
} from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { hrmBillStatusRecord } from '@/constants/business';
import { useTable } from '@/hooks/common/table';
import EntryOperateModal from './modules/entry-operate-modal.vue';
import EntrySearch from './modules/entry-search.vue';

const appStore = useAppStore();

type LoadingStatus = Record<string, boolean>;
const deleteLoadingStatus = reactive<LoadingStatus>({});
const actionLoadingStatus = reactive<LoadingStatus>({});

/** 单据状态可修改/删除/提交的状态集（未提交/审批不通过/已撤销，演示站规则） */
const EDITABLE_STATUSES: Api.Hrm.BillStatus[] = ['unsubmitted', 'rejected', 'revoked'];

const HR201 = '201';

function isEditable(row: Api.Hrm.HrmEntryBill) {
  return row.status !== null && EDITABLE_STATUSES.includes(row.status);
}

function statusTag(row: Api.Hrm.HrmEntryBill) {
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
  apiFn: GetEntryList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    billNo: null,
    status: null,
    name: null
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
      title: $t('page.hrm.employeeRelation.entry.billNo'),
      align: 'center',
      width: 170
    },
    {
      key: 'status',
      title: $t('page.hrm.employeeRelation.entry.billStatus.title'),
      align: 'center',
      width: 100,
      render: row => statusTag(row)
    },
    {
      key: 'name',
      title: $t('page.hrm.employeeRelation.entry.name'),
      align: 'center',
      width: 90
    },
    {
      key: 'phone',
      title: $t('page.hrm.employeeRelation.entry.phone'),
      align: 'center',
      width: 120,
      render: row => row.phone || '-'
    },
    {
      key: 'email',
      title: $t('page.hrm.employeeRelation.entry.email'),
      align: 'center',
      width: 150,
      render: row => row.email || '-'
    },
    {
      key: 'position',
      title: $t('page.hrm.employeeRelation.entry.position'),
      align: 'center',
      width: 110,
      render: row => row.position || '-'
    },
    {
      key: 'post',
      title: $t('page.hrm.employeeRelation.entry.post'),
      align: 'center',
      width: 110,
      render: row => row.post || '-'
    },
    {
      key: 'hireDate',
      title: $t('page.hrm.employeeRelation.entry.hireDate'),
      align: 'center',
      width: 110,
      render: row => row.hireDate || '-'
    },
    {
      key: 'expectFormalDate',
      title: $t('page.hrm.employeeRelation.entry.expectFormalDate'),
      align: 'center',
      width: 120,
      render: row => row.expectFormalDate || '-'
    },
    {
      key: 'employeeUuid',
      title: $t('page.hrm.employeeRelation.entry.archived'),
      align: 'center',
      width: 90,
      render: row =>
        row.employeeUuid ? (
          <NTag type="success" size="small">
            {$t('page.hrm.employeeRelation.entry.archivedYes')}
          </NTag>
        ) : (
          <span>{$t('page.hrm.employeeRelation.entry.archivedNo')}</span>
        )
    },
    {
      key: 'creatorName',
      title: $t('page.hrm.employeeRelation.entry.creatorName'),
      align: 'center',
      width: 90
    },
    {
      key: 'createTime',
      title: $t('page.hrm.employeeRelation.entry.createTime'),
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

function renderOperate(row: Api.Hrm.HrmEntryBill) {
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
            default: () => $t('page.hrm.employeeRelation.entry.confirmSubmit'),
            trigger: () => (
              <NButton loading={actionLoadingStatus[row.uuid]} type="info" ghost size="small">
                {$t('page.hrm.employeeRelation.entry.submit')}
              </NButton>
            )
          }}
        </NPopconfirm>
      )}
      {actions.has('approve') && (
        <NPopconfirm onPositiveClick={() => handleBillAction(row, ApproveBill)}>
          {{
            default: () => $t('page.hrm.employeeRelation.entry.confirmApprove'),
            trigger: () => (
              <NButton loading={actionLoadingStatus[row.uuid]} type="success" ghost size="small">
                {$t('page.hrm.employeeRelation.entry.approve')}
              </NButton>
            )
          }}
        </NPopconfirm>
      )}
      {actions.has('reject') && (
        <NPopconfirm onPositiveClick={() => handleBillAction(row, RejectBill)}>
          {{
            default: () => $t('page.hrm.employeeRelation.entry.confirmReject'),
            trigger: () => (
              <NButton loading={actionLoadingStatus[row.uuid]} type="error" ghost size="small">
                {$t('page.hrm.employeeRelation.entry.reject')}
              </NButton>
            )
          }}
        </NPopconfirm>
      )}
      {actions.has('revoke') && (
        <NPopconfirm onPositiveClick={() => handleBillAction(row, RevokeBill)}>
          {{
            default: () => $t('page.hrm.employeeRelation.entry.confirmRevoke'),
            trigger: () => (
              <NButton loading={actionLoadingStatus[row.uuid]} type="warning" ghost size="small">
                {$t('page.hrm.employeeRelation.entry.revoke')}
              </NButton>
            )
          }}
        </NPopconfirm>
      )}
      {actions.has('records') && (
        <NButton size="small" onClick={() => handleShowRecords(row)}>
          {$t('page.hrm.employeeRelation.entry.records')}
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
const editingData = ref<Api.Hrm.HrmEntryBill | null>(null);

function handleAdd() {
  operateType.value = 'add';
  editingData.value = null;
  operateVisible.value = true;
}

async function handleEdit(row: Api.Hrm.HrmEntryBill) {
  // 详情带三子表快照, 编辑时先拉全量
  const { data: detail, error } = await GetEntryDetail(row.uuid);
  if (error) return;
  operateType.value = 'edit';
  editingData.value = detail?.entry ?? row;
  operateVisible.value = true;
}

function handleSubmitted() {
  getDataByPage();
}

type BillActionFn = (req: Api.Hrm.BillActionRequest) => Promise<{ data: unknown; error: unknown }>;

async function handleBillAction(row: Api.Hrm.HrmEntryBill, action: BillActionFn) {
  actionLoadingStatus[row.uuid] = true;
  const { error } = await action({ billType: HR201, billUuid: row.uuid });
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
  const { error } = await DeleteEntry(uuids);

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
  const { error } = await DeleteEntry(uuids);
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

async function handleShowRecords(row: Api.Hrm.HrmEntryBill) {
  recordsBillNo.value = row.billNo;
  const { data: recordResp, error } = await GetApprovalRecordList({ billType: HR201, billUuid: row.uuid });
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
    submit: 'page.hrm.employeeRelation.entry.submit',
    approve: 'page.hrm.employeeRelation.entry.approve',
    reject: 'page.hrm.employeeRelation.entry.reject',
    revoke: 'page.hrm.employeeRelation.entry.revoke'
  };
  const key = keys[action];
  return key ? $t(key as App.I18n.I18nKey) : action;
}
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <EntrySearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <NCard
      :title="$t('page.hrm.employeeRelation.entry.title')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <TableHeaderOperation
          v-model:columns="columnChecks"
          :disabled-delete="!hasCheckedRows"
          :loading="loading"
          add-auth="hrm:entry:add"
          delete-auth="hrm:entry:delete"
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
        :scroll-x="1700"
        :loading="loading"
        remote
        :row-key="row => row.uuid"
        :pagination="mobilePagination"
        class="sm:h-full"
      />
      <EntryOperateModal
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
        <NEmpty v-else :description="$t('page.hrm.employeeRelation.entry.noRecords')" />
      </NModal>
    </NCard>
  </div>
</template>

<style scoped></style>
