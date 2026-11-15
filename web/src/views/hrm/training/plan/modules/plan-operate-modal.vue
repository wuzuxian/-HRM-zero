<script setup lang="ts">
import { computed, reactive, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { AddTrainingPlan, EditTrainingPlan } from '@/service/api';
import { $t } from '@/locales';
import { hrmTrainingPlanTypeOptions } from '@/constants/business';
import { translateOptions } from '@/utils/common';

defineOptions({
  name: 'PlanOperateModal'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Hrm.TrainingPlan | null;
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
    add: $t('page.hrm.training.plan.addPlan'),
    edit: $t('page.hrm.training.plan.editPlan')
  };
  return titles[props.operateType];
});

type Model = Api.Hrm.AddTrainingPlanRequest;

const model: Model = reactive(createDefaultModel());

function createDefaultModel(): Model {
  return {
    planName: '',
    theme: '',
    planYear: String(new Date().getFullYear()),
    planType: 'special',
    startDate: '',
    endDate: ''
  };
}

type RuleKey = Extract<keyof Model, 'planName' | 'planYear' | 'planType' | 'startDate' | 'endDate'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  planName: defaultRequiredRule,
  planYear: defaultRequiredRule,
  planType: defaultRequiredRule,
  startDate: defaultRequiredRule,
  endDate: defaultRequiredRule
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
  if (model.endDate < model.startDate) {
    window.$message?.error($t('page.hrm.training.plan.form.endDateBeforeStart'));
    return;
  }
  confirmStartLoading();

  const request =
    props.operateType === 'add'
      ? AddTrainingPlan({ ...model })
      : EditTrainingPlan({ ...model, uuid: props.rowData?.uuid ?? '' });

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
  <NModal v-model:show="visible" :title="title" preset="card" class="w-680px">
    <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="100">
      <NGrid responsive="screen" item-responsive>
        <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.training.plan.planName')" path="planName">
          <NInput v-model:value="model.planName" :placeholder="$t('page.hrm.training.plan.form.planName')" />
        </NFormItemGi>
        <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.training.plan.theme')" path="theme">
          <NInput v-model:value="model.theme" :placeholder="$t('page.hrm.training.plan.form.theme')" />
        </NFormItemGi>
        <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.training.plan.planYear')" path="planYear">
          <NInput v-model:value="model.planYear" :placeholder="$t('page.hrm.training.plan.form.planYear')" />
        </NFormItemGi>
        <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.training.plan.planType')" path="planType">
          <NSelect
            v-model:value="model.planType"
            :options="translateOptions(hrmTrainingPlanTypeOptions)"
            :placeholder="$t('page.hrm.training.plan.form.planType')"
          />
        </NFormItemGi>
        <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.training.plan.startDate')" path="startDate">
          <NDatePicker
            v-model:formatted-value="model.startDate"
            type="date"
            value-format="yyyy-MM-dd"
            class="w-full"
          />
        </NFormItemGi>
        <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.training.plan.endDate')" path="endDate">
          <NDatePicker
            v-model:formatted-value="model.endDate"
            type="date"
            value-format="yyyy-MM-dd"
            class="w-full"
          />
        </NFormItemGi>
      </NGrid>
    </NForm>
    <template #footer>
      <NSpace justify="end" :size="16">
        <NButton @click="closeModel">{{ $t('common.cancel') }}</NButton>
        <NButton type="primary" :loading="confirmLoading" @click="handleSubmit">{{ $t('common.save') }}</NButton>
      </NSpace>
    </template>
  </NModal>
</template>

<style scoped></style>
