<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { AddTaxProfile, EditTaxProfile, GetEmployeeList } from '@/service/api';
import { $t } from '@/locales';
import { hrmSalaryStatusOptions, hrmTaxDeductTypeOptions, hrmTaxResidentTypeOptions } from '@/constants/business';

defineOptions({
  name: 'TaxProfileOperateModal'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Hrm.TaxProfile | null;
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
    add: $t('page.hrm.salary.tax.profile.addProfile'),
    edit: $t('page.hrm.salary.tax.profile.editProfile')
  };
  return titles[props.operateType];
});

type Model = {
  employeeUuid: string;
  residentType: Api.Hrm.TaxResidentType;
  taxCity: string;
  defaultSpecialExtra: string[];
  status: Api.Hrm.SalarySubjectStatus;
  remark: string;
};

const model: Model = reactive(createDefaultModel());

function createDefaultModel(): Model {
  return {
    employeeUuid: '',
    residentType: 'resident',
    taxCity: '',
    defaultSpecialExtra: [],
    status: '1',
    remark: ''
  };
}

type RuleKey = Extract<keyof Model, 'employeeUuid' | 'residentType'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  employeeUuid: defaultRequiredRule,
  residentType: defaultRequiredRule
};

// ---- employee options (员工档案 from HRM-03) ----
interface EmployeeOption {
  label: string;
  value: string;
}

const employeeOptions = ref<EmployeeOption[]>([]);
const employeeNameByUuid = ref<Map<string, string>>(new Map());

async function loadEmployeeOptions() {
  if (employeeOptions.value.length > 0) return;
  const { data, error } = await GetEmployeeList({ current: 1, size: 500 });
  if (!error && data) {
    const options: EmployeeOption[] = [];
    const map = new Map<string, string>();
    for (const rec of data.records ?? []) {
      options.push({ label: `${rec.employeeNo} ${rec.name}`, value: rec.uuid });
      map.set(rec.uuid, rec.name);
    }
    employeeOptions.value = options;
    employeeNameByUuid.value = map;
  }
}

function handleInitModel() {
  const defaults = createDefaultModel();
  Object.assign(model, defaults);

  if (props.operateType === 'edit' && props.rowData) {
    model.employeeUuid = props.rowData.employeeUuid;
    model.residentType = props.rowData.residentType;
    model.taxCity = props.rowData.taxCity;
    model.defaultSpecialExtra = props.rowData.defaultSpecialExtra
      ? props.rowData.defaultSpecialExtra.split(',').filter(Boolean)
      : [];
    model.status = props.rowData.status ?? '1';
    model.remark = props.rowData.remark;
  }
  loadEmployeeOptions();
}

function closeModel() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();
  confirmStartLoading();

  const payload: Api.Hrm.AddTaxProfileRequest = {
    employeeUuid: model.employeeUuid,
    residentType: model.residentType,
    taxCity: model.taxCity,
    defaultSpecialExtra: model.defaultSpecialExtra.join(','),
    status: model.status,
    remark: model.remark
  };

  const request =
    props.operateType === 'add'
      ? AddTaxProfile(payload)
      : EditTaxProfile({ ...payload, uuid: props.rowData?.uuid ?? '', status: model.status });

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
    <NScrollbar class="h-360px pr-20px">
      <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="120">
        <NFormItem :label="$t('page.hrm.salary.tax.profile.employeeName')" path="employeeUuid">
          <NSelect
            v-model:value="model.employeeUuid"
            :options="employeeOptions"
            :placeholder="$t('page.hrm.salary.tax.profile.form.employeeName')"
            filterable
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.tax.profile.residentType')" path="residentType">
          <NRadioGroup v-model:value="model.residentType">
            <NRadio
              v-for="item in hrmTaxResidentTypeOptions"
              :key="item.value"
              :value="item.value"
              :label="$t(item.label)"
            />
          </NRadioGroup>
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.tax.profile.taxCity')" path="taxCity">
          <NInput v-model:value="model.taxCity" :placeholder="$t('page.hrm.salary.tax.profile.form.taxCity')" />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.tax.profile.defaultSpecialExtra')" path="defaultSpecialExtra">
          <NSelect
            v-model:value="model.defaultSpecialExtra"
            :options="hrmTaxDeductTypeOptions"
            :placeholder="$t('page.hrm.salary.tax.profile.form.defaultSpecialExtra')"
            multiple
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.tax.profile.status')" path="status">
          <NRadioGroup v-model:value="model.status">
            <NRadio
              v-for="item in hrmSalaryStatusOptions"
              :key="item.value"
              :value="item.value"
              :label="$t(item.label)"
            />
          </NRadioGroup>
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.tax.profile.remark')" path="remark">
          <NInput
            v-model:value="model.remark"
            type="textarea"
            :placeholder="$t('page.hrm.salary.tax.profile.form.remark')"
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
