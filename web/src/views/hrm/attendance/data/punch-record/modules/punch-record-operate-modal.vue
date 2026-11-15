<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { AddPunchRecord, EditPunchRecord } from '@/service/api';
import { $t } from '@/locales';
import { useNaiveForm } from '@/hooks/common/form';

defineOptions({
  name: 'PunchRecordOperateModal'
});

interface Props {
  /** operate type: add | edit */
  operateType: 'add' | 'edit';
  /** edited punch row (edit) */
  rowData?: Api.Hrm.Punch | null;
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

const { formRef, validate, restoreValidation } = useNaiveForm();

type Model = {
  employeeUuid: string;
  punchDate: string | null;
  clockInTime: string | null;
  clockInStatus: string;
  clockInLocation: string;
  clockOutTime: string | null;
  clockOutStatus: string;
  clockOutLocation: string;
  overtime: string;
};

function createDefaultModel(): Model {
  return {
    employeeUuid: '',
    punchDate: null,
    clockInTime: null,
    clockInStatus: 'normal',
    clockInLocation: '',
    clockOutTime: null,
    clockOutStatus: 'normal',
    clockOutLocation: '',
    overtime: '0'
  };
}

const model = reactive<Model>(createDefaultModel());

const statusOptions = computed(() => [
  { label: $t('page.hrm.attendance.punchRecord.statusNormal'), value: 'normal' },
  { label: $t('page.hrm.attendance.punchRecord.statusField'), value: 'field' }
]);

const overtimeOptions = computed(() => [
  { label: $t('page.hrm.attendance.punchRecord.overtimeYes'), value: '1' },
  { label: $t('page.hrm.attendance.punchRecord.overtimeNo'), value: '0' }
]);

const title = computed(() =>
  props.operateType === 'add'
    ? $t('page.hrm.attendance.punchRecord.addTitle')
    : $t('page.hrm.attendance.punchRecord.editTitle')
);

const submitting = ref(false);

async function handleSubmit() {
  await validate();
  submitting.value = true;
  if (props.operateType === 'add') {
    const { error } = await AddPunchRecord({
      employeeUuid: model.employeeUuid,
      punchDate: model.punchDate ?? '',
      clockInTime: model.clockInTime ?? undefined,
      clockInStatus: model.clockInTime ? model.clockInStatus : undefined,
      clockInLocation: model.clockInLocation,
      clockOutTime: model.clockOutTime ?? undefined,
      clockOutStatus: model.clockOutTime ? model.clockOutStatus : undefined,
      clockOutLocation: model.clockOutLocation,
      overtime: model.overtime
    });
    submitting.value = false;
    if (error) return;
  } else {
    const { error } = await EditPunchRecord({
      uuid: props.rowData?.uuid ?? '',
      clockInTime: model.clockInTime ?? undefined,
      clockInStatus: model.clockInTime ? model.clockInStatus : undefined,
      clockInLocation: model.clockInLocation,
      clockOutTime: model.clockOutTime ?? undefined,
      clockOutStatus: model.clockOutTime ? model.clockOutStatus : undefined,
      clockOutLocation: model.clockOutLocation,
      overtime: model.overtime
    });
    submitting.value = false;
    if (error) return;
  }
  window.$message?.success($t('common.updateSuccess'));
  emit('submitted');
  close();
}

watch(visible, newVal => {
  if (!newVal) return;
  restoreValidation();
  Object.assign(model, createDefaultModel());
  if (props.operateType === 'edit' && props.rowData) {
    model.clockInTime = props.rowData.clockInTime || null;
    model.clockInStatus = props.rowData.clockInStatus === 'field' ? 'field' : 'normal';
    model.clockInLocation = props.rowData.clockInLocation;
    model.clockOutTime = props.rowData.clockOutTime || null;
    model.clockOutStatus = props.rowData.clockOutStatus === 'field' ? 'field' : 'normal';
    model.clockOutLocation = props.rowData.clockOutLocation;
    model.overtime = props.rowData.overtime === '1' ? '1' : '0';
  }
});
</script>

<template>
  <NModal v-model:show="visible" :title="title" preset="card" class="w-640px">
    <NForm ref="formRef" :model="model" label-placement="left" :label-width="100">
      <template v-if="operateType === 'add'">
        <NFormItem
          :label="$t('page.hrm.attendance.punchRecord.employeeUuid')"
          path="employeeUuid"
          :rule="{ required: true, message: $t('page.hrm.attendance.punchRecord.form.employeeUuid') }"
        >
          <NInput v-model:value="model.employeeUuid" :placeholder="$t('page.hrm.attendance.punchRecord.form.employeeUuid')" />
        </NFormItem>
        <NFormItem
          :label="$t('page.hrm.attendance.punchRecord.punchDate')"
          path="punchDate"
          :rule="{ required: true, message: $t('page.hrm.attendance.punchRecord.form.punchDate') }"
        >
          <NDatePicker
            v-model:formatted-value="model.punchDate"
            type="date"
            :value-format="'yyyy-MM-dd'"
            class="w-full"
          />
        </NFormItem>
      </template>
      <NFormItem :label="$t('page.hrm.attendance.punchRecord.clockInTime')" path="clockInTime">
        <NDatePicker
          v-model:formatted-value="model.clockInTime"
          type="datetime"
          :value-format="'yyyy-MM-dd HH:mm:ss'"
          class="w-full"
          clearable
        />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.attendance.punchRecord.clockInStatus')" path="clockInStatus">
        <NSelect v-model:value="model.clockInStatus" :options="statusOptions" />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.attendance.punchRecord.clockInLocation')" path="clockInLocation">
        <NInput v-model:value="model.clockInLocation" :placeholder="$t('page.hrm.attendance.punchRecord.form.location')" />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.attendance.punchRecord.clockOutTime')" path="clockOutTime">
        <NDatePicker
          v-model:formatted-value="model.clockOutTime"
          type="datetime"
          :value-format="'yyyy-MM-dd HH:mm:ss'"
          class="w-full"
          clearable
        />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.attendance.punchRecord.clockOutStatus')" path="clockOutStatus">
        <NSelect v-model:value="model.clockOutStatus" :options="statusOptions" />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.attendance.punchRecord.clockOutLocation')" path="clockOutLocation">
        <NInput v-model:value="model.clockOutLocation" :placeholder="$t('page.hrm.attendance.punchRecord.form.location')" />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.attendance.punchRecord.overtime')" path="overtime">
        <NRadioGroup v-model:value="model.overtime">
          <NRadio value="0">{{ $t('page.hrm.attendance.punchRecord.overtimeNo') }}</NRadio>
          <NRadio value="1">{{ $t('page.hrm.attendance.punchRecord.overtimeYes') }}</NRadio>
        </NRadioGroup>
      </NFormItem>
    </NForm>
    <template #footer>
      <NSpace justify="end">
        <NButton @click="close">{{ $t('common.cancel') }}</NButton>
        <NButton type="primary" :loading="submitting" @click="handleSubmit">{{ $t('common.confirm') }}</NButton>
      </NSpace>
    </template>
  </NModal>
</template>
