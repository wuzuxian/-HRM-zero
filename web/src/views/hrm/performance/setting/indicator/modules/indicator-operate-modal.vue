<script setup lang="ts">
import { computed, reactive, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { AddPerfIndicator, EditPerfIndicator } from '@/service/api';
import { $t } from '@/locales';
import {
  perfDataSourceOptions,
  perfIndicatorTypeOptions,
  perfScoreRuleOptions,
  perfStatusOptions
} from '@/constants/hrm-performance';
import { translateOptions } from '@/utils/common';

defineOptions({
  name: 'PerfIndicatorOperateModal'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Hrm.Performance.Indicator | null;
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
    add: $t('page.hrm.performance.indicator.addIndicator'),
    edit: $t('page.hrm.performance.indicator.editIndicator')
  };
  return titles[props.operateType];
});

type Model = Pick<
  Api.Hrm.Performance.Indicator,
  'code' | 'name' | 'indicatorType' | 'unit' | 'scoreRule' | 'dataSource' | 'defaultWeight' | 'status' | 'remark'
>;

const model: Model = reactive(createDefaultModel());

function createDefaultModel(): Model {
  return {
    code: '',
    name: '',
    indicatorType: 'kpi',
    unit: '',
    scoreRule: 'manual',
    dataSource: 'manual',
    defaultWeight: 0,
    status: '1',
    remark: ''
  };
}

type RuleKey = Extract<keyof Model, 'code' | 'name' | 'indicatorType' | 'scoreRule' | 'dataSource'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  code: defaultRequiredRule,
  name: defaultRequiredRule,
  indicatorType: defaultRequiredRule,
  scoreRule: defaultRequiredRule,
  dataSource: defaultRequiredRule
};

function handleInitModel() {
  Object.assign(model, createDefaultModel());

  if (props.operateType === 'edit' && props.rowData) {
    Object.assign(model, props.rowData);
  }
}

function closeModel() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();
  confirmStartLoading();
  const { error } =
    props.operateType === 'add'
      ? await AddPerfIndicator({ ...model, status: model.status ?? '1' })
      : await EditPerfIndicator({
          uuid: props.rowData?.uuid,
          ...model,
          status: model.status ?? '1'
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
  }
});
</script>

<template>
  <NModal v-model:show="visible" :title="title" preset="card" class="w-600px">
    <NScrollbar class="h-408px pr-20px">
      <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="100">
        <NFormItem :label="$t('page.hrm.performance.indicator.code')" path="code">
          <NInput
            v-model:value="model.code"
            :placeholder="$t('page.hrm.performance.indicator.form.code')"
            :disabled="props.operateType === 'edit'"
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.performance.indicator.name')" path="name">
          <NInput v-model:value="model.name" :placeholder="$t('page.hrm.performance.indicator.form.name')" />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.performance.indicator.indicatorType')" path="indicatorType">
          <NSelect
            v-model:value="model.indicatorType"
            :placeholder="$t('page.hrm.performance.indicator.form.indicatorType')"
            :options="translateOptions(perfIndicatorTypeOptions)"
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.performance.indicator.unit')" path="unit">
          <NInput v-model:value="model.unit" :placeholder="$t('page.hrm.performance.indicator.form.unit')" />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.performance.indicator.scoreRule')" path="scoreRule">
          <NSelect
            v-model:value="model.scoreRule"
            :placeholder="$t('page.hrm.performance.indicator.form.scoreRule')"
            :options="translateOptions(perfScoreRuleOptions)"
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.performance.indicator.dataSource')" path="dataSource">
          <NSelect
            v-model:value="model.dataSource"
            :placeholder="$t('page.hrm.performance.indicator.form.dataSource')"
            :options="translateOptions(perfDataSourceOptions)"
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.performance.indicator.defaultWeight')" path="defaultWeight">
          <NInputNumber v-model:value="model.defaultWeight" :min="0" :max="100" :precision="2" class="w-full">
            <template #suffix>%</template>
          </NInputNumber>
        </NFormItem>
        <NFormItem :label="$t('page.hrm.performance.indicator.status')" path="status">
          <NRadioGroup v-model:value="model.status">
            <NRadio
              v-for="item in translateOptions(perfStatusOptions)"
              :key="item.value"
              :value="item.value"
              :label="item.label"
            />
          </NRadioGroup>
        </NFormItem>
        <NFormItem :label="$t('page.hrm.performance.indicator.remark')" path="remark">
          <NInput v-model:value="model.remark" :placeholder="$t('page.hrm.performance.indicator.form.remark')" />
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
