<script setup lang="ts">
import { computed, reactive, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { AddShift, EditShift } from '@/service/api';
import { $t } from '@/locales';
import { hrmAttendanceStatusOptions } from '@/constants/business';

defineOptions({
  name: 'ShiftOperateModal'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Hrm.Shift | null;
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
    add: $t('page.hrm.attendance.shift.addShift'),
    edit: $t('page.hrm.attendance.shift.editShift')
  };
  return titles[props.operateType];
});

type Model = Pick<
  Api.Hrm.AddShiftRequest,
  'shiftName' | 'startTime' | 'endTime' | 'restStart' | 'restEnd' | 'flexibleMinutes' | 'status' | 'remark'
>;

const model: Model = reactive(createDefaultModel());

function createDefaultModel(): Model {
  return {
    shiftName: '',
    startTime: '09:00',
    endTime: '18:00',
    restStart: '',
    restEnd: '',
    flexibleMinutes: 0,
    status: '1',
    remark: ''
  };
}

type RuleKey = Extract<keyof Model, 'shiftName' | 'startTime' | 'endTime'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  shiftName: defaultRequiredRule,
  startTime: defaultRequiredRule,
  endTime: defaultRequiredRule
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
    props.operateType === 'add' ? AddShift({ ...model }) : EditShift({ ...model, uuid: props.rowData?.uuid ?? '' });

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
      <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="100">
        <NFormItem :label="$t('page.hrm.attendance.shift.shiftName')" path="shiftName">
          <NInput v-model:value="model.shiftName" :placeholder="$t('page.hrm.attendance.shift.form.shiftName')" />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.shift.startTime')" path="startTime">
          <NTimePicker
            :value="model.startTime as unknown as number"
            format="HH:mm"
            value-format="HH:mm"
            @update:value="(v: string | number | null) => (model.startTime = String(v ?? ''))"
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.shift.endTime')" path="endTime">
          <NTimePicker
            :value="model.endTime as unknown as number"
            format="HH:mm"
            value-format="HH:mm"
            @update:value="(v: string | number | null) => (model.endTime = String(v ?? ''))"
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.shift.restStart')" path="restStart">
          <NTimePicker
            :value="model.restStart as unknown as number"
            format="HH:mm"
            value-format="HH:mm"
            :placeholder="$t('page.hrm.attendance.shift.form.rest')"
            clearable
            @update:value="(v: string | number | null) => (model.restStart = String(v ?? ''))"
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.shift.restEnd')" path="restEnd">
          <NTimePicker
            :value="model.restEnd as unknown as number"
            format="HH:mm"
            value-format="HH:mm"
            :placeholder="$t('page.hrm.attendance.shift.form.rest')"
            clearable
            @update:value="(v: string | number | null) => (model.restEnd = String(v ?? ''))"
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.shift.flexibleMinutes')" path="flexibleMinutes">
          <NInputNumber v-model:value="model.flexibleMinutes" :min="0" class="w-full" />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.shift.status')" path="status">
          <NRadioGroup v-model:value="model.status">
            <NRadio
              v-for="item in hrmAttendanceStatusOptions"
              :key="item.value"
              :value="item.value"
              :label="$t(item.label)"
            />
          </NRadioGroup>
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.shift.remark')" path="remark">
          <NInput
            v-model:value="model.remark"
            type="textarea"
            :placeholder="$t('page.hrm.attendance.shift.form.remark')"
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
