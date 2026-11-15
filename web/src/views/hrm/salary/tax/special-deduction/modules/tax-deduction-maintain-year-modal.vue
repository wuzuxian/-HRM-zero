<script setup lang="ts">
import { reactive, ref, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { GetEmployeeList, MaintainTaxSpecialDeductionYear } from '@/service/api';
import { $t } from '@/locales';
import { hrmTaxDeductTypeOptions } from '@/constants/business';

defineOptions({
  name: 'TaxDeductionMaintainYearModal'
});

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

const title = ref('');
title.value = $t('page.hrm.salary.tax.deduction.maintainYearTitle');

// ---- employee options ----
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

const model = reactive({
  employeeUuid: '',
  deductYear: String(new Date().getFullYear()),
  deductType: 'CHILDREN_EDU' as Api.Hrm.TaxDeductionType,
  amount: 0,
  months: [] as number[]
});

const yearOptions = (() => {
  const now = new Date().getFullYear();
  return [now + 1, now, now - 1, now - 2].map(y => ({ label: String(y), value: String(y) }));
})();

const monthOptions = Array.from({ length: 12 }, (_, i) => ({ label: String(i + 1), value: i + 1 }));

function handleInitModel() {
  model.employeeUuid = '';
  model.deductYear = String(new Date().getFullYear());
  model.deductType = 'CHILDREN_EDU';
  model.amount = 0;
  model.months = [];
}

function closeModel() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();
  confirmStartLoading();
  const { error, data: resp } = await MaintainTaxSpecialDeductionYear({
    employeeUuid: model.employeeUuid,
    deductYear: model.deductYear,
    deductType: model.deductType,
    amount: model.amount,
    months: model.months.length > 0 ? model.months : undefined
  });
  confirmEndLoading();
  if (error) return;
  window.$message?.success(`${$t('page.hrm.salary.tax.deduction.saved')}: ${resp?.saved ?? 0}`);
  closeModel();
  emit('submitted');
}

watch(visible, () => {
  if (visible.value) {
    handleInitModel();
    restoreValidation();
    loadEmployeeOptions();
  }
});
</script>

<template>
  <NModal v-model:show="visible" :title="title" preset="card" class="w-560px">
    <NForm ref="formRef" :model="model" label-placement="left" :label-width="140">
      <NFormItem
        :label="$t('page.hrm.salary.tax.deduction.employeeName')"
        path="employeeUuid"
        :rule="defaultRequiredRule"
      >
        <NSelect
          v-model:value="model.employeeUuid"
          :options="employeeOptions"
          :placeholder="$t('page.hrm.salary.tax.deduction.form.employeeName')"
          filterable
        />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.salary.tax.deduction.deductYear')" path="deductYear" :rule="defaultRequiredRule">
        <NSelect
          v-model:value="model.deductYear"
          :options="yearOptions"
          :placeholder="$t('page.hrm.salary.tax.deduction.form.deductYear')"
        />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.salary.tax.deduction.deductType')" path="deductType" :rule="defaultRequiredRule">
        <NSelect
          v-model:value="model.deductType"
          :options="hrmTaxDeductTypeOptions"
          :placeholder="$t('page.hrm.salary.tax.deduction.form.deductYear')"
        />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.salary.tax.deduction.amount')" path="amount" :rule="defaultRequiredRule">
        <NInputNumber v-model:value="model.amount" :min="0" class="w-full" placeholder="0.00" />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.salary.tax.deduction.form.months')" path="months">
        <NSelect
          v-model:value="model.months"
          :options="monthOptions"
          multiple
          clearable
          :placeholder="$t('page.hrm.salary.tax.deduction.form.months')"
        />
      </NFormItem>
    </NForm>
    <template #footer>
      <NSpace :size="16" justify="end">
        <NButton @click="closeModel">{{ $t('common.cancel') }}</NButton>
        <NButton type="primary" :loading="confirmLoading" @click="handleSubmit">
          {{ $t('common.confirm') }}
        </NButton>
      </NSpace>
    </template>
  </NModal>
</template>

<style scoped></style>
