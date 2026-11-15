<script setup lang="tsx">
import { computed, reactive, ref, watch } from 'vue';
import { NTag } from 'naive-ui';
import { AdjustLeaveBalance, GetLeaveBalanceLogList } from '@/service/api';
import { $t } from '@/locales';
import { useNaiveForm } from '@/hooks/common/form';

defineOptions({
  name: 'LeaveBalanceOperateModal'
});

interface Props {
  /** operate type: adjust | log */
  operateType: 'adjust' | 'log';
  /** edited balance row (adjust) */
  rowData?: Api.Hrm.LeaveBalance | null;
}

const props = defineProps<Props>();

interface Emits {
  (e: 'submitted'): void;
}

const emit = defineEmits<Emits>();

const visible = defineModel<boolean>('visible', { default: false });

function close() {
  visible.value = false;
}

// ---- adjust form ----
const { formRef, validate, restoreValidation } = useNaiveForm();

type AdjustModel = {
  balanceUuid: string;
  days: number | null;
  reason: string;
};

const adjustModel = reactive<AdjustModel>(createDefaultAdjustModel());

function createDefaultAdjustModel(): AdjustModel {
  return {
    balanceUuid: '',
    days: null,
    reason: ''
  };
}

const adjustRules: Record<'days', App.Global.FormRule> = {
  days: {
    validator: (_rule: unknown, value: number | null) => {
      if (value === null || value === 0) {
        return new Error($t('page.hrm.attendance.leaveBalance.adjustDaysRequired'));
      }
      return true;
    }
  }
};

const adjustTitle = computed(() =>
  props.rowData
    ? `${$t('page.hrm.attendance.leaveBalance.adjustTitle')} — ${props.rowData.employeeName} / ${props.rowData.leaveTypeName} (${props.rowData.year})`
    : $t('page.hrm.attendance.leaveBalance.adjustTitle')
);

const adjustSubmitting = ref(false);

async function handleSubmit() {
  await validate();
  if (!props.rowData) return;
  adjustSubmitting.value = true;
  const { error } = await AdjustLeaveBalance({
    balanceUuid: props.rowData.uuid,
    days: adjustModel.days ?? 0,
    reason: adjustModel.reason
  });
  adjustSubmitting.value = false;
  if (error) return;
  window.$message?.success($t('page.hrm.attendance.leaveBalance.adjustSuccess'));
  emit('submitted');
  close();
}

// ---- log modal ----
const logTitle = computed(() =>
  props.rowData
    ? `${$t('page.hrm.attendance.leaveBalance.logTitle')} — ${props.rowData.employeeName} / ${props.rowData.leaveTypeName} (${props.rowData.year})`
    : $t('page.hrm.attendance.leaveBalance.logTitle')
);

interface LogRow {
  uuid: string;
  changeType: string;
  days: number;
  grantedAfter: number;
  usedAfter: number;
  frozenAfter: number;
  reason: string;
  billUuid: string;
  operatorName: string;
  createTime: string;
}

const logRows = ref<LogRow[]>([]);
const logLoading = ref(false);

const changeTypeTag: Record<string, { label: string; type: 'success' | 'warning' | 'info' | 'error' }> = {
  init: { label: 'page.hrm.attendance.leaveBalance.changeTypeInit', type: 'success' },
  adjust: { label: 'page.hrm.attendance.leaveBalance.changeTypeAdjust', type: 'warning' },
  consume: { label: 'page.hrm.attendance.leaveBalance.changeTypeConsume', type: 'error' },
  accrue: { label: 'page.hrm.attendance.leaveBalance.changeTypeAccrue', type: 'info' }
};

const logColumns = computed(() => [
  {
    key: 'changeType',
    title: $t('page.hrm.attendance.leaveBalance.changeType'),
    align: 'center' as const,
    width: 90,
    render: (row: LogRow) => {
      const meta = changeTypeTag[row.changeType];
      if (!meta) return row.changeType;
      return (
        <NTag type={meta.type} size="small">
          {$t(meta.label as App.I18n.I18nKey)}
        </NTag>
      );
    }
  },
  {
    key: 'days',
    title: $t('page.hrm.attendance.leaveBalance.logDays'),
    align: 'center' as const,
    width: 90,
    render: (row: LogRow) => (row.days > 0 ? `+${row.days}` : `${row.days}`)
  },
  {
    key: 'grantedAfter',
    title: $t('page.hrm.attendance.leaveBalance.granted'),
    align: 'center' as const,
    width: 80
  },
  {
    key: 'usedAfter',
    title: $t('page.hrm.attendance.leaveBalance.used'),
    align: 'center' as const,
    width: 80
  },
  {
    key: 'reason',
    title: $t('page.hrm.attendance.leaveBalance.reason'),
    minWidth: 140
  },
  {
    key: 'billUuid',
    title: $t('page.hrm.attendance.leaveBalance.billUuid'),
    width: 130,
    render: (row: LogRow) => (row.billUuid ? row.billUuid.slice(0, 8) : '-')
  },
  {
    key: 'operatorName',
    title: $t('page.hrm.attendance.leaveBalance.operatorName'),
    align: 'center' as const,
    width: 100,
    render: (row: LogRow) => row.operatorName || '-'
  },
  {
    key: 'createTime',
    title: $t('page.hrm.attendance.leaveBalance.logTime'),
    align: 'center' as const,
    width: 170
  }
]);

async function loadLogs() {
  if (!props.rowData) return;
  logLoading.value = true;
  const { data, error } = await GetLeaveBalanceLogList({
    current: 1,
    size: 100,
    balanceUuid: props.rowData.uuid
  });
  logLoading.value = false;
  if (error) return;
  logRows.value = (data?.records ?? []) as unknown as LogRow[];
}

watch(visible, newVal => {
  if (newVal) {
    if (props.operateType === 'log') {
      loadLogs();
    } else {
      Object.assign(adjustModel, createDefaultAdjustModel());
      restoreValidation();
    }
  }
});
</script>

<template>
  <NModal v-model:show="visible" :title="operateType === 'log' ? logTitle : adjustTitle" preset="card" class="w-800px">
    <template v-if="operateType === 'adjust'">
      <NForm ref="formRef" :model="adjustModel" :rules="adjustRules" label-placement="left" :label-width="90">
        <NFormItem :label="$t('page.hrm.attendance.leaveBalance.granted')" path="granted">
          <span>{{ rowData?.granted ?? 0 }}</span>
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.leaveBalance.available')" path="available">
          <span>{{ rowData?.available ?? 0 }}</span>
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.leaveBalance.adjustDays')" path="days">
          <NInputNumber
            v-model:value="adjustModel.days"
            :step="0.5"
            :precision="1"
            class="w-full"
            :placeholder="$t('page.hrm.attendance.leaveBalance.adjustDaysPlaceholder')"
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.leaveBalance.reason')" path="reason">
          <NInput
            v-model:value="adjustModel.reason"
            type="textarea"
            :rows="2"
            :placeholder="$t('page.hrm.attendance.leaveBalance.form.reason')"
          />
        </NFormItem>
      </NForm>
    </template>
    <template v-else>
      <NDataTable
        :columns="logColumns"
        :data="logRows"
        :loading="logLoading"
        size="small"
        :scroll-x="900"
        :max-height="420"
      />
    </template>
    <template #footer>
      <NSpace justify="end">
        <NButton @click="close">{{ $t('common.close') }}</NButton>
        <NButton v-if="operateType === 'adjust'" type="primary" :loading="adjustSubmitting" @click="handleSubmit">
          {{ $t('common.confirm') }}
        </NButton>
      </NSpace>
    </template>
  </NModal>
</template>
