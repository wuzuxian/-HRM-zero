<script setup lang="tsx">
import { ref } from 'vue';
import { NButton, NInput, NPopconfirm, NTag } from 'naive-ui';
import { ApproveBill, GetPerfPeriodList, GetPerfResultApprovalList, RejectBill, SubmitPerfResultApproval } from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { perfPeriodStatusRecord, perfResultApprovalStatusTag } from '@/constants/hrm-performance';
import { useTable } from '@/hooks/common/table';

defineOptions({
  name: 'PerfResultConfirmList'
});

const appStore = useAppStore();

// 批次选择器（提交结果审批用）：仅 '4'结果已发布 待审批的批次（HRM-43 任务链联动后的口径）。
// 下拉项从批次列表取（分页拉取一页，演示规模足够）。
const periodOptions = ref<CommonType.Option<string>[]>([]);
async function loadPeriodOptions() {
  const { data, error } = await GetPerfPeriodList({
    current: 1,
    size: 100,
    name: null,
    status: '4'
  });
  if (!error && data) {
    periodOptions.value = data.records.map(r => ({ label: `${r.periodNo} ${r.name}`, value: r.uuid }));
  }
}
loadPeriodOptions();

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
  apiFn: GetPerfResultApprovalList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    periodUuid: null,
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
      key: 'billNo',
      title: $t('page.hrm.performance.resultConfirm.billNo'),
      align: 'center',
      width: 150
    },
    {
      key: 'periodNo',
      title: $t('page.hrm.performance.resultConfirm.period'),
      align: 'center',
      width: 170,
      render: row => `${row.periodNo} ${row.periodName}`
    },
    {
      key: 'companySnapshot',
      title: $t('page.hrm.performance.resultConfirm.companies'),
      align: 'center',
      minWidth: 120,
      render: row => parseNames(row.companySnapshot).join('、')
    },
    {
      key: 'orgSnapshot',
      title: $t('page.hrm.performance.resultConfirm.orgs'),
      align: 'center',
      minWidth: 140,
      render: row => parseNames(row.orgSnapshot).join('、')
    },
    {
      key: 'participantCount',
      title: $t('page.hrm.performance.resultConfirm.participantCount'),
      align: 'center',
      width: 80
    },
    {
      key: 'avgScore',
      title: $t('page.hrm.performance.resultConfirm.avgScore'),
      align: 'center',
      width: 100
    },
    {
      key: 'totalAmount',
      title: $t('page.hrm.performance.resultConfirm.totalAmount'),
      align: 'center',
      width: 120
    },
    {
      key: 'status',
      title: $t('page.hrm.performance.resultConfirm.status'),
      align: 'center',
      width: 100,
      render: row => {
        const tag = perfResultApprovalStatusTag(row.status);
        if (!row.status || tag === null) {
          return null;
        }
        return <NTag type={tag}>{$t(perfResultApprovalStatusLabel(row.status))}</NTag>;
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
                default: () => $t('page.hrm.performance.resultConfirm.approveConfirm'),
                trigger: () => (
                  <NButton
                    v-auth={['hrm:bill:approve']}
                    type="primary"
                    ghost
                    size="small"
                    disabled={!isApproving}
                  >
                    {$t('page.hrm.performance.resultConfirm.approve')}
                  </NButton>
                )
              }}
            </NPopconfirm>
            <NPopconfirm onPositiveClick={() => handleReject(row)}>
              {{
                default: () => $t('page.hrm.performance.resultConfirm.rejectConfirm'),
                trigger: () => (
                  <NButton
                    v-auth={['hrm:bill:reject']}
                    type="error"
                    ghost
                    size="small"
                    disabled={!isApproving}
                  >
                    {$t('page.hrm.performance.resultConfirm.reject')}
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

function parseNames(raw: string): string[] {
  try {
    const arr = JSON.parse(raw) as string[];
    return Array.isArray(arr) ? arr : [];
  } catch {
    return [];
  }
}

function perfResultApprovalStatusLabel(status: Api.Hrm.BillStatus): App.I18n.I18nKey {
  const record: Record<Api.Hrm.BillStatus, App.I18n.I18nKey> = {
    unsubmitted: 'page.hrm.performance.resultApprovalStatus.unsubmitted',
    approving: 'page.hrm.performance.resultApprovalStatus.approving',
    approved: 'page.hrm.performance.resultApprovalStatus.approved',
    rejected: 'page.hrm.performance.resultApprovalStatus.rejected',
    revoked: 'page.hrm.performance.resultApprovalStatus.revoked'
  };
  return record[status];
}

// 提交结果审批（HR211）：选批次 → SubmitPerfResultApproval（批次须 结果已发布）
const submitVisible = ref(false);
const submitPeriodUuid = ref<string | null>(null);
const submitLoading = ref(false);

function handleSubmit() {
  submitPeriodUuid.value = null;
  loadPeriodOptions();
  submitVisible.value = true;
}

async function confirmSubmit() {
  if (!submitPeriodUuid.value) {
    window.$message?.error($t('page.hrm.performance.resultConfirm.periodRequired'));
    return;
  }
  submitLoading.value = true;
  const { error } = await SubmitPerfResultApproval({ periodUuid: submitPeriodUuid.value });
  submitLoading.value = false;
  if (!error) {
    window.$message?.success($t('page.hrm.performance.resultConfirm.submitSuccess'));
    submitVisible.value = false;
    getData();
  }
}

async function handleApprove(row: Api.Hrm.Performance.PerfResultApproval) {
  const { error } = await ApproveBill({ billType: '211', billUuid: row.uuid });
  if (!error) {
    window.$message?.success($t('page.hrm.performance.resultConfirm.approveSuccess'));
    getData();
  }
}

async function handleReject(row: Api.Hrm.Performance.PerfResultApproval) {
  const { error } = await RejectBill({ billType: '211', billUuid: row.uuid });
  if (!error) {
    window.$message?.success($t('page.hrm.performance.resultConfirm.rejectSuccess'));
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
            :label="$t('page.hrm.performance.resultConfirm.period')"
            path="periodUuid"
            class="pr-24px"
          >
            <NInput
              v-model:value="searchParams.periodUuid"
              :placeholder="$t('page.hrm.performance.resultConfirm.periodFilter')"
              clearable
            />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:6"
            :label="$t('page.hrm.performance.resultConfirm.status')"
            path="status"
            class="pr-24px"
          >
            <NSelect
              v-model:value="searchParams.status"
              :placeholder="$t('page.hrm.performance.resultConfirm.statusFilter')"
              :options="[
                { label: $t('page.hrm.performance.resultApprovalStatus.unsubmitted'), value: 'unsubmitted' },
                { label: $t('page.hrm.performance.resultApprovalStatus.approving'), value: 'approving' },
                { label: $t('page.hrm.performance.resultApprovalStatus.approved'), value: 'approved' },
                { label: $t('page.hrm.performance.resultApprovalStatus.rejected'), value: 'rejected' },
                { label: $t('page.hrm.performance.resultApprovalStatus.revoked'), value: 'revoked' }
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
    <NCard
      :title="$t('page.hrm.performance.resultConfirm.title')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <TableHeaderOperation v-model:columns="columnChecks" :loading="loading" @refresh="getData">
          <NButton v-auth="['hrm:bill:submit']" size="small" type="primary" @click="handleSubmit">
            <template #icon>
              <icon-ic-round-plus class="text-icon" />
            </template>
            {{ $t('page.hrm.performance.resultConfirm.submit') }}
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
      <NModal
        v-model:show="submitVisible"
        preset="card"
        :title="$t('page.hrm.performance.resultConfirm.submit')"
        class="w-500px"
      >
        <NForm label-placement="left" :label-width="80">
          <NFormItem :label="$t('page.hrm.performance.resultConfirm.period')" required>
            <NSelect
              v-model:value="submitPeriodUuid"
              :options="periodOptions"
              :placeholder="$t('page.hrm.performance.resultConfirm.periodPlaceholder')"
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
