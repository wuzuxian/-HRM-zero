<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import {
  AddSalaryArchive,
  EditSalaryArchive,
  GetEmployeeList,
  GetSalarySchemeList,
  GetSalarySubjectList
} from '@/service/api';
import { $t } from '@/locales';
import { hrmSalaryStatusOptions } from '@/constants/business';

defineOptions({
  name: 'SalaryArchiveOperateModal'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Hrm.SalaryArchive | null;
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
    add: $t('page.hrm.salary.archive.addArchive'),
    edit: $t('page.hrm.salary.archive.editArchive')
  };
  return titles[props.operateType];
});

type Model = {
  employeeUuid: string;
  subjectUuid: string;
  schemeUuid: string;
  fixedSalary: number | null;
  fixedAllowance: number | null;
  status: Api.Hrm.SalarySubjectStatus;
  effectiveDate: number | null;
  expireDate: number | null;
  remark: string;
};

const model: Model = reactive(createDefaultModel());

function createDefaultModel(): Model {
  return {
    employeeUuid: '',
    subjectUuid: '',
    schemeUuid: '',
    fixedSalary: 0,
    fixedAllowance: 0,
    status: '1',
    effectiveDate: null,
    expireDate: null,
    remark: ''
  };
}

type RuleKey = Extract<keyof Model, 'employeeUuid' | 'subjectUuid' | 'effectiveDate'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  employeeUuid: defaultRequiredRule,
  subjectUuid: defaultRequiredRule,
  effectiveDate: {
    required: true,
    type: 'number',
    message: $t('page.hrm.salary.archive.form.effectiveDateRequired'),
    trigger: 'change'
  }
};

// ---- options: 员工 / 主体 / 方案 ----
interface EmployeeOption {
  label: string;
  value: string;
}

const employeeOptions = ref<EmployeeOption[]>([]);
const subjectOptions = ref<CommonType.Option<string>[]>([]);
const schemeOptions = ref<CommonType.Option<string>[]>([]);

async function loadEmployeeOptions() {
  const { data, error } = await GetEmployeeList({ current: 1, size: 500 });
  if (!error && data) {
    employeeOptions.value = data.records.map(row => ({
      label: `${row.name}（${row.employeeNo}）`,
      value: row.uuid
    }));
  }
}

async function loadSubjectOptions() {
  const { data, error } = await GetSalarySubjectList({ current: 1, size: 500 });
  if (!error && data) {
    subjectOptions.value = data.records.map(row => ({ label: row.subjectName, value: row.uuid }));
  }
}

async function loadSchemeOptions() {
  const { data, error } = await GetSalarySchemeList({ current: 1, size: 500 });
  if (!error && data) {
    schemeOptions.value = data.records.map(row => ({ label: row.schemeName, value: row.uuid }));
  }
}

/** 选中员工带出的 公司/部门 展示（编辑时回显档案快照列） */
const selectedEmployeeInfo = computed(() => {
  if (props.operateType === 'edit') {
    return props.rowData
      ? { company: props.rowData.companyName, org: props.rowData.orgName }
      : { company: '-', org: '-' };
  }
  return { company: '-', org: '-' };
});

function handleInitModel() {
  Object.assign(model, createDefaultModel());

  if (props.operateType === 'edit' && props.rowData) {
    model.employeeUuid = props.rowData.employeeUuid;
    model.subjectUuid = props.rowData.subjectUuid;
    model.schemeUuid = props.rowData.schemeUuid || '';
    model.fixedSalary = props.rowData.fixedSalary;
    model.fixedAllowance = props.rowData.fixedAllowance;
    model.status = (props.rowData.status ?? '1') as Api.Hrm.SalarySubjectStatus;
    model.effectiveDate = props.rowData.effectiveDate ? new Date(props.rowData.effectiveDate).getTime() : null;
    model.expireDate = props.rowData.expireDate ? new Date(props.rowData.expireDate).getTime() : null;
    model.remark = props.rowData.remark;
  }
}

function closeModel() {
  visible.value = false;
}

