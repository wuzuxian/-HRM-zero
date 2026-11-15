<script setup lang="ts">
import { computed, reactive, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { AddSalaryItem, EditSalaryItem } from '@/service/api';
import { $t } from '@/locales';
import {
  hrmSalaryItemClassOptions,
  hrmSalaryItemTypeOptions,
  hrmSalaryItemValueModeOptions,
  hrmSalaryStatusOptions
} from '@/constants/business';
import { translateOptions } from '@/utils/common';

defineOptions({
  name: 'SalaryItemOperateModal'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Hrm.SalaryItem | null;
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
    add: $t('page.hrm.salary.item.addItem'),
    edit: $t('page.hrm.salary.item.editItem')
  };
  return titles[props.operateType];
});

type Model = Pick<
  Api.Hrm.AddSalaryItemRequest,
  | 'itemCode'
  | 'itemName'
  | 'itemType'
  | 'itemClass'
  | 'valueMode'
  | 'calcOrder'
  | 'orderNum'
  | 'formula'
  | 'status'
  | 'remark'
>;

const model: Model = reactive(createDefaultModel());

function createDefaultModel(): Model {
  return {
    itemCode: '',
    itemName: '',
    itemType: 'fixed',
    itemClass: 'earning',
    valueMode: 'manual',
    calcOrder: 0,
    orderNum: 0,
    formula: '',
    status: '1',
    remark: ''
  };
}

type RuleKey = Extract<keyof Model, 'itemCode' | 'itemName' | 'itemType' | 'itemClass' | 'valueMode'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  itemCode: defaultRequiredRule,
  itemName: defaultRequiredRule,
  itemType: defaultRequiredRule,
  itemClass: defaultRequiredRule,
  valueMode: defaultRequiredRule
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

  const request =
    props.operateType === 'add'
      ? AddSalaryItem({ ...model })
      : EditSalaryItem({ ...model, uuid: props.rowData?.uuid ?? '' });

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
    <NScrollbar class="h-408px pr-20px">
      <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="120">
        <NFormItem :label="$t('page.hrm.salary.item.itemName')" path="itemName">
          <NInput v-model:value="model.itemName" :placeholder="$t('page.hrm.salary.item.form.itemName')" />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.item.itemCode')" path="itemCode">
          <NInput
            v-model:value="model.itemCode"
            :placeholder="$t('page.hrm.salary.item.form.itemCode')"
            :disabled="props.rowData?.isPreset === 1"
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.item.itemType.title')" path="itemType">
          <NSelect
            v-model:value="model.itemType"
            :placeholder="$t('page.hrm.salary.item.form.itemType')"
            :options="translateOptions(hrmSalaryItemTypeOptions)"
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.item.itemClass.title')" path="itemClass">
          <NSelect
            v-model:value="model.itemClass"
            :placeholder="$t('page.hrm.salary.item.form.itemClass')"
            :options="translateOptions(hrmSalaryItemClassOptions)"
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.item.valueMode.title')" path="valueMode">
          <NSelect
            v-model:value="model.valueMode"
            :placeholder="$t('page.hrm.salary.item.form.valueMode')"
            :options="translateOptions(hrmSalaryItemValueModeOptions)"
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.item.calcOrder')" path="calcOrder">
          <NInputNumber v-model:value="model.calcOrder" :min="0" class="w-full" />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.item.orderNum')" path="orderNum">
          <NInputNumber v-model:value="model.orderNum" :min="0" class="w-full" />
        </NFormItem>
        <NFormItem v-if="model.itemType === 'formula'" :label="$t('page.hrm.salary.item.formula')" path="formula">
          <NInput
            v-model:value="model.formula"
            type="textarea"
            :placeholder="$t('page.hrm.salary.item.form.formula')"
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.item.status')" path="status">
          <NRadioGroup v-model:value="model.status">
            <NRadio
              v-for="item in hrmSalaryStatusOptions"
              :key="item.value"
              :value="item.value"
              :label="$t(item.label)"
            />
          </NRadioGroup>
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.item.remark')" path="remark">
          <NInput v-model:value="model.remark" type="textarea" :placeholder="$t('page.hrm.salary.item.form.remark')" />
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
