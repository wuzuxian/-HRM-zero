<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { AddSalarySubject, EditSalarySubject, GetOrgTree } from '@/service/api';
import { $t } from '@/locales';
import { hrmSalaryStatusOptions } from '@/constants/business';
import { translateOptions } from '@/utils/common';

defineOptions({
  name: 'SalarySubjectOperateModal'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Hrm.SalarySubject | null;
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
    add: $t('page.hrm.salary.subject.addSubject'),
    edit: $t('page.hrm.salary.subject.editSubject')
  };
  return titles[props.operateType];
});

type Model = Pick<
  Api.Hrm.AddSalarySubjectRequest,
  'subjectName' | 'orgUuid' | 'payAccount' | 'payBank' | 'status' | 'remark'
>;

const model: Model = reactive(createDefaultModel());

function createDefaultModel(): Model {
  return {
    subjectName: '',
    orgUuid: '',
    payAccount: '',
    payBank: '',
    status: '1',
    remark: ''
  };
}

type RuleKey = Extract<keyof Model, 'subjectName' | 'orgUuid'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  subjectName: defaultRequiredRule,
  orgUuid: defaultRequiredRule
};

// ---- org tree options (关联公司 from 组织管理) ----
interface OrgOption {
  label: string;
  value: string;
  children?: OrgOption[];
}

const orgOptions = ref<OrgOption[]>([]);

function toOrgOptions(nodes: Api.Hrm.Org[]): OrgOption[] {
  return nodes.map(node => ({
    label: node.orgName,
    value: node.uuid,
    children: node.children && node.children.length > 0 ? toOrgOptions(node.children) : undefined
  }));
}

async function loadOrgOptions() {
  if (orgOptions.value.length > 0) return;
  const { data, error } = await GetOrgTree();
  if (!error && data) {
    orgOptions.value = toOrgOptions(data.records ?? []);
  }
}

function handleInitModel() {
  Object.assign(model, createDefaultModel());

  if (props.operateType === 'edit' && props.rowData) {
    Object.assign(model, props.rowData);
  }
  loadOrgOptions();
}

function closeModel() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();
  confirmStartLoading();

  const request =
    props.operateType === 'add'
      ? AddSalarySubject({ ...model })
      : EditSalarySubject({ ...model, uuid: props.rowData?.uuid ?? '' });

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
        <NFormItem :label="$t('page.hrm.salary.subject.subjectName')" path="subjectName">
          <NInput
            v-model:value="model.subjectName"
            :placeholder="$t('page.hrm.salary.subject.form.subjectName')"
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.subject.orgName')" path="orgUuid">
          <NSelect
            v-model:value="model.orgUuid"
            :options="orgOptions"
            :placeholder="$t('page.hrm.salary.subject.form.orgName')"
            filterable
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.subject.payAccount')" path="payAccount">
          <NInput
            v-model:value="model.payAccount"
            :placeholder="$t('page.hrm.salary.subject.form.payAccount')"
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.subject.payBank')" path="payBank">
          <NInput v-model:value="model.payBank" :placeholder="$t('page.hrm.salary.subject.form.payBank')" />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.subject.status')" path="status">
          <NRadioGroup v-model:value="model.status">
            <NRadio
              v-for="item in hrmSalaryStatusOptions"
              :key="item.value"
              :value="item.value"
              :label="$t(item.label)"
            />
          </NRadioGroup>
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.subject.remark')" path="remark">
          <NInput
            v-model:value="model.remark"
            type="textarea"
            :placeholder="$t('page.hrm.salary.subject.form.remark')"
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