function toDate(value: number | null): string {
  if (!value) return '';
  const d = new Date(value);
  const mm = `${d.getMonth() + 1}`.padStart(2, '0');
  const dd = `${d.getDate()}`.padStart(2, '0');
  return `${d.getFullYear()}-${mm}-${dd}`;
}

async function handleSubmit() {
  await validate();

  confirmStartLoading();
  const request =
    props.operateType === 'add'
      ? AddArchiveRequest()
      : EditArchiveRequest();

  const { error } = await request;
  if (!error) {
    window.$message?.success(props.operateType === 'add' ? $t('common.addSuccess') : $t('common.updateSuccess'));
    closeModel();
    emit('submitted');
  }
  confirmEndLoading();
}

function commonPayload() {
  return {
    employeeUuid: model.employeeUuid,
    subjectUuid: model.subjectUuid,
    schemeUuid: model.schemeUuid,
    fixedSalary: model.fixedSalary ?? 0,
    fixedAllowance: model.fixedAllowance ?? 0,
    status: model.status,
    effectiveDate: toDate(model.effectiveDate),
    expireDate: toDate(model.expireDate),
    remark: model.remark
  };
}

function AddArchiveRequest() {
  return AddSalaryArchive(commonPayload());
}

function EditArchiveRequest() {
  return EditSalaryArchive({
    uuid: props.rowData?.uuid ?? '',
    ...commonPayload()
  });
}

watch(visible, () => {
  if (visible.value) {
    handleInitModel();
    restoreValidation();
    loadEmployeeOptions();
    loadSubjectOptions();
    loadSchemeOptions();
  }
});
</script>

<template>
  <NModal v-model:show="visible" :title="title" preset="card" class="w-640px">
    <NScrollbar class="h-480px pr-20px">
      <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="110">
        <NFormItem :label="$t('page.hrm.salary.archive.employee')" path="employeeUuid">
          <NSelect
            v-model:value="model.employeeUuid"
            :options="employeeOptions"
            :placeholder="$t('page.hrm.salary.archive.form.employee')"
            filterable
            :disabled="operateType === 'edit'"
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.archive.companyName')" :show-feedback="false">
          <NAlert :type="selectedEmployeeInfo.company === '-' ? 'default' : 'info'" :bordered="false" class="w-full">
            {{ $t('page.hrm.salary.archive.companyName') }}: {{ selectedEmployeeInfo.company }} /
            {{ $t('page.hrm.salary.archive.orgName') }}: {{ selectedEmployeeInfo.org || '-' }}
          </NAlert>
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.archive.subject')" path="subjectUuid">
          <NSelect
            v-model:value="model.subjectUuid"
            :options="subjectOptions"
            :placeholder="$t('page.hrm.salary.archive.form.subject')"
            filterable
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.archive.scheme')" path="schemeUuid">
          <NSelect
            v-model:value="model.schemeUuid"
            :options="schemeOptions"
            :placeholder="$t('page.hrm.salary.archive.form.scheme')"
            filterable
            clearable
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.archive.fixedSalary')" path="fixedSalary">
          <NInputNumber v-model:value="model.fixedSalary" :min="0" :precision="2" class="w-220px" />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.archive.fixedAllowance')" path="fixedAllowance">
          <NInputNumber v-model:value="model.fixedAllowance" :min="0" :precision="2" class="w-220px" />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.archive.effectiveDate')" path="effectiveDate">
          <NDatePicker v-model:value="model.effectiveDate" type="date" class="w-220px" />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.archive.expireDate')" path="expireDate">
          <NDatePicker v-model:value="model.expireDate" type="date" class="w-220px" clearable />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.archive.status')" path="status">
          <NRadioGroup v-model:value="model.status">
            <NRadio
              v-for="item in hrmSalaryStatusOptions"
              :key="item.value"
              :value="item.value"
              :label="$t(item.label)"
            />
          </NRadioGroup>
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.archive.remark')" path="remark">
          <NInput v-model:value="model.remark" type="textarea" :placeholder="$t('page.hrm.salary.archive.form.remark')" />
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
