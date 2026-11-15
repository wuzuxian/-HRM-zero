<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { AddScheduleCycle, EditScheduleCycle, GetAttendanceGroupList, GetShiftList } from '@/service/api';
import { $t } from '@/locales';
import { hrmAttendanceStatusOptions } from '@/constants/business';

defineOptions({
  name: 'ScheduleCycleOperateModal'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Hrm.ScheduleCycle | null;
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
    add: $t('page.hrm.attendance.scheduleCycle.addCycle'),
    edit: $t('page.hrm.attendance.scheduleCycle.editCycle')
  };
  return titles[props.operateType];
});

type Model = {
  cycleName: string;
  shiftUuids: string[];
  groupUuids: string[];
  status: Api.Hrm.AttendanceStatus;
  remark: string;
};

const model: Model = reactive(createDefaultModel());

function createDefaultModel(): Model {
  return {
    cycleName: '',
    shiftUuids: [],
    groupUuids: [],
    status: '1',
    remark: ''
  };
}

type RuleKey = Extract<keyof Model, 'cycleName' | 'shiftUuids'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  cycleName: defaultRequiredRule,
  shiftUuids: defaultRequiredRule
};

// ---- 班次 options（启用班次）----
const shiftOptions = ref<{ label: string; value: string }[]>([]);

async function loadShiftOptions() {
  const { error, data } = await GetShiftList({ current: 1, size: 500, status: '1' });
  if (!error && data) {
    shiftOptions.value = data.records.map(row => ({
      label: row.shiftName,
      value: row.uuid
    }));
  }
}

// ---- 适用考勤组 options ----
const groupOptions = ref<{ label: string; value: string }[]>([]);

async function loadGroupOptions() {
  const { error, data } = await GetAttendanceGroupList({ current: 1, size: 500, status: '1' });
  if (!error && data) {
    groupOptions.value = data.records.map(row => ({
      label: row.groupName,
      value: row.uuid
    }));
  }
}

/** 周期天数 = 班次序列长度（派生, 不落库） */
const cycleDays = computed(() => model.shiftUuids.length);

function handleInitModel() {
  Object.assign(model, createDefaultModel());

  if (props.operateType === 'edit' && props.rowData) {
    Object.assign(model, {
      ...props.rowData,
      shiftUuids: props.rowData.shiftUuids ? [...props.rowData.shiftUuids] : [],
      groupUuids: props.rowData.groupUuids ? [...props.rowData.groupUuids] : [],
      status: props.rowData.status ?? '1'
    });
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
      ? AddScheduleCycle({ ...model })
      : EditScheduleCycle({ ...model, uuid: props.rowData?.uuid ?? '', status: model.status ?? '1' });

  const { error } = await request;
  if (!error) {
    window.$message?.success(props.operateType === 'add' ? $t('common.addSuccess') : $t('common.updateSuccess'));
    closeModel();
    emit('submitted');
  }
  confirmEndLoading();
}

watch(visible, v => {
  if (v) {
    handleInitModel();
    restoreValidation();
    loadShiftOptions();
    loadGroupOptions();
  }
});
</script>

<template>
  <NModal v-model:show="visible" :title="title" preset="card" class="w-680px">
    <NScrollbar class="h-480px pr-20px">
      <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="110">
        <NFormItem :label="$t('page.hrm.attendance.scheduleCycle.cycleName')" path="cycleName">
          <NInput
            v-model:value="model.cycleName"
            :placeholder="$t('page.hrm.attendance.scheduleCycle.form.cycleName')"
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.scheduleCycle.shiftSequence')" path="shiftUuids">
          <NSelect
            v-model:value="model.shiftUuids"
            multiple
            filterable
            :options="shiftOptions"
            :placeholder="$t('page.hrm.attendance.scheduleCycle.form.shiftSequence')"
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.scheduleCycle.cycleDays')">
          <span class="text-14px font-bold">{{ cycleDays }}</span>
          <span class="ml-8px text-12px text-gray-400">{{ $t('page.hrm.attendance.scheduleCycle.cycleDaysTip') }}</span>
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.scheduleCycle.usedGroups')" path="groupUuids">
          <NSelect
            v-model:value="model.groupUuids"
            multiple
            filterable
            :options="groupOptions"
            :placeholder="$t('page.hrm.attendance.scheduleCycle.form.usedGroups')"
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.scheduleCycle.status')">
          <NRadioGroup v-model:value="model.status">
            <NRadio
              v-for="item in hrmAttendanceStatusOptions"
              :key="item.value"
              :value="item.value"
              :label="$t(item.label)"
            />
          </NRadioGroup>
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.scheduleCycle.remark')">
          <NInput
            v-model:value="model.remark"
            type="textarea"
            :placeholder="$t('page.hrm.attendance.scheduleCycle.form.remark')"
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
