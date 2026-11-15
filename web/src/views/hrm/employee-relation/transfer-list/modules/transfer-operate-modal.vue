<script setup lang="tsx">
import { computed, reactive, ref, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { AddTransfer, EditTransfer, GetOrgTree } from '@/service/api';
import { $t } from '@/locales';
import { hrmTransferChangeTypeOptions } from '@/constants/business';
import { translateOptions } from '@/utils/common';

defineOptions({
  name: 'TransferOperateModal'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Hrm.HrmTransfer | null;
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
    add: $t('page.hrm.employeeRelation.transfer.addTransfer'),
    edit: $t('page.hrm.employeeRelation.transfer.editTransfer')
  };
  return titles[props.operateType];
});

type Model = Api.Hrm.AddTransferRequest;

const model: Model = reactive(createDefaultModel());

function createDefaultModel(): Model {
  return {
    employeeUuid: '',
    employeeNo: '',
    employeeName: '',
    changeType: 'post',
    originalCompany: '',
    originalOrg: '',
    originalPosition: '',
    originalPost: '',
    newCompany: '',
    newOrg: '',
    newPosition: '',
    newPost: '',
    newSalary: 0,
    transferReason: '',
    effectiveImmediately: 1,
    effectiveDate: ''
  };
}

type RuleKey = Extract<keyof Model, 'employeeNo' | 'employeeName' | 'changeType'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  employeeNo: defaultRequiredRule,
  employeeName: defaultRequiredRule,
  changeType: defaultRequiredRule
};

// ---- org tree options (公司/部门 from 组织管理) ----
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

// delayed bills need a 生效日期
const effectiveDateRequired = computed(() => model.effectiveImmediately === 0);

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
  if (effectiveDateRequired.value && !model.effectiveDate) {
    window.$message?.error($t('page.hrm.employeeRelation.transfer.form.effectiveDateRequired'));
    return;
  }
  confirmStartLoading();

  const request =
    props.operateType === 'add'
      ? AddTransfer({ ...model })
      : EditTransfer({ ...model, uuid: props.rowData?.uuid ?? '' });

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
  <NModal v-model:show="visible" :title="title" preset="card" class="w-760px">
    <NScrollbar class="max-h-480px pr-20px">
      <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="100">
        <NDivider :title="$t('page.hrm.employeeRelation.transfer.sectionEmployee')" title-placement="left" />
        <NGrid responsive="screen" item-responsive>
          <NFormItemGi
            span="24 s:12 m:8"
            :label="$t('page.hrm.employeeRelation.transfer.employeeNo')"
            path="employeeNo"
          >
            <NInput
              v-model:value="model.employeeNo"
              :placeholder="$t('page.hrm.employeeRelation.transfer.form.employeeNo')"
            />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:8"
            :label="$t('page.hrm.employeeRelation.transfer.employeeName')"
            path="employeeName"
          >
            <NInput
              v-model:value="model.employeeName"
              :placeholder="$t('page.hrm.employeeRelation.transfer.form.employeeName')"
            />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:8"
            :label="$t('page.hrm.employeeRelation.transfer.changeType.title')"
            path="changeType"
          >
            <NSelect
              v-model:value="model.changeType"
              :options="translateOptions(hrmTransferChangeTypeOptions)"
              :placeholder="$t('page.hrm.employeeRelation.transfer.form.changeType')"
            />
          </NFormItemGi>
        </NGrid>

        <NDivider :title="$t('page.hrm.employeeRelation.transfer.sectionOriginal')" title-placement="left" />
        <NGrid responsive="screen" item-responsive>
          <NFormItemGi
            span="24 s:12 m:8"
            :label="$t('page.hrm.employeeRelation.transfer.originalCompany')"
            path="originalCompany"
          >
            <NTreeSelect
              v-model:value="model.originalCompany"
              :options="orgOptions"
              clearable
              :placeholder="$t('page.hrm.employeeRelation.transfer.form.org')"
            />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:8"
            :label="$t('page.hrm.employeeRelation.transfer.originalOrg')"
            path="originalOrg"
          >
            <NTreeSelect
              v-model:value="model.originalOrg"
              :options="orgOptions"
              clearable
              :placeholder="$t('page.hrm.employeeRelation.transfer.form.org')"
            />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:8"
            :label="$t('page.hrm.employeeRelation.transfer.originalPosition')"
            path="originalPosition"
          >
            <NInput v-model:value="model.originalPosition" />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:8"
            :label="$t('page.hrm.employeeRelation.transfer.originalPost')"
            path="originalPost"
          >
            <NInput v-model:value="model.originalPost" />
          </NFormItemGi>
        </NGrid>

        <NDivider :title="$t('page.hrm.employeeRelation.transfer.sectionNew')" title-placement="left" />
        <NGrid responsive="screen" item-responsive>
          <NFormItemGi
            span="24 s:12 m:8"
            :label="$t('page.hrm.employeeRelation.transfer.newCompany')"
            path="newCompany"
          >
            <NTreeSelect
              v-model:value="model.newCompany"
              :options="orgOptions"
              clearable
              :placeholder="$t('page.hrm.employeeRelation.transfer.form.org')"
            />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employeeRelation.transfer.newOrg')" path="newOrg">
            <NTreeSelect
              v-model:value="model.newOrg"
              :options="orgOptions"
              clearable
              :placeholder="$t('page.hrm.employeeRelation.transfer.form.org')"
            />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:8"
            :label="$t('page.hrm.employeeRelation.transfer.newPosition')"
            path="newPosition"
          >
            <NInput v-model:value="model.newPosition" />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employeeRelation.transfer.newPost')" path="newPost">
            <NInput v-model:value="model.newPost" />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employeeRelation.transfer.newSalary')" path="newSalary">
            <NInputNumber v-model:value="model.newSalary" :min="0" :precision="2" />
          </NFormItemGi>
        </NGrid>

        <NDivider :title="$t('page.hrm.employeeRelation.transfer.sectionEffect')" title-placement="left" />
        <NGrid responsive="screen" item-responsive>
          <NFormItemGi
            span="24 s:12 m:8"
            :label="$t('page.hrm.employeeRelation.transfer.effectiveImmediately')"
            path="effectiveImmediately"
          >
            <NRadioGroup v-model:value="model.effectiveImmediately">
              <NRadio :value="1">{{ $t('page.hrm.employeeRelation.transfer.yes') }}</NRadio>
              <NRadio :value="0">{{ $t('page.hrm.employeeRelation.transfer.no') }}</NRadio>
            </NRadioGroup>
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:8"
            :label="$t('page.hrm.employeeRelation.transfer.effectiveDate')"
            path="effectiveDate"
          >
            <NDatePicker
              v-model:formatted-value="model.effectiveDate"
              type="date"
              value-format="yyyy-MM-dd"
              :disabled="model.effectiveImmediately === 1"
              class="w-full"
            />
          </NFormItemGi>
          <NFormItemGi span="24" :label="$t('page.hrm.employeeRelation.transfer.transferReason')" path="transferReason">
            <NInput
  type="textarea"
              v-model:value="model.transferReason"
              :rows="2"
              :placeholder="$t('page.hrm.employeeRelation.transfer.form.transferReason')"
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
