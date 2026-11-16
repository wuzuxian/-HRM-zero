<script setup lang="tsx">
import { ref } from 'vue';
import { NButton, NTag } from 'naive-ui';
import { GetSalaryBatchList, GetSalaryPaymentList, RegisterSalaryPayment } from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { useTable } from '@/hooks/common/table';

defineOptions({
  name: 'SalaryPaymentList'
});

const appStore = useAppStore();

const payModeTagType: Record<string, 'info' | 'warning' | 'default'> = {
  bank: 'info',
  cash: 'warning',
  other: 'default'
};

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
  apiFn: GetSalaryPaymentList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    batchUuid: null,
    payMode: null
  },
  columns: () => [
    {
      key: 'index',
      title: $t('common.index'),
      align: 'center',
      width: 64
    },
    {
      key: 'batchNo',
      title: $t('page.hrm.salary.payment.batch'),
      align: 'center',
      width: 170,
      render: row => `${row.batchNo} ${row.batchName}`
    },
    {
      key: 'subjectName',
      title: $t('page.hrm.salary.payment.subject'),
      align: 'center',
      width: 120
    },
    {
      key: 'payAmount',
      title: $t('page.hrm.salary.payment.payAmount'),
      align: 'center',
      width: 120
    },
    {
      key: 'payMode',
      title: $t('page.hrm.salary.payment.payMode'),
      align: 'center',
      width: 110,
      render: row => {
        const key = `page.hrm.salary.payment.payModeEnum.${row.payMode}` as App.I18n.I18nKey;
        return <NTag type={payModeTagType[row.payMode] ?? 'default'}>{$t(key)}</NTag>;
      }
    },
    {
      key: 'receiptNo',
      title: $t('page.hrm.salary.payment.receiptNo'),
      align: 'center',
      width: 150
    },
    {
      key: 'arriveTime',
      title: $t('page.hrm.salary.payment.arriveTime'),
      align: 'center',
      width: 170
    },
    {
      key: 'registerName',
      title: $t('page.hrm.salary.payment.registerName'),
      align: 'center',
      width: 100
    },
    {
      key: 'registerTime',
      title: $t('page.hrm.salary.payment.registerTime'),
      align: 'center',
      width: 170
    }
  ]
});

// 发薪登记（HR210）：选已确认未发薪的批次（confirm_status='1'）→ RegisterSalaryPayment。
// 发薪金额默认=批次实发合计（可改）。
const registerVisible = ref(false);
const registerBatchUuid = ref<string | null>(null);
const registerForm = ref({
  payAmount: null as number | null,
  payMode: 'bank' as string,
  receiptNo: '',
  arriveTime: null as number | null
});
const registerLoading = ref(false);

const payModeOptions = [
  { label: $t('page.hrm.salary.payment.payModeEnum.bank'), value: 'bank' },
  { label: $t('page.hrm.salary.payment.payModeEnum.cash'), value: 'cash' },
  { label: $t('page.hrm.salary.payment.payModeEnum.other'), value: 'other' }
];

// 已确认未发薪的批次下拉项
const confirmBatchOptions = ref<CommonType.Option<string>[]>([]);
async function loadConfirmBatchOptions() {
  const { data, error } = await GetSalaryBatchList({
    current: 1,
    size: 100,
    batchName: null,
    subjectUuid: null,
    batchType: null,
    trialStatus: '1'
  });
  if (!error && data) {
    confirmBatchOptions.value = data.records
      .filter(r => r.confirmStatus === '1' && r.payStatus !== '1')
      .map(r => ({
        label: `${r.batchNo} ${r.batchName}（${r.subjectName} 实发 ${r.netTotal}）`,
        value: r.uuid,
        netTotal: r.netTotal
      }));
  }
}

function handleRegister() {
  registerBatchUuid.value = null;
  registerForm.value = { payAmount: null, payMode: 'bank', receiptNo: '', arriveTime: null };
  loadConfirmBatchOptions();
  registerVisible.value = true;
}

