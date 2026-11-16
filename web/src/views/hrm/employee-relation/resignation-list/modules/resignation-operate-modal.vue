<script setup lang="tsx">
import { computed, reactive, ref, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { AddResignation, EditResignation, GetEmployeeList } from '@/service/api';
import { $t } from '@/locales';
import { hrmResignTypeOptions } from '@/constants/business';
import { translateOptions } from '@/utils/common';

defineOptions({
  name: 'ResignationOperateModal'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Hrm.HrmResignation | null;
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
    add: $t('page.hrm.employeeRelation.resignation.addResignation'),
    edit: $t('page.hrm.employeeRelation.resignation.editResignation')
  };
  return titles[props.operateType];
});

type Model = Omit<Api.Hrm.AddResignationRequest, 'applyDate' | 'resignDate' | 'lastWorkDate' | 'salarySettleDate'> & {
  // NDatePicker 的 formatted-value 不能是空串（解析 "" 抛 Invalid time value），未选择时必须为 null
  applyDate: string | null;
  resignDate: string | null;
  lastWorkDate: string | null;
  salarySettleDate: string | null;
};

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
    resignType: '',
    resignReason: '',
    applyDate: null,
    resignDate: null,
    lastWorkDate: null,
    handoverPerson: '',
    salarySettleDate: null
  };
}

type RuleKey = Extract<keyof Model, 'employeeUuid' | 'resignType' | 'resignDate'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  employeeUuid: defaultRequiredRule,
  resignType: defaultRequiredRule,
  resignDate: defaultRequiredRule
};

// ---- employee options (选择员工带出档案快照, spec 用户故事 22) ----
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
      options.push({ label: `${rec.employeeNo} ${rec.name}`, value: rec.uuid });
      map.set(rec.uuid, rec);
    }
    employeeOptions.value = options;
    employeeMap.value = map;
  }
}

/** 选择员工 → 带出档案快照（工号/姓名/部门/公司/职位/职务） */
function handleEmployeeChange(uuid: string) {
  const emp = employeeMap.value.get(uuid);
  if (!emp) return;
  model.employeeNo = emp.employeeNo;
  model.employeeName = emp.name;
  model.orgUuid = emp.orgUuid;
  model.companyUuid = emp.companyUuid;
  model.position = emp.position;
  model.post = emp.post;
}

