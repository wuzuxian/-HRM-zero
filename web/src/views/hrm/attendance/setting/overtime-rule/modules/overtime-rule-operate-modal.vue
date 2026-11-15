<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { AddOvertimeRule, EditOvertimeRule } from '@/service/api';
import { $t } from '@/locales';
import { hrmAttendanceStatusOptions } from '@/constants/business';

defineOptions({
  name: 'OvertimeRuleOperateModal'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Hrm.OvertimeRule | null;
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
    add: $t('page.hrm.attendance.overtimeRule.addRule'),
    edit: $t('page.hrm.attendance.overtimeRule.editRule')
  };
  return titles[props.operateType];
});

type Model = Pick<
  Api.Hrm.AddOvertimeRuleRequest,
  | 'ruleName'
  | 'workday'
  | 'restday'
  | 'holiday'
  | 'minUnitMinutes'
  | 'monthlyLimitHours'
  | 'groupUuids'
  | 'status'
  | 'remark'
>;

const model: Model = reactive(createDefaultModel());

function createDefaultPolicy(): Api.Hrm.DayOvertimePolicy {
  return { approvalRequired: 1, punchRequired: 1, compMultiplier: 1 };
}

function createDefaultModel(): Model {
  return {
    ruleName: '',
    workday: createDefaultPolicy(),
    restday: { approvalRequired: 1, punchRequired: 0, compMultiplier: 1.5 },
    holiday: { approvalRequired: 1, punchRequired: 0, compMultiplier: 3 },
    minUnitMinutes: 30,
    monthlyLimitHours: 0,
    groupUuids: [],
    status: '1',
    remark: ''
  };
}

type RuleKey = Extract<keyof Model, 'ruleName'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  ruleName: defaultRequiredRule
};

const minUnitOptions = [
  { value: 30, label: $t('page.hrm.attendance.overtimeRule.minUnitHalfHour') },
  { value: 60, label: $t('page.hrm.attendance.overtimeRule.minUnitHour') }
];

/** 三种当天类型的策略 tab（工作日/休息日/法定节假日） */
const dayTabs: Array<{ key: 'workday' | 'restday' | 'holiday'; labelKey: App.I18n.I18nKey }> = [
  { key: 'workday', labelKey: 'page.hrm.attendance.overtimeRule.workday' },
  { key: 'restday', labelKey: 'page.hrm.attendance.overtimeRule.restday' },
  { key: 'holiday', labelKey: 'page.hrm.attendance.overtimeRule.holiday' }
];
const activeDay = ref<'workday' | 'restday' | 'holiday'>('workday');
const activePolicy = computed<Api.Hrm.DayOvertimePolicy>(() => {
  const key: keyof Model = activeDay.value;
  return model[key];
});

const approvalOptions = computed(() => [
  { value: 1, label: $t('page.hrm.attendance.overtimeRule.needApproval') },
  { value: 0, label: $t('page.hrm.attendance.overtimeRule.noApproval') }
]);
const punchOptions = computed(() => [
  { value: 1, label: $t('page.hrm.attendance.overtimeRule.needPunch') },
  { value: 0, label: $t('page.hrm.attendance.overtimeRule.noPunch') }
]);

function handleInitModel() {
  Object.assign(model, createDefaultModel());

  if (props.operateType === 'edit' && props.rowData) {
    Object.assign(model, {
      ...props.rowData,
      workday: { ...props.rowData.workday },
      restday: { ...props.rowData.restday },
      holiday: { ...props.rowData.holiday }
    });
  }
  activeDay.value = 'workday';
}

function closeModel() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();
  confirmStartLoading();

  const request =
    props.operateType === 'add'
      ? AddOvertimeRule({ ...model })
      : EditOvertimeRule({ ...model, uuid: props.rowData?.uuid ?? '', status: model.status ?? '1' });

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
  <NModal v-model:show="visible" :title="title" preset="card" class="w-680px">
    <NScrollbar class="h-480px pr-20px">
      <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="110">
        <NFormItem :label="$t('page.hrm.attendance.overtimeRule.ruleName')" path="ruleName">
          <NInput v-model:value="model.ruleName" :placeholder="$t('page.hrm.attendance.overtimeRule.form.ruleName')" />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.overtimeRule.dayPolicy')">
          <NTabs v-model:value="activeDay" type="segment" size="small" class="w-full">
            <NTab v-for="tab in dayTabs" :key="tab.key" :name="tab.key">
              {{ $t(tab.labelKey) }}
            </NTab>
          </NTabs>
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.overtimeRule.approvalRequired')">
          <NRadioGroup v-model:value="activePolicy.approvalRequired">
            <NRadio v-for="item in approvalOptions" :key="item.value" :value="item.value" :label="item.label" />
          </NRadioGroup>
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.overtimeRule.punchRequired')">
          <NRadioGroup v-model:value="activePolicy.punchRequired">
            <NRadio v-for="item in punchOptions" :key="item.value" :value="item.value" :label="item.label" />
          </NRadioGroup>
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.overtimeRule.compMultiplier')">
          <NInputNumber v-model:value="activePolicy.compMultiplier" :min="0.25" :step="0.5" class="w-200px" />
          <span class="ml-8px text-12px text-gray-400">{{ $t('page.hrm.attendance.overtimeRule.multiplierTip') }}</span>
        </NFormItem>
        <NDivider />
        <NFormItem :label="$t('page.hrm.attendance.overtimeRule.minUnitMinutes')">
          <NSelect v-model:value="model.minUnitMinutes" :options="minUnitOptions" class="w-200px" />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.overtimeRule.monthlyLimitHours')">
          <NInputNumber v-model:value="model.monthlyLimitHours" :min="0" :step="1" class="w-200px" />
          <span class="ml-8px text-12px text-gray-400">
            {{ $t('page.hrm.attendance.overtimeRule.monthlyLimitTip') }}
          </span>
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.overtimeRule.status')">
          <NRadioGroup v-model:value="model.status">
            <NRadio
              v-for="item in hrmAttendanceStatusOptions"
              :key="item.value"
              :value="item.value"
              :label="$t(item.label)"
            />
          </NRadioGroup>
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.overtimeRule.remark')">
          <NInput
            v-model:value="model.remark"
            type="textarea"
            :placeholder="$t('page.hrm.attendance.overtimeRule.form.remark')"
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
