<script setup lang="tsx">
import { computed, reactive, ref, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { AddRegularization, EditRegularization, GetEmployeeList } from '@/service/api';
import { $t } from '@/locales';
import { hrmEmployeePersonStatusRecord } from '@/constants/business';

defineOptions({
  name: 'RegularizationOperateModal'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Hrm.HrmRegularization | null;
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
    add: $t('page.hrm.employeeRelation.regularization.addRegularization'),
    edit: $t('page.hrm.employeeRelation.regularization.editRegularization')
  };
  return titles[props.operateType];
});

type Model = Api.Hrm.AddRegularizationRequest;

const model: Model = reactive(createDefaultModel());

function createDefaultModel(): Model {
  return {
    employeeUuid: '',
    employeeNo: '',
    employeeName: '',
    orgUuid: '',
    companyUuid: '',
    position: '',
    post: '',
    personStatus: '',
    hireDate: '',
    expectedFormalDate: '',
    formalDate: '',
    remark: ''
  };
}

type RuleKey = Extract<keyof Model, 'employeeUuid' | 'formalDate'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  employeeUuid: defaultRequiredRule,
  formalDate: defaultRequiredRule
};

// ---- employee options (选择员工带出档案快照, spec 用户故事 20) ----
interface EmployeeOption {
  label: string;
  value: string;
}

const employeeOptions = ref<EmployeeOption[]>([]);
const employeeMap = ref<Map<string, Api.Hrm.Employee>>(new Map());

function personStatusLabel(status: string) {
  const key = (hrmEmployeePersonStatusRecord as Record<string, App.I18n.I18nKey>)[status];
  return key ? $t(key) : status;
}

async function loadEmployeeOptions() {
  if (employeeOptions.value.length > 0) return;
  const { data, error } = await GetEmployeeList({ current: 1, size: 200 });
  if (!error && data) {
    const options: EmployeeOption[] = [];
    const map = new Map<string, Api.Hrm.Employee>();
    for (const rec of data.records ?? []) {
      const label = `${rec.employeeNo} ${rec.name}（${personStatusLabel(rec.personStatus)}）`;
      options.push({ label, value: rec.uuid });
      map.set(rec.uuid, rec);
    }
    employeeOptions.value = options;
    employeeMap.value = map;
  }
}

/** 选择员工 → 带出档案快照（工号/姓名/部门/公司/职位/职务/入职日期/预计转正日期） */
function handleEmployeeChange(uuid: string) {
  const emp = employeeMap.value.get(uuid);
  if (!emp) return;
  model.employeeNo = emp.employeeNo;
  model.employeeName = emp.name;
  model.orgUuid = emp.orgUuid;
  model.companyUuid = emp.companyUuid;
  model.position = emp.position;
  model.post = emp.post;
  model.personStatus = emp.personStatus;
  model.hireDate = emp.hireDate;
  model.expectedFormalDate = emp.formalDate;
}

function handleInitModel() {
  Object.assign(model, createDefaultModel());

  if (props.operateType === 'edit' && props.rowData) {
    Object.assign(model, props.rowData);
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
      ? AddRegularization({ ...model })
      : EditRegularization({ ...model, uuid: props.rowData?.uuid ?? '' });

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
  <NModal v-model:show="visible" :title="title" preset="card" class="w-720px">
    <NScrollbar class="max-h-480px pr-20px">
      <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="110">
        <NDivider :title="$t('page.hrm.employeeRelation.regularization.sectionEmployee')" title-placement="left" />
        <NGrid responsive="screen" item-responsive>
          <NFormItemGi
            span="24 s:12"
            :label="$t('page.hrm.employeeRelation.regularization.selectEmployee')"
            path="employeeUuid"
          >
            <NSelect
              v-model:value="model.employeeUuid"
              :options="employeeOptions"
              filterable
              :disabled="operateType === 'edit'"
              :placeholder="$t('page.hrm.employeeRelation.regularization.form.selectEmployee')"
              @update:value="handleEmployeeChange"
            />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:8"
            :label="$t('page.hrm.employeeRelation.regularization.employeeNo')"
            path="employeeNo"
          >
            <NInput v-model:value="model.employeeNo" disabled />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:8"
            :label="$t('page.hrm.employeeRelation.regularization.employeeName')"
            path="employeeName"
          >
            <NInput v-model:value="model.employeeName" disabled />
          </NFormItemGi>
        </NGrid>

        <NDivider :title="$t('page.hrm.employeeRelation.regularization.sectionSnapshot')" title-placement="left" />
        <NGrid responsive="screen" item-responsive>
          <NFormItemGi
            span="24 s:12 m:8"
            :label="$t('page.hrm.employeeRelation.regularization.orgUuid')"
            path="orgUuid"
          >
            <NInput
              v-model:value="model.orgUuid"
              disabled
              :placeholder="$t('page.hrm.employeeRelation.regularization.form.none')"
            />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:8"
            :label="$t('page.hrm.employeeRelation.regularization.companyUuid')"
            path="companyUuid"
          >
            <NInput
              v-model:value="model.companyUuid"
              disabled
              :placeholder="$t('page.hrm.employeeRelation.regularization.form.none')"
            />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:8"
            :label="$t('page.hrm.employeeRelation.regularization.position')"
            path="position"
          >
            <NInput
              v-model:value="model.position"
              disabled
              :placeholder="$t('page.hrm.employeeRelation.regularization.form.none')"
            />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employeeRelation.regularization.post')" path="post">
            <NInput
              v-model:value="model.post"
              disabled
              :placeholder="$t('page.hrm.employeeRelation.regularization.form.none')"
            />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:8"
            :label="$t('page.hrm.employeeRelation.regularization.hireDate')"
            path="hireDate"
          >
            <NInput
              v-model:value="model.hireDate"
              disabled
              :placeholder="$t('page.hrm.employeeRelation.regularization.form.none')"
            />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:8"
            :label="$t('page.hrm.employeeRelation.regularization.expectedFormalDate')"
            path="expectedFormalDate"
          >
            <NInput
              v-model:value="model.expectedFormalDate"
              disabled
              :placeholder="$t('page.hrm.employeeRelation.regularization.form.none')"
            />
          </NFormItemGi>
        </NGrid>

        <NDivider :title="$t('page.hrm.employeeRelation.regularization.sectionRegular')" title-placement="left" />
        <NGrid responsive="screen" item-responsive>
          <NFormItemGi
            span="24 s:12 m:8"
            :label="$t('page.hrm.employeeRelation.regularization.formalDate')"
            path="formalDate"
          >
            <NDatePicker
              v-model:formatted-value="model.formalDate"
              type="date"
              value-format="yyyy-MM-dd"
              class="w-full"
            />
          </NFormItemGi>
          <NFormItemGi span="24" :label="$t('page.hrm.employeeRelation.regularization.remark')" path="remark">
            <NTextarea
              v-model:value="model.remark"
              :rows="2"
              :placeholder="$t('page.hrm.employeeRelation.regularization.form.remark')"
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
