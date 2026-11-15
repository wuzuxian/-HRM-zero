<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { AddOvertimeBill, EditOvertimeBill } from '@/service/api';
import { $t } from '@/locales';
import { hrmOvertimeDayTypeOptions } from '@/constants/business';
import { translateOptions } from '@/utils/common';

defineOptions({
  name: 'OvertimeBillOperateModal'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Hrm.HrmOvertimeBill | null;
}

const props = defineProps<Props>();

interface Emits {
  (e: 'submitted'): void;
}

const emit = defineEmits<Emits>();

const visible = defineModel<boolean>('visible', {
  default: false
});

const { formRef, validate, restoreValidation } = useNaiveForm();
const { defaultRequiredRule } = useFormRules();
const { loading: confirmLoading, startLoading: confirmStartLoading, endLoading: confirmEndLoading } = useLoading();

const title = computed(() => {
  const titles: Record<NaiveUI.TableOperateType, string> = {
    add: $t('page.hrm.attendance.overtimeBill.addBill'),
    edit: $t('page.hrm.attendance.overtimeBill.editBill')
  };
  return titles[props.operateType];
});

type Model = Api.Hrm.AddOvertimeBillRequest;

const model: Model = reactive(createDefaultModel());

function createDefaultModel(): Model {
  return {
    employeeUuid: '',
    employeeNo: '',
    employeeName: '',
    orgUuid: '',
    orgName: '',
    overtimeDate: '',
    dayType: '1',
    startTime: '',
    endTime: '',
    reason: ''
  };
}

/** 加班日期（NDatePicker type="date" 单值） */
const overtimeDate = ref<number | null>(null);
/** 起止时间（NDatePicker type="datetimerange" 双值） */
const timeRange = ref<[number, number] | null>(null);

const hoursHint = computed(() => $t('page.hrm.attendance.overtimeBill.hoursPrecision'));

type RuleKey = Extract<keyof Model, 'employeeNo' | 'employeeName'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  employeeNo: defaultRequiredRule,
  employeeName: defaultRequiredRule
};

function dateToBillDate(ts: number) {
  const d = new Date(ts);
  const pad = (n: number) => `${n}`.padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
}

function rangeToBillTime(ts: number) {
  const d = new Date(ts);
  const pad = (n: number) => `${n}`.padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:00`;
}

function handleInitModel() {
  Object.assign(model, createDefaultModel());
  overtimeDate.value = null;
  timeRange.value = null;

  if (props.operateType === 'edit' && props.rowData) {
    Object.assign(model, props.rowData);
    if (props.rowData.overtimeDate) {
      const d = new Date(`${props.rowData.overtimeDate}T00:00:00`).getTime();
      if (!Number.isNaN(d)) overtimeDate.value = d;
    }
    if (props.rowData.startTime && props.rowData.endTime) {
      const start = new Date(props.rowData.startTime.replace(' ', 'T')).getTime();
      const end = new Date(props.rowData.endTime.replace(' ', 'T')).getTime();
      if (!Number.isNaN(start) && !Number.isNaN(end)) {
        timeRange.value = [start, end];
      }
    }
  }
}

function closeModel() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();
  if (overtimeDate.value === null) {
    window.$message?.error($t('page.hrm.attendance.overtimeBill.form.overtimeDateRequired'));
    return;
  }
  if (!timeRange.value || timeRange.value.length !== 2) {
    window.$message?.error($t('page.hrm.attendance.overtimeBill.form.timeRangeRequired'));
    return;
  }
  confirmStartLoading();
  model.overtimeDate = dateToBillDate(overtimeDate.value);
  model.startTime = rangeToBillTime(timeRange.value[0]);
  model.endTime = rangeToBillTime(timeRange.value[1]);

  const request =
    props.operateType === 'add'
      ? AddOvertimeBill({ ...model })
      : EditOvertimeBill({ ...model, uuid: props.rowData?.uuid ?? '' });

  const { error } = await request;
  confirmEndLoading();
  if (!error) {
    window.$message?.success(props.operateType === 'add' ? $t('common.addSuccess') : $t('common.updateSuccess'));
    closeModel();
    emit('submitted');
  }
}

watch(visible, () => {
  if (visible.value) {
    handleInitModel();
    restoreValidation();
  }
});
</script>

<template>
  <NModal v-model:show="visible" :title="title" preset="card" class="w-680px">
    <NScrollbar class="max-h-480px pr-20px">
      <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="100">
        <NGrid responsive="screen" item-responsive>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.attendance.overtimeBill.employeeNo')" path="employeeNo">
            <NInput
              v-model:value="model.employeeNo"
              :placeholder="$t('page.hrm.attendance.overtimeBill.form.employeeNo')"
            />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:8"
            :label="$t('page.hrm.attendance.overtimeBill.employeeName')"
            path="employeeName"
          >
            <NInput
              v-model:value="model.employeeName"
              :placeholder="$t('page.hrm.attendance.overtimeBill.form.employeeName')"
            />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.attendance.overtimeBill.dayType')" path="dayType">
            <NSelect
              v-model:value="model.dayType"
              :options="translateOptions(hrmOvertimeDayTypeOptions)"
              :placeholder="$t('page.hrm.attendance.overtimeBill.form.dayType')"
            />
          </NFormItemGi>
        </NGrid>
        <NGrid responsive="screen" item-responsive>
          <NFormItemGi
            span="24 s:12 m:8"
            :label="$t('page.hrm.attendance.overtimeBill.overtimeDate')"
            path="overtimeDate"
          >
            <NDatePicker v-model:value="overtimeDate" type="date" clearable class="w-full" />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:12" :label="$t('page.hrm.attendance.overtimeBill.timeRange')" path="timeRange">
            <NDatePicker
              v-model:value="timeRange"
              type="datetimerange"
              clearable
              :start-placeholder="$t('page.hrm.attendance.overtimeBill.form.start')"
              :end-placeholder="$t('page.hrm.attendance.overtimeBill.form.end')"
            />
          </NFormItemGi>
        </NGrid>
        <NFormItem :label="$t('page.hrm.attendance.overtimeBill.reason')" path="reason">
          <NInput
            v-model:value="model.reason"
            type="textarea"
            :rows="2"
            :placeholder="$t('page.hrm.attendance.overtimeBill.form.reason')"
          />
        </NFormItem>
        <div class="pb-8px text-12px text-gray-400">{{ hoursHint }}</div>
      </NForm>
    </NScrollbar>
    <template #footer>
      <NSpace justify="end">
        <NButton @click="closeModel">{{ $t('common.cancel') }}</NButton>
        <NButton type="primary" :loading="confirmLoading" @click="handleSubmit">{{ $t('common.confirm') }}</NButton>
      </NSpace>
    </template>
  </NModal>
</template>

<style scoped></style>
