<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { AddLeave, EditLeave, GetLeaveTypeList } from '@/service/api';
import { $t } from '@/locales';

defineOptions({
  name: 'LeaveCancelOperateModal'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Hrm.HrmLeave | null;
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
    add: $t('page.hrm.attendance.leaveCancel.addLeave'),
    edit: $t('page.hrm.attendance.leaveCancel.editLeave')
  };
  return titles[props.operateType];
});

// ---- 假期类型 options ----
interface LeaveTypeOption {
  label: string;
  value: string;
}

const leaveTypeOptions = ref<LeaveTypeOption[]>([]);

async function loadLeaveTypes() {
  if (leaveTypeOptions.value.length > 0) return;
  const { data, error } = await GetLeaveTypeList({ current: 1, size: 100 });
  if (!error && data) {
    leaveTypeOptions.value = (data.records ?? []).map(item => ({
      label: item.typeName,
      value: item.typeCode
    }));
  }
}

type Model = Api.Hrm.AddLeaveRequest;

const model: Model = reactive(createDefaultModel());

function createDefaultModel(): Model {
  return {
    employeeUuid: '',
    employeeNo: '',
    employeeName: '',
    orgUuid: '',
    orgName: '',
    leaveTypeCode: 'annual_leave',
    planStartTime: '',
    planEndTime: '',
    reason: ''
  };
}

/** 预计时间区间（NDatePicker type="datetimerange" 双值） */
const planRange = ref<[number, number] | null>(null);

const dayPrecisionHint = computed(() => $t('page.hrm.attendance.leaveCancel.dayPrecision'));

type RuleKey = Extract<keyof Model, 'employeeNo' | 'employeeName' | 'leaveTypeCode'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  employeeNo: defaultRequiredRule,
  employeeName: defaultRequiredRule,
  leaveTypeCode: defaultRequiredRule
};

function handleInitModel() {
  Object.assign(model, createDefaultModel());
  planRange.value = null;

  if (props.operateType === 'edit' && props.rowData) {
    Object.assign(model, props.rowData);
    if (props.rowData.planStartTime && props.rowData.planEndTime) {
      const start = new Date(props.rowData.planStartTime.replace(' ', 'T')).getTime();
      const end = new Date(props.rowData.planEndTime.replace(' ', 'T')).getTime();
      if (!Number.isNaN(start) && !Number.isNaN(end)) {
        planRange.value = [start, end];
      }
    }
  }
  loadLeaveTypes();
}

function rangeToBillTime(ts: number) {
  const d = new Date(ts);
  const pad = (n: number) => `${n}`.padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:00`;
}

function closeModel() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();
  if (!planRange.value || planRange.value.length !== 2) {
    window.$message?.error($t('page.hrm.attendance.leaveCancel.form.planRangeRequired'));
    return;
  }
  confirmStartLoading();
  model.planStartTime = rangeToBillTime(planRange.value[0]);
  model.planEndTime = rangeToBillTime(planRange.value[1]);

  const request =
    props.operateType === 'add'
      ? AddLeave({ ...model })
      : EditLeave({ ...model, uuid: props.rowData?.uuid ?? '' });

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
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.attendance.leaveCancel.employeeNo')" path="employeeNo">
            <NInput
              v-model:value="model.employeeNo"
              :placeholder="$t('page.hrm.attendance.leaveCancel.form.employeeNo')"
            />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.attendance.leaveCancel.employeeName')" path="employeeName">
            <NInput
              v-model:value="model.employeeName"
              :placeholder="$t('page.hrm.attendance.leaveCancel.form.employeeName')"
            />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:8"
            :label="$t('page.hrm.attendance.leaveCancel.leaveTypeName')"
            path="leaveTypeCode"
          >
            <NSelect
              v-model:value="model.leaveTypeCode"
              :options="leaveTypeOptions"
              :placeholder="$t('page.hrm.attendance.leaveCancel.form.leaveTypeCode')"
            />
          </NFormItemGi>
        </NGrid>
        <NGrid responsive="screen" item-responsive>
          <NFormItemGi span="24 s:12 m:12" :label="$t('page.hrm.attendance.leaveCancel.planRange')" path="planRange">
            <NDatePicker
              v-model:value="planRange"
              type="datetimerange"
              clearable
              :start-placeholder="$t('page.hrm.attendance.leaveCancel.form.planStart')"
              :end-placeholder="$t('page.hrm.attendance.leaveCancel.form.planEnd')"
            />
          </NFormItemGi>
        </NGrid>
        <NFormItem :label="$t('page.hrm.attendance.leaveCancel.reason')" path="reason">
          <NInput
            v-model:value="model.reason"
            type="textarea"
            :rows="2"
            :placeholder="$t('page.hrm.attendance.leaveCancel.form.reason')"
          />
        </NFormItem>
        <div class="pb-8px text-12px text-gray-400">{{ dayPrecisionHint }}</div>
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
