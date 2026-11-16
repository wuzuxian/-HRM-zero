<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { AddSalaryBatch, EditSalaryBatch, GetSalarySubjectList } from '@/service/api';
import { $t } from '@/locales';

defineOptions({
  name: 'SalaryBatchOperateModal'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Hrm.SalaryBatch | null;
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
    add: $t('page.hrm.salary.batch.addBatch'),
    edit: $t('page.hrm.salary.batch.editBatch')
  };
  return titles[props.operateType];
});

type Model = {
  batchName: string;
  batchType: string;
  subjectUuid: string;
  batchYear: number | null;
  batchMonth: number | null;
  remark: string;
};

const model: Model = reactive(createDefaultModel());

function createDefaultModel(): Model {
  return {
    batchName: '',
    batchType: 'monthly',
    subjectUuid: '',
    batchYear: new Date().getFullYear(),
    batchMonth: new Date().getMonth() + 1,
    remark: ''
  };
}

type RuleKey = Extract<keyof Model, 'batchName' | 'subjectUuid' | 'batchYear' | 'batchMonth'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  batchName: defaultRequiredRule,
  subjectUuid: defaultRequiredRule,
  batchYear: { required: true, type: 'number', message: $t('page.hrm.salary.batch.form.yearRequired'), trigger: 'change' },
  batchMonth: { required: true, type: 'number', message: $t('page.hrm.salary.batch.form.monthRequired'), trigger: 'change' }
};

const yearOptions = computed(() => {
  const current = new Date().getFullYear();
  const out: CommonType.Option<number>[] = [];
  for (let y = current - 5; y <= current + 1; y += 1) {
    out.push({ label: `${y}`, value: y });
  }
  return out;
});

const monthOptions = computed(() => {
  const out: CommonType.Option<number>[] = [];
  for (let m = 1; m <= 12; m += 1) {
    out.push({ label: `${m}`, value: m });
  }
  return out;
});

// ---- options: 主体 ----
const subjectOptions = ref<CommonType.Option<string>[]>([]);

async function loadSubjectOptions() {
  const { data, error } = await GetSalarySubjectList({ current: 1, size: 500 });
  if (!error && data) {
    subjectOptions.value = data.records.map(row => ({ label: row.subjectName, value: row.uuid }));
  }
}

function handleInitModel() {
  Object.assign(model, createDefaultModel());

  if (props.operateType === 'edit' && props.rowData) {
    model.batchName = props.rowData.batchName;
    model.batchType = props.rowData.batchType || 'monthly';
    model.subjectUuid = props.rowData.subjectUuid;
    model.batchYear = props.rowData.batchYear;
    model.batchMonth = props.rowData.batchMonth;
    model.remark = props.rowData.remark;
  }
}

function closeModel() {
  visible.value = false;
}

function commonPayload() {
  return {
    batchName: model.batchName,
    batchType: model.batchType,
    batchYear: model.batchYear ?? 0,
    batchMonth: model.batchMonth ?? 0,
    remark: model.remark
  };
}

async function handleSubmit() {
  await validate();

  confirmStartLoading();
  const request =
    props.operateType === 'add'
      ? AddSalaryBatch({ subjectUuid: model.subjectUuid, ...commonPayload() })
      : EditSalaryBatch({ uuid: props.rowData?.uuid ?? '', ...commonPayload() });

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
    loadSubjectOptions();
  }
});
</script>

<template>
  <NModal v-model:show="visible" :title="title" preset="card" class="w-560px">
    <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="110">
      <NFormItem :label="$t('page.hrm.salary.batch.batchName')" path="batchName">
        <NInput v-model:value="model.batchName" :placeholder="$t('page.hrm.salary.batch.form.batchName')" />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.salary.batch.subject')" path="subjectUuid">
        <NSelect
          v-model:value="model.subjectUuid"
          :options="subjectOptions"
          :placeholder="$t('page.hrm.salary.batch.form.subject')"
          filterable
          :disabled="operateType === 'edit'"
        />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.salary.batch.year')" path="batchYear">
        <NSelect v-model:value="model.batchYear" :options="yearOptions" filterable />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.salary.batch.month')" path="batchMonth">
        <NSelect v-model:value="model.batchMonth" :options="monthOptions" />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.salary.batch.batchType')" path="batchType">
        <NSelect
          v-model:value="model.batchType"
          :options="[{ label: $t('page.hrm.salary.batch.batchTypeEnum.monthly'), value: 'monthly' }]"
          :disabled="true"
        />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.salary.batch.remark')" path="remark">
        <NInput v-model:value="model.remark" type="textarea" :placeholder="$t('page.hrm.salary.batch.form.remark')" />
      </NFormItem>
    </NForm>
    <template #footer>
      <NSpace :size="16" justify="end">
        <NButton @click="closeModel">{{ $t('common.cancel') }}</NButton>
        <NButton type="primary" :loading="confirmLoading" @click="handleSubmit">{{ $t('common.confirm') }}</NButton>
      </NSpace>
    </template>
  </NModal>
</template>

<style scoped></style>
