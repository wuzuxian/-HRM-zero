<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { AddEmployeeContract, EditEmployeeContract, GetEmployeeList } from '@/service/api';
import { $t } from '@/locales';

defineOptions({
  name: 'ContractOperateModal'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Hrm.HrmEmployeeContract | null;
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
    add: $t('page.hrm.employee.contractExpiring.addContract'),
    edit: $t('page.hrm.employee.contractExpiring.editContract')
  };
  return titles[props.operateType];
});

type Model = Api.Hrm.AddContractRequest & { status: string };

const model: Model = reactive(createDefaultModel());

function createDefaultModel(): Model {
  return {
    employeeUuid: '',
    contractNo: '',
    startDate: '',
    endDate: '',
    status: '1',
    remark: ''
  };
}

type RuleKey = Extract<keyof Model, 'employeeUuid' | 'startDate' | 'endDate'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  employeeUuid: defaultRequiredRule,
  startDate: defaultRequiredRule,
  endDate: defaultRequiredRule
};

// ---- employee options (选择员工带出档案快照) ----
interface EmployeeOption {
  label: string;
  value: string;
}

const employeeOptions = ref<EmployeeOption[]>([]);
const employeeMap = ref<Map<string, Api.Hrm.Employee>>(new Map());

async function loadEmployeeOptions() {
  if (employeeOptions.value.length > 0) return;
  const { data, error } = await GetEmployeeList({ current: 1, size: 200 });
  if (!error && data) {
    const options: EmployeeOption[] = [];
    const map = new Map<string, Api.Hrm.Employee>();
    for (const rec of data.records ?? []) {
      const label = `${rec.employeeNo} ${rec.name}`;
      options.push({ label, value: rec.uuid });
      map.set(rec.uuid, rec);
    }
    employeeOptions.value = options;
    employeeMap.value = map;
  }
}

/** 选择员工 → 带出快照（工号/姓名在列表展示，模型只传 uuid） */
function handleEmployeeChange(_uuid: string) {
  // 快照由后端按 employeeUuid 带出，前端无需回填
}

const statusOptions = computed(() => [
  { label: $t('page.hrm.employee.contractExpiring.statusActive'), value: '1' },
  { label: $t('page.hrm.employee.contractExpiring.statusExpired'), value: '2' },
  { label: $t('page.hrm.employee.contractExpiring.statusTerminated'), value: '3' }
]);

function handleInitModel() {
  Object.assign(model, createDefaultModel());

  if (props.operateType === 'edit' && props.rowData) {
    Object.assign(model, {
      employeeUuid: props.rowData.employeeUuid,
      contractNo: props.rowData.contractNo,
      startDate: props.rowData.startDate,
      endDate: props.rowData.endDate,
      status: props.rowData.status || '1',
      remark: props.rowData.remark
    });
  }
  loadEmployeeOptions();
}

function closeModel() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();
  confirmStartLoading();

  const request =
    props.operateType === 'add'
      ? AddEmployeeContract({ ...model })
      : EditEmployeeContract({ ...model, uuid: props.rowData?.uuid ?? '' });

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
  <NModal v-model:show="visible" preset="card" class="w-700px" :title="title">
    <NScrollbar class="max-h-500px">
      <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="90">
        <NGrid responsive="screen" item-responsive>
          <NFormItemGi
            span="24 s:12 m:12"
            :label="$t('page.hrm.employee.contractExpiring.selectEmployee')"
            path="employeeUuid"
          >
            <NSelect
              v-model:value="model.employeeUuid"
              :options="employeeOptions"
              :placeholder="$t('page.hrm.employee.contractExpiring.form.selectEmployee')"
              :disabled="operateType === 'edit'"
              filterable
              @update:value="handleEmployeeChange"
            />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:12"
            :label="$t('page.hrm.employee.contractExpiring.contractNo')"
            path="contractNo"
          >
            <NInput
              v-model:value="model.contractNo"
              :placeholder="$t('page.hrm.employee.contractExpiring.form.contractNo')"
            />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:12"
            :label="$t('page.hrm.employee.contractExpiring.startDate')"
            path="startDate"
          >
            <NDatePicker
              v-model:formatted-value="model.startDate"
              type="date"
              value-format="yyyy-MM-dd"
              class="w-full"
            />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:12"
            :label="$t('page.hrm.employee.contractExpiring.endDate')"
            path="endDate"
          >
            <NDatePicker
              v-model:formatted-value="model.endDate"
              type="date"
              value-format="yyyy-MM-dd"
              class="w-full"
            />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:12" :label="$t('page.hrm.employee.contractExpiring.status')" path="status">
            <NSelect v-model:value="model.status" :options="statusOptions" />
          </NFormItemGi>
          <NFormItemGi span="24" :label="$t('page.hrm.employee.contractExpiring.remark')" path="remark">
            <NTextarea
              v-model:value="model.remark"
              :rows="2"
              :placeholder="$t('page.hrm.employee.contractExpiring.form.remark')"
            />
          </NFormItemGi>
        </NGrid>
      </NForm>
    </NScrollbar>
    <template #footer>
      <NSpace justify="end" :size="16">
        <NButton @click="closeModel">{{ $t('common.cancel') }}</NButton>
        <NButton type="primary" :loading="confirmLoading" @click="handleSubmit">{{ $t('common.save') }}</NButton>
      </NSpace>
    </template>
  </NModal>
</template>

<style scoped></style>
