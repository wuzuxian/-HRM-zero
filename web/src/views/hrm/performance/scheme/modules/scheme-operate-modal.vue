<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { AddPerfScheme, EditPerfScheme, GetPerfTemplateList } from '@/service/api';
import { $t } from '@/locales';
import {
  perfSchemeAssessmentModeOptions,
  perfSchemePayLinkModeOptions,
  perfSchemePeriodTypeOptions,
  perfSchemeScoreModeOptions,
  perfStatusOptions
} from '@/constants/hrm-performance';
import { translateOptions } from '@/utils/common';

defineOptions({
  name: 'PerfSchemeOperateModal'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Hrm.Performance.Scheme | null;
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
    add: $t('page.hrm.performance.scheme.addScheme'),
    edit: $t('page.hrm.performance.scheme.editScheme')
  };
  return titles[props.operateType];
});

type Model = {
  name: string;
  templateUuid: string | null;
  periodType: Api.Hrm.Performance.SchemePeriodType;
  assessmentMode: Api.Hrm.Performance.SchemeAssessmentMode;
  scoreMode: Api.Hrm.Performance.SchemeScoreMode;
  payLinkMode: Api.Hrm.Performance.SchemePayLinkMode;
  status: Api.Hrm.Performance.CommonStatus;
  remark: string;
};

const model: Model = reactive(createDefaultModel());

function createDefaultModel(): Model {
  return {
    name: '',
    templateUuid: null,
    periodType: 'month',
    assessmentMode: 'kpi',
    scoreMode: 'percent',
    payLinkMode: 'coefficient',
    status: '1',
    remark: ''
  };
}

type RuleKey = Extract<keyof Model, 'name' | 'templateUuid'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  name: defaultRequiredRule,
  templateUuid: defaultRequiredRule
};

/** 考核模板选项（关联模板，来自 HRM-40 模板管理，取启用模板全量） */
const templateOptions = ref<CommonType.Option<string>[]>([]);

async function getTemplateOptions() {
  const { error, data } = await GetPerfTemplateList({ current: 1, size: 1000 });
  if (!error && data) {
    templateOptions.value = data.records
      .filter(item => item.status === '1')
      .map(item => ({
        label: item.name,
        value: item.uuid
      }));
  }
}

function handleInitModel() {
  Object.assign(model, createDefaultModel());

  if (props.operateType === 'edit' && props.rowData?.uuid) {
    model.name = props.rowData.name;
    model.templateUuid = props.rowData.templateUuid;
    model.periodType = props.rowData.periodType;
    model.assessmentMode = props.rowData.assessmentMode;
    model.scoreMode = props.rowData.scoreMode;
    model.payLinkMode = props.rowData.payLinkMode;
    model.status = (props.rowData.status ?? '1') as Api.Hrm.Performance.CommonStatus;
    model.remark = props.rowData.remark;
  }
}

function closeModel() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();

  if (!model.templateUuid) {
    window.$message?.error($t('page.hrm.performance.scheme.form.templateUuid'));
    return;
  }

  confirmStartLoading();
  const { error } =
    props.operateType === 'add'
      ? await AddPerfScheme({
          name: model.name,
          templateUuid: model.templateUuid,
          periodType: model.periodType,
          assessmentMode: model.assessmentMode,
          scoreMode: model.scoreMode,
          payLinkMode: model.payLinkMode,
          status: model.status,
          remark: model.remark
        })
      : await EditPerfScheme({
          uuid: props.rowData?.uuid,
          name: model.name,
          templateUuid: model.templateUuid,
          periodType: model.periodType,
          assessmentMode: model.assessmentMode,
          scoreMode: model.scoreMode,
          payLinkMode: model.payLinkMode,
          status: model.status,
          remark: model.remark
        });
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
    getTemplateOptions();
  }
});
</script>

<template>
  <NModal v-model:show="visible" :title="title" preset="card" class="w-600px">
    <NScrollbar class="h-480px pr-20px">
      <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="100">
        <NFormItem :label="$t('page.hrm.performance.scheme.name')" path="name">
          <NInput v-model:value="model.name" :placeholder="$t('page.hrm.performance.scheme.form.name')" />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.performance.scheme.templateName')" path="templateUuid">
          <NSelect
            v-model:value="model.templateUuid"
            :placeholder="$t('page.hrm.performance.scheme.form.templateUuid')"
            :options="templateOptions"
            filterable
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.performance.scheme.periodType')" path="periodType">
          <NSelect
            v-model:value="model.periodType"
            :placeholder="$t('page.hrm.performance.scheme.form.periodType')"
            :options="translateOptions(perfSchemePeriodTypeOptions)"
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.performance.scheme.assessmentMode')" path="assessmentMode">
          <NRadioGroup v-model:value="model.assessmentMode">
            <NRadio
              v-for="item in translateOptions(perfSchemeAssessmentModeOptions)"
              :key="item.value"
              :value="item.value"
              :label="item.label"
            />
          </NRadioGroup>
        </NFormItem>
        <NFormItem :label="$t('page.hrm.performance.scheme.scoreMode')" path="scoreMode">
          <NRadioGroup v-model:value="model.scoreMode">
            <NRadio
              v-for="item in translateOptions(perfSchemeScoreModeOptions)"
              :key="item.value"
              :value="item.value"
              :label="item.label"
            />
          </NRadioGroup>
        </NFormItem>
        <NFormItem :label="$t('page.hrm.performance.scheme.payLinkMode')" path="payLinkMode">
          <NRadioGroup v-model:value="model.payLinkMode">
            <NRadio
              v-for="item in translateOptions(perfSchemePayLinkModeOptions)"
              :key="item.value"
              :value="item.value"
              :label="item.label"
            />
          </NRadioGroup>
        </NFormItem>
        <NFormItem :label="$t('page.hrm.performance.scheme.status')" path="status">
          <NRadioGroup v-model:value="model.status">
            <NRadio
              v-for="item in translateOptions(perfStatusOptions)"
              :key="item.value"
              :value="item.value"
              :label="item.label"
            />
          </NRadioGroup>
        </NFormItem>
        <NFormItem :label="$t('page.hrm.performance.scheme.remark')" path="remark">
          <NInput v-model:value="model.remark" :placeholder="$t('page.hrm.performance.scheme.form.remark')" />
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