function onBatchPick(value: string | null) {
  const opt = confirmBatchOptions.value.find(o => o.value === value) as
    | (CommonType.Option<string> & { netTotal?: number })
    | undefined;
  registerForm.value.payAmount = opt?.netTotal ?? null;
}

async function confirmRegister() {
  if (!registerBatchUuid.value) {
    window.$message?.error($t('page.hrm.salary.payment.batchRequired'));
    return;
  }
  registerLoading.value = true;
  const { error } = await RegisterSalaryPayment({
    batchUuid: registerBatchUuid.value,
    payAmount: registerForm.value.payAmount ?? 0,
    payMode: registerForm.value.payMode,
    receiptNo: registerForm.value.receiptNo || undefined,
    arriveTime: registerForm.value.arriveTime
      ? new Date(registerForm.value.arriveTime).toLocaleString('sv-SE').replace('T', ' ')
      : undefined
  });
  registerLoading.value = false;
  if (!error) {
    window.$message?.success($t('page.hrm.salary.payment.registerSuccess'));
    registerVisible.value = false;
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
            :label="$t('page.hrm.salary.payment.batch')"
            path="batchUuid"
            class="pr-24px"
          >
            <NSelect
              v-model:value="searchParams.batchUuid"
              :options="confirmBatchOptions"
              :placeholder="$t('page.hrm.salary.payment.batchFilter')"
              clearable
              filterable
            />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:6" :label="$t('page.hrm.salary.payment.payMode')" path="payMode" class="pr-24px">
            <NSelect
              v-model:value="searchParams.payMode"
              :options="payModeOptions"
              :placeholder="$t('page.hrm.salary.payment.payModeFilter')"
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
    <NCard :title="$t('page.hrm.salary.payment.title')" :bordered="false" size="small" class="sm:flex-1-hidden card-wrapper">
      <template #header-extra>
        <TableHeaderOperation v-model:columns="columnChecks" :loading="loading" @refresh="getData">
          <NButton v-auth="['hrm:salary_payment:register']" size="small" type="primary" @click="handleRegister">
            <template #icon>
              <icon-ic-round-plus class="text-icon" />
            </template>
            {{ $t('page.hrm.salary.payment.register') }}
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
        v-model:show="registerVisible"
        preset="card"
        :title="$t('page.hrm.salary.payment.register')"
        class="w-560px"
      >
        <NForm label-placement="left" :label-width="90">
          <NFormItem :label="$t('page.hrm.salary.payment.batch')" required>
            <NSelect
              v-model:value="registerBatchUuid"
              :options="confirmBatchOptions"
              :placeholder="$t('page.hrm.salary.payment.batchPlaceholder')"
              filterable
              @update:value="onBatchPick"
            />
          </NFormItem>
          <NFormItem :label="$t('page.hrm.salary.payment.payAmount')">
            <NInputNumber
              v-model:value="registerForm.payAmount"
              :placeholder="$t('page.hrm.salary.payment.payAmountPlaceholder')"
              :min="0"
              :precision="2"
              class="w-full"
            />
          </NFormItem>
          <NFormItem :label="$t('page.hrm.salary.payment.payMode')">
            <NSelect v-model:value="registerForm.payMode" :options="payModeOptions" />
          </NFormItem>
          <NFormItem :label="$t('page.hrm.salary.payment.receiptNo')">
            <NInput
              v-model:value="registerForm.receiptNo"
              :placeholder="$t('page.hrm.salary.payment.receiptNoPlaceholder')"
            />
          </NFormItem>
          <NFormItem :label="$t('page.hrm.salary.payment.arriveTime')">
            <NDatePicker
              v-model:value="registerForm.arriveTime"
              type="datetime"
              class="w-full"
              :placeholder="$t('page.hrm.salary.payment.arriveTimePlaceholder')"
            />
          </NFormItem>
          <NSpace justify="end">
            <NButton @click="registerVisible = false">{{ $t('common.cancel') }}</NButton>
            <NButton type="primary" :loading="registerLoading" @click="confirmRegister">
              {{ $t('common.confirm') }}
            </NButton>
          </NSpace>
        </NForm>
      </NModal>
    </NCard>
  </div>
</template>

<style scoped></style>
