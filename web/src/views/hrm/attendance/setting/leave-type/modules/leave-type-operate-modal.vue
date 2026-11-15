<script setup lang="ts">
import { computed, reactive, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { AddLeaveType, EditLeaveType } from '@/service/api';
import { $t } from '@/locales';
import { hrmAttendanceStatusOptions } from '@/constants/business';

defineOptions({
  name: 'LeaveTypeOperateModal'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Hrm.LeaveType | null;
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
    add: $t('page.hrm.attendance.leaveType.addLeaveType'),
    edit: $t('page.hrm.attendance.leaveType.editLeaveType')
  };
  return titles[props.operateType];
});

type Model = Pick<
  Api.Hrm.AddLeaveTypeRequest,
  | 'typeCode'
  | 'typeName'
  | 'typeValue'
  | 'balanceEnabled'
  | 'attachmentRequired'
  | 'defaultDays'
  | 'orderNum'
  | 'status'
  | 'remark'
>;

const model: Model = reactive(createDefaultModel());

function createDefaultModel(): Model {
  return {
    typeCode: '',
    typeName: '',
    typeValue: 0,
    balanceEnabled: 0,
    attachmentRequired: 0,
    defaultDays: 0,
    orderNum: 0,
    status: '1',
    remark: ''
  };
}

type RuleKey = Extract<keyof Model, 'typeCode' | 'typeName'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  typeCode: defaultRequiredRule,
  typeName: defaultRequiredRule
};

function handleInitModel() {
  Object.assign(model, createDefaultModel());

  if (props.operateType === 'edit' && props.rowData) {
    Object.assign(model, props.rowData);
  }
}

function closeModel() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();
  confirmStartLoading();

  const request =
    props.operateType === 'add'
      ? AddLeaveType({ ...model })
      : EditLeaveType({ ...model, uuid: props.rowData?.uuid ?? '' });

  const { error } = await request;
  if (!error) {
    window.$message?.success(props.operateType === 'add' ? $t('common.addSuccess') : $t('common.updateSuccess'));
    closeModel();
    emit('submitted');
  }
  confirmEndLoading();
}

watch(visible, () => {
  if (visible.value) {
    handleInitModel();
    restoreValidation();
  }
});
</script>

<template>
  <NModal v-model:show="visible" :title="title" preset="card" class="w-600px">
    <NScrollbar class="h-408px pr-20px">
      <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="120">
        <NFormItem :label="$t('page.hrm.attendance.leaveType.typeName')" path="typeName">
          <NInput v-model:value="model.typeName" :placeholder="$t('page.hrm.attendance.leaveType.form.typeName')" />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.leaveType.typeCode')" path="typeCode">
          <NInput v-model:value="model.typeCode" :placeholder="$t('page.hrm.attendance.leaveType.form.typeCode')" />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.leaveType.typeValue')" path="typeValue">
          <NInputNumber v-model:value="model.typeValue" :min="0" class="w-full" />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.leaveType.balanceEnabled')" path="balanceEnabled">
          <NRadioGroup v-model:value="model.balanceEnabled">
            <NRadio :value="1" :label="$t('page.hrm.attendance.yes')" />
            <NRadio :value="0" :label="$t('page.hrm.attendance.no')" />
          </NRadioGroup>
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.leaveType.attachmentRequired')" path="attachmentRequired">
          <NRadioGroup v-model:value="model.attachmentRequired">
            <NRadio :value="1" :label="$t('page.hrm.attendance.yes')" />
            <NRadio :value="0" :label="$t('page.hrm.attendance.no')" />
          </NRadioGroup>
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.leaveType.defaultDays')" path="defaultDays">
          <NInputNumber v-model:value="model.defaultDays" :min="0" :precision="1" class="w-full" />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.leaveType.orderNum')" path="orderNum">
          <NInputNumber v-model:value="model.orderNum" :min="0" class="w-full" />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.leaveType.status')" path="status">
          <NRadioGroup v-model:value="model.status">
            <NRadio
              v-for="item in hrmAttendanceStatusOptions"
              :key="item.value"
              :value="item.value"
              :label="$t(item.label)"
            />
          </NRadioGroup>
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.leaveType.remark')" path="remark">
          <NInput
            v-model:value="model.remark"
            type="textarea"
            :placeholder="$t('page.hrm.attendance.leaveType.form.remark')"
          />
        </NFormItem>
      </NForm>
    </NScrollbar>
    <template #footer>
      <NSpace :size="16" justify="end">
        <NButton @click="closeModel">{{ $t('common.cancel') }}</NButton>
        <NButton type="primary" :loading="confirmLoading" @click="handleSubmit">{{ $t('common.confirm') }}</NButton>
      </NSpace>
    </template>
  </NModal>
</template>

<style scoped></style>
