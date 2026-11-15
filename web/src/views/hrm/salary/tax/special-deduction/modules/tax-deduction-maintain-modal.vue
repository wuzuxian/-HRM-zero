<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { GetEmployeeList, GetTaxSpecialDeductionList, MaintainTaxSpecialDeduction } from '@/service/api';
import { $t } from '@/locales';
import { hrmTaxDeductTypeOptions } from '@/constants/business';

defineOptions({
  name: 'TaxDeductionMaintainModal'
});

interface Props {
  /** 预选员工（行操作「维护」带出）；空=弹窗内选择 */
  employeeUuid?: string;
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

const title = computed(() => $t('page.hrm.salary.tax.deduction.maintainModalTitle'));

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
  deductYear: String(new Date().getFullYear())
});

const yearOptions = (() => {
  const now = new Date().getFullYear();
  return [now + 1, now, now - 1, now - 2].map(y => ({ label: String(y), value: String(y) }));
})();

// ---- 12 格 × 7 类月度金额 ----
type Grid = Record<number, Partial<Record<Api.Hrm.TaxDeductionType, number>>>;

const grid = ref<Grid>({});

const monthLabels = computed(() => Array.from({ length: 12 }, (_, i) => `${i + 1}`));

function cellAmount(month: number, deductType: Api.Hrm.TaxDeductionType): number | undefined {
  return grid.value[month]?.[deductType];
}

function setCellAmount(month: number, deductType: Api.Hrm.TaxDeductionType, value: number | null) {
  if (!grid.value[month]) {
    grid.value[month] = {};
  }
  if (value === null || Number.isNaN(value)) {
    delete grid.value[month][deductType];
  } else {
    grid.value[month][deductType] = value;
  }
}

// ---- 打开弹窗时回填该员工该年度已有数据 ----
async function loadExisting() {
  grid.value = {};
  if (!model.employeeUuid || !model.deductYear) return;
  const { data, error } = await GetTaxSpecialDeductionList({
    current: 1,
    size: 200,
    employeeName: null,
    deductYear: model.deductYear,
    deductType: null,
    source: null
  });
  if (error || !data) return;
  for (const rec of data.records ?? []) {
    if (rec.employeeUuid !== model.employeeUuid) continue;
    if (!grid.value[rec.deductMonth]) {
      grid.value[rec.deductMonth] = {};
    }
    grid.value[rec.deductMonth][rec.deductType] = rec.amount;
  }
}

function closeModel() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();

  const rows: Api.Hrm.TaxSpecialDeductionMaintainRow[] = [];
  for (const monthStr of Object.keys(grid.value)) {
    const month = Number(monthStr);
    for (const [deductType, amount] of Object.entries(grid.value[month])) {
      rows.push({ deductMonth: month, deductType: deductType as Api.Hrm.TaxDeductionType, amount: amount as number });
    }
  }
  if (rows.length === 0) {
    window.$message?.warning($t('page.hrm.salary.tax.deduction.form.amount'));
    return;
  }

  confirmStartLoading();
  const { error, data: resp } = await MaintainTaxSpecialDeduction({
    employeeUuid: model.employeeUuid,
    deductYear: model.deductYear,
    rows
  });
  confirmEndLoading();
  if (error) return;
  window.$message?.success(`${$t('page.hrm.salary.tax.deduction.saved')}: ${resp?.saved ?? rows.length}`);
  closeModel();
  emit('submitted');
}

watch(visible, () => {
  if (visible.value) {
    model.employeeUuid = props.employeeUuid ?? '';
    model.deductYear = String(new Date().getFullYear());
    restoreValidation();
    loadEmployeeOptions().then(() => loadExisting());
  }
});
</script>

<script lang="ts">
const deductTypes: Api.Hrm.TaxDeductionType[] = [
  'CHILDREN_EDU',
  'CONTINUING_EDU',
  'CRITICAL_ILLNESS',
  'HOUSE_LOAN',
  'HOUSE_RENT',
  'SUPPORT_ELDERLY',
  'INFANT_CARE'
];

export default { deductTypes };
</script>

<template>
  <NModal v-model:show="visible" :title="title" preset="card" class="w-880px">
    <NForm ref="formRef" :model="model" label-placement="left" :label-width="100">
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
          :disabled="!!props.employeeUuid"
          @update:value="loadExisting"
        />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.salary.tax.deduction.deductYear')" path="deductYear" :rule="defaultRequiredRule">
        <NSelect
          v-model:value="model.deductYear"
          :options="yearOptions"
          :placeholder="$t('page.hrm.salary.tax.deduction.form.deductYear')"
          @update:value="loadExisting"
        />
      </NFormItem>
    </NForm>
    <NScrollbar class="h-360px">
      <table class="w-full border-collapse text-center text-13px">
        <thead>
          <tr>
            <th class="border border-gray-200 bg-gray-50 px-8px py-6px text-left">
              {{ $t('page.hrm.salary.tax.deduction.deductType') }}
            </th>
            <th v-for="m in monthLabels" :key="m" class="w-56px border border-gray-200 bg-gray-50 px-4px py-6px">
              {{ m }}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="dt in deductTypes" :key="dt">
            <td class="whitespace-nowrap border border-gray-200 px-8px py-4px text-left">
              {{ $t(`page.hrm.salary.tax.deduction.deductType.${dt}`) }}
            </td>
            <td v-for="(m, idx) in monthLabels" :key="`${dt}-${idx}`" class="border border-gray-200 px-2px py-2px">
              <NInputNumber
                :value="cellAmount(idx + 1, dt)"
                size="tiny"
                :min="0"
                :show-button="false"
                class="w-full"
                placeholder="0"
                @update:value="(v: number | null) => setCellAmount(idx + 1, dt, v)"
              />
            </td>
          </tr>
        </tbody>
      </table>
    </NScrollbar>
    <div class="mt-8px text-12px text-gray-400">{{ $t('page.hrm.salary.tax.deduction.chooseFileTip') }}</div>
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