function handleInitModel() {
  Object.assign(model, createDefaultModel());

  if (props.operateType === 'edit' && props.rowData) {
    Object.assign(model, props.rowData);
    // DB 列 NOT NULL DEFAULT ''，编辑回填可能拿到空串；空串会让 NDatePicker 抛 Invalid time value
    model.applyDate = props.rowData.applyDate || null;
    model.resignDate = props.rowData.resignDate || null;
    model.lastWorkDate = props.rowData.lastWorkDate || null;
    model.salarySettleDate = props.rowData.salarySettleDate || null;
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
      ? AddResignation({
          ...model,
          applyDate: model.applyDate ?? '',
          resignDate: model.resignDate ?? '',
          lastWorkDate: model.lastWorkDate ?? '',
          salarySettleDate: model.salarySettleDate ?? ''
        })
      : EditResignation({
          ...model,
          applyDate: model.applyDate ?? '',
          resignDate: model.resignDate ?? '',
          lastWorkDate: model.lastWorkDate ?? '',
          salarySettleDate: model.salarySettleDate ?? '',
          uuid: props.rowData?.uuid ?? ''
        });

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
        <NDivider :title="$t('page.hrm.employeeRelation.resignation.sectionEmployee')" title-placement="left" />
        <NGrid responsive="screen" item-responsive>
          <NFormItemGi
            span="24 s:12"
            :label="$t('page.hrm.employeeRelation.resignation.selectEmployee')"
            path="employeeUuid"
          >
            <NSelect
              v-model:value="model.employeeUuid"
              :options="employeeOptions"
              filterable
              :disabled="operateType === 'edit'"
              :placeholder="$t('page.hrm.employeeRelation.resignation.form.selectEmployee')"
              @update:value="handleEmployeeChange"
            />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:8"
            :label="$t('page.hrm.employeeRelation.resignation.employeeNo')"
            path="employeeNo"
          >
            <NInput v-model:value="model.employeeNo" disabled />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:8"
            :label="$t('page.hrm.employeeRelation.resignation.employeeName')"
            path="employeeName"
          >
            <NInput v-model:value="model.employeeName" disabled />
          </NFormItemGi>
        </NGrid>

        <NDivider :title="$t('page.hrm.employeeRelation.resignation.sectionSnapshot')" title-placement="left" />
        <NGrid responsive="screen" item-responsive>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employeeRelation.resignation.orgUuid')" path="orgUuid">
            <NInput
              v-model:value="model.orgUuid"
              disabled
              :placeholder="$t('page.hrm.employeeRelation.resignation.form.none')"
            />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:8"
            :label="$t('page.hrm.employeeRelation.resignation.companyUuid')"
            path="companyUuid"
          >
            <NInput
              v-model:value="model.companyUuid"
              disabled
              :placeholder="$t('page.hrm.employeeRelation.resignation.form.none')"
            />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employeeRelation.resignation.position')" path="position">
            <NInput
              v-model:value="model.position"
              disabled
              :placeholder="$t('page.hrm.employeeRelation.resignation.form.none')"
            />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employeeRelation.resignation.post')" path="post">
            <NInput
              v-model:value="model.post"
              disabled
              :placeholder="$t('page.hrm.employeeRelation.resignation.form.none')"
            />
          </NFormItemGi>
        </NGrid>

        <NDivider :title="$t('page.hrm.employeeRelation.resignation.sectionResign')" title-placement="left" />
        <NGrid responsive="screen" item-responsive>
          <NFormItemGi
            span="24 s:12 m:8"
            :label="$t('page.hrm.employeeRelation.resignation.resignType.title')"
            path="resignType"
          >
            <NSelect
              v-model:value="model.resignType"
              :options="translateOptions(hrmResignTypeOptions)"
              :placeholder="$t('page.hrm.employeeRelation.resignation.form.resignType')"
              clearable
            />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:8"
            :label="$t('page.hrm.employeeRelation.resignation.applyDate')"
            path="applyDate"
          >
            <NDatePicker
              v-model:formatted-value="model.applyDate"
              type="date"
              value-format="yyyy-MM-dd"
              class="w-full"
            />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:8"
            :label="$t('page.hrm.employeeRelation.resignation.resignDate')"
            path="resignDate"
          >
            <NDatePicker
              v-model:formatted-value="model.resignDate"
              type="date"
              value-format="yyyy-MM-dd"
              class="w-full"
            />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:8"
            :label="$t('page.hrm.employeeRelation.resignation.lastWorkDate')"
            path="lastWorkDate"
          >
            <NDatePicker
              v-model:formatted-value="model.lastWorkDate"
              type="date"
              value-format="yyyy-MM-dd"
              class="w-full"
            />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:8"
            :label="$t('page.hrm.employeeRelation.resignation.salarySettleDate')"
            path="salarySettleDate"
          >
            <NDatePicker
              v-model:formatted-value="model.salarySettleDate"
              type="date"
              value-format="yyyy-MM-dd"
              class="w-full"
            />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:8"
            :label="$t('page.hrm.employeeRelation.resignation.handoverPerson')"
            path="handoverPerson"
          >
            <NInput
              v-model:value="model.handoverPerson"
              :placeholder="$t('page.hrm.employeeRelation.resignation.form.handoverPerson')"
            />
          </NFormItemGi>
          <NFormItemGi span="24" :label="$t('page.hrm.employeeRelation.resignation.resignReason')" path="resignReason">
            <NInput
              v-model:value="model.resignReason"
              type="textarea"
              :rows="2"
              :placeholder="$t('page.hrm.employeeRelation.resignation.form.resignReason')"
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
