<script setup lang="tsx">
import { ref } from 'vue';
import { NButton, NPopconfirm, NTag } from 'naive-ui';
import { ApproveBill, GetSalaryBatchList, GetSalaryConfirmList, RejectBill, SubmitSalaryConfirm } from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { useTable } from '@/hooks/common/table';

defineOptions({
  name: 'SalaryConfirmList'
});

const appStore = useAppStore();

// 批次选择器（提交薪资确认审批用）：已试算且未确认的批次（HRM-37/38 契约口径）。
// 下拉项从批次列表取（分页拉取一页，演示规模足够；后端二次守卫）。
const batchOptions = ref<CommonType.Option<string>[]>([]);
async function loadBatchOptions() {
  const { data, error } = await GetSalaryBatchList({
    current: 1,
    size: 100,
    batchName: null,
    subjectUuid: null,
    batchType: null,
    trialStatus: '1'
  });
  if (!error && data) {
    batchOptions.value = data.records
      .filter(r => r.confirmStatus !== '1')
      .map(r => ({
        label: `${r.batchNo} ${r.batchName}（${r.subjectName} ${r.batchYear}-${String(r.batchMonth).padStart(2, '0')} 实发 ${r.netTotal}）`,
        value: r.uuid
      }));
  }
}
loadBatchOptions();

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
  apiFn: GetSalaryConfirmList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    batchUuid: null,
    status: null,
    mine: 0
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
      title: $t('page.hrm.salary.confirm.billNo'),
      align: 'center',
      width: 150
    },
    {
      key: 'batchNo',
      title: $t('page.hrm.salary.confirm.batch'),
      align: 'center',
      width: 170,
      render: row => `${row.batchNo} ${row.batchName}`
    },
    {
      key: 'subjectName',
      title: $t('page.hrm.salary.confirm.subject'),
      align: 'center',
      width: 120
    },
    {
      key: 'batchYear',
      title: $t('page.hrm.salary.confirm.period'),
      align: 'center',
      width: 100,
      render: row => `${row.batchYear}-${String(row.batchMonth).padStart(2, '0')}`
    },
    {
      key: 'headcount',
      title: $t('page.hrm.salary.confirm.headcount'),
      align: 'center',
      width: 80
    },
    {
      key: 'netTotal',
      title: $t('page.hrm.salary.confirm.netTotal'),
      align: 'center',
      width: 120
    },
    {
      key: 'creatorName',
      title: $t('page.hrm.salary.confirm.creator'),
      align: 'center',
      width: 100,
      render: row => row.creatorName || row.submitBy
    },
    {
      key: 'status',
      title: $t('page.hrm.salary.confirm.status'),
      align: 'center',
      width: 100,
      render: row => {
        const label = billStatusLabel(row.status);
        if (!label) {
          return null;
        }
        return <NTag type={billStatusTagType(row.status)}>{$t(label)}</NTag>;
      }
    },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      width: 200,
      render: row => {
        const isApproving = row.status === 'approving';
        return (
          <div class="flex-center gap-8px">
            <NPopconfirm onPositiveClick={() => handleApprove(row)}>
              {{
                default: () => $t('page.hrm.salary.confirm.approveConfirm'),
                trigger: () => (
                  <NButton v-auth={['hrm:bill:approve']} type="primary" ghost size="small" disabled={!isApproving}>
                    {$t('page.hrm.salary.confirm.approve')}
                  </NButton>
                )
              }}
            </NPopconfirm>
            <NPopconfirm onPositiveClick={() => handleReject(row)}>
              {{
                default: () => $t('page.hrm.salary.confirm.rejectConfirm'),
                trigger: () => (
                  <NButton v-auth={['hrm:bill:reject']} type="error" ghost size="small" disabled={!isApproving}>
                    {$t('page.hrm.salary.confirm.reject')}
                  </NButton>
                )
              }}
            </NPopconfirm>
          </div>
        );
      }
    }
  ]
});

function billStatusLabel(status: Api.Hrm.BillStatus): App.I18n.I18nKey | '' {
  const record: Record<Api.Hrm.BillStatus, App.I18n.I18nKey> = {
    unsubmitted: 'page.hrm.salary.billStatus.unsubmitted',
    approving: 'page.hrm.salary.billStatus.approving',
    approved: 'page.hrm.salary.billStatus.approved',
    rejected: 'page.hrm.salary.billStatus.rejected',
    revoked: 'page.hrm.salary.billStatus.revoked'
  };
  return record[status] ?? '';
}

function billStatusTagType(status: Api.Hrm.BillStatus): 'default' | 'info' | 'success' | 'error' | 'warning' {
  const record: Record<Api.Hrm.BillStatus, 'default' | 'info' | 'success' | 'error' | 'warning'> = {
    unsubmitted: 'default',
    approving: 'info',
    approved: 'success',
    rejected: 'error',
    revoked: 'warning'
  };
  return record[status] ?? 'default';
}

