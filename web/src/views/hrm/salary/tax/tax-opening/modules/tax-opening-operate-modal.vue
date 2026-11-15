<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { AddTaxOpening, EditTaxOpening, GetEmployeeList, GetSalarySubjectList } from '@/service/api';
import { $t } from '@/locales';

defineOptions({
  name: 'TaxOpeningOperateModal'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Hrm.TaxOpening | null;
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
    add: $t('page.hrm.salary.tax.opening.addOpening'),
    edit: $t('page.hrm.salary.tax.opening.editOpening')
  };
  return titles[props.operateType];
});

type Model = Api.Hrm.AddTaxOpeningRequest;

const model: Model = reactive(createDefaultModel());

function createDefaultModel(): Model {
  return {
    employeeUuid: '',
    subjectUuid: '',
    taxYear: String(new Date().getFullYear()),
    untilMonth: 12,
    incomeTotal: 0,
    specialDeductionTotal: 0,
    specialExtraTotal: 0,
    exemptTotal: 0,
    otherDeductionTotal: 0,
    taxableIncomeTotal: 0,
    taxPaidTotal: 0,
    remark: ''
  };
}

type RuleKey = Extract<keyof Model, 'employeeUuid' | 'taxYear' | 'untilMonth'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  employeeUuid: defaultRequiredRule,
  taxYear: defaultRequiredRule,
  untilMonth: defaultRequiredRule
};

// ---- options ----
interface EmployeeOption {
  label: string;
  value: string;
}

const employeeOptions = ref<EmployeeOption[]>([]);

async function loadEmployeeOptions() {
  if (employeeOptions.value.length > 0) return;
  const { data, error } = await GetEmployeeList({ current: 1, size: 500 });
  if (!error && data) {
    employeeOptions.value = (data.records ?? []).map(rec => ({
      label: `${rec.employeeNo} ${rec.name}`,
      value: rec.uuid
    }));
  }
}

const subjectOptions = ref<{ label: string; value: string }[]>([]);

async function loadSubjectOptions() {
  if (subjectOptions.value.length > 0) return;
  const { data, error } = await GetSalarySubjectList({ current: 1, size: 200 });
  if (!error && data) {
    subjectOptions.value = (data.records ?? []).map(rec => ({
      label: rec.subjectName,
      value: rec.uuid
    }));
  }
}

const yearOptions = (() => {
  const now = new Date().getFullYear();
  return [now + 1, now, now - 1, now - 2].map(y => ({ label: String(y), value: String(y) }));
})();

const monthOptions = Array.from({ length: 12 }, (_, i) => ({ label: String(i + 1), value: i + 1 }));

function handleInitModel() {
  Object.assign(model, createDefaultModel());

  if (props.operateType === 'edit' && props.rowData) {
    model.employeeUuid = props.rowData.employeeUuid;
    model.subjectUuid = props.rowData.subjectUuid;
    model.taxYear = props.rowData.taxYear;
    model.untilMonth = props.rowData.untilMonth;
    model.incomeTotal = props.rowData.incomeTotal;
    model.specialDeductionTotal = props.rowData.specialDeductionTotal;
    model.specialExtraTotal = props.rowData.specialExtraTotal;
    model.exemptTotal = props.rowData.exemptTotal;
    model.otherDeductionTotal = props.rowData.otherDeductionTotal;
    model.taxableIncomeTotal = props.rowData.taxableIncomeTotal;
    model.taxPaidTotal = props.rowData.taxPaidTotal;
    model.remark = props.rowData.remark;
  }
  loadEmployeeOptions();
  loadSubjectOptions();
}

function closeModel() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();
  confirmStartLoading();

  const request =
    props.operateType === 'add'
      ? AddTaxOpening({ ...model })
      : EditTaxOpening({ ...model, uuid: props.rowData?.uuid ?? '' });

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
    <NScrollbar class="h-420px pr-20px">
      <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="150">
        <NFormItem :label="$t('page.hrm.salary.tax.opening.employeeName')" path="employeeUuid">
          <NSelect
            v-model:value="model.employeeUuid"
            :options="employeeOptions"
            :placeholder="$t('page.hrm.salary.tax.opening.form.employeeName')"
            filterable
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.tax.opening.subjectName')" path="subjectUuid">
          <NSelect
            v-model:value="model.subjectUuid"
            :options="subjectOptions"
            :placeholder="$t('page.hrm.salary.tax.opening.form.subjectName')"
            clearable
            filterable
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.tax.opening.taxYear')" path="taxYear">
          <NSelect
            v-model:value="model.taxYear"
            :options="yearOptions"
            :placeholder="$t('page.hrm.salary.tax.opening.form.taxYear')"
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.tax.opening.untilMonth')" path="untilMonth">
          <NSelect
            v-model:value="model.untilMonth"
            :options="monthOptions"
            :placeholder="$t('page.hrm.salary.tax.opening.form.untilMonth')"
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.tax.opening.incomeTotal')" path="incomeTotal">
          <NInputNumber v-model:value="model.incomeTotal" :min="0" class="w-full" placeholder="0.00" />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.tax.opening.specialDeductionTotal')" path="specialDeductionTotal">
          <NInputNumber v-model:value="model.specialDeductionTotal" :min="0" class="w-full" placeholder="0.00" />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.tax.opening.specialExtraTotal')" path="specialExtraTotal">
          <NInputNumber v-model:value="model.specialExtraTotal" :min="0" class="w-full" placeholder="0.00" />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.tax.opening.exemptTotal')" path="exemptTotal">
          <NInputNumber v-model:value="model.exemptTotal" :min="0" class="w-full" placeholder="0.00" />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.tax.opening.otherDeductionTotal')" path="otherDeductionTotal">
          <NInputNumber v-model:value="model.otherDeductionTotal" :min="0" class="w-full" placeholder="0.00" />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.tax.opening.taxableIncomeTotal')" path="taxableIncomeTotal">
          <NInputNumber v-model:value="model.taxableIncomeTotal" class="w-full" placeholder="0.00" />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.tax.opening.taxPaidTotal')" path="taxPaidTotal">
          <NInputNumber v-model:value="model.taxPaidTotal" :min="0" class="w-full" placeholder="0.00" />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.tax.opening.remark')" path="remark">
          <NInput
            v-model:value="model.remark"
            type="textarea"
            :placeholder="$t('page.hrm.salary.tax.opening.remark')"
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