// 提交薪资确认审批（HR209）：选已试算未确认批次 → SubmitSalaryConfirm
const submitVisible = ref(false);
const submitBatchUuid = ref<string | null>(null);
const submitLoading = ref(false);

function handleSubmit() {
  submitBatchUuid.value = null;
  loadBatchOptions();
  submitVisible.value = true;
}

async function confirmSubmit() {
  if (!submitBatchUuid.value) {
    window.$message?.error($t('page.hrm.salary.confirm.batchRequired'));
    return;
  }
  submitLoading.value = true;
  const { error } = await SubmitSalaryConfirm({ batchUuid: submitBatchUuid.value });
  submitLoading.value = false;
  if (!error) {
    window.$message?.success($t('page.hrm.salary.confirm.submitSuccess'));
    submitVisible.value = false;
    getData();
  }
}

async function handleApprove(row: Api.Hrm.SalaryConfirmBill) {
  const { error } = await ApproveBill({ billType: '209', billUuid: row.uuid });
  if (!error) {
    window.$message?.success($t('page.hrm.salary.confirm.approveSuccess'));
    getData();
  }
}

async function handleReject(row: Api.Hrm.SalaryConfirmBill) {
  const { error } = await RejectBill({ billType: '209', billUuid: row.uuid });
  if (!error) {
    window.$message?.success($t('page.hrm.salary.confirm.rejectSuccess'));
    getData();
  }
}
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <NCard :bordered="false" size="small" class="card-wrapper">
      <NForm label-placement="left" :label-width="80">
        <NGrid responsive="screen" item-responsive>
          <NFormItemGi
            span="24 s:12 m:6"
            :label="$t('page.hrm.salary.confirm.batch')"
            path="batchUuid"
            class="pr-24px"
          >
            <NSelect
              v-model:value="searchParams.batchUuid"
              :options="batchOptions"
              :placeholder="$t('page.hrm.salary.confirm.batchFilter')"
              clearable
              filterable
            />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:6" :label="$t('page.hrm.salary.confirm.status')" path="status" class="pr-24px">
            <NSelect
              v-model:value="searchParams.status"
              :placeholder="$t('page.hrm.salary.confirm.statusFilter')"
              :options="[
                { label: $t('page.hrm.salary.billStatus.unsubmitted'), value: 'unsubmitted' },
                { label: $t('page.hrm.salary.billStatus.approving'), value: 'approving' },
                { label: $t('page.hrm.salary.billStatus.approved'), value: 'approved' },
                { label: $t('page.hrm.salary.billStatus.rejected'), value: 'rejected' },
                { label: $t('page.hrm.salary.billStatus.revoked'), value: 'revoked' }
              ]"
              clearable
            />
          </NFormItemGi>
          <NFormItemGi span="24 m:12" class="pr-24px">
            <NSpace class="w-full" justify="end">
              <NButton @click="resetSearchParams">
                <template #icon>
                  <icon-ic-round-refresh class="text-icon" />
                </template>
                {{ $t('common.reset') }}
              </NButton>
              <NButton type="primary" ghost @click="() => getDataByPage()">
                <template #icon>
                  <icon-ic-round-search class="text-icon" />
                </template>
                {{ $t('common.search') }}
              </NButton>
            </NSpace>
          </NFormItemGi>
        </NGrid>
      </NForm>
    </NCard>
    <NCard :title="$t('page.hrm.salary.confirm.title')" :bordered="false" size="small" class="sm:flex-1-hidden card-wrapper">
      <template #header-extra>
        <TableHeaderOperation v-model:columns="columnChecks" :loading="loading" @refresh="getData">
          <NButton v-auth="['hrm:bill:submit']" size="small" type="primary" @click="handleSubmit">
            <template #icon>
              <icon-ic-round-plus class="text-icon" />
            </template>
            {{ $t('page.hrm.salary.confirm.submit') }}
          </NButton>
        </TableHeaderOperation>
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
      <NModal v-model:show="submitVisible" preset="card" :title="$t('page.hrm.salary.confirm.submit')" class="w-560px">
        <NForm label-placement="left" :label-width="80">
          <NFormItem :label="$t('page.hrm.salary.confirm.batch')" required>
            <NSelect
              v-model:value="submitBatchUuid"
              :options="batchOptions"
              :placeholder="$t('page.hrm.salary.confirm.batchPlaceholder')"
              filterable
            />
          </NFormItem>
          <NSpace justify="end">
            <NButton @click="submitVisible = false">{{ $t('common.cancel') }}</NButton>
            <NButton type="primary" :loading="submitLoading" @click="confirmSubmit">
              {{ $t('common.confirm') }}
            </NButton>
          </NSpace>
        </NForm>
      </NModal>
    </NCard>
  </div>
</template>

<style scoped></style>
