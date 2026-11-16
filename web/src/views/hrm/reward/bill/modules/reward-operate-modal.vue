<script setup lang="ts">
import { computed, reactive, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { AddReward, EditReward } from '@/service/api';
import { $t } from '@/locales';
import { hrmRewardLevelOptions, hrmRewardTypeOptions } from '@/constants/business';
import { translateOptions } from '@/utils/common';

defineOptions({
  name: 'RewardOperateModal'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Hrm.HrmReward | null;
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
    add: $t('page.hrm.reward.addReward'),
    edit: $t('page.hrm.reward.editReward')
  };
  return titles[props.operateType];
});

type Model = Omit<Api.Hrm.AddRewardRequest, 'occurDate'> & {
  // NDatePicker 的 formatted-value 不能是空串（解析 "" 抛 Invalid time value），未选择时必须为 null
  occurDate: string | null;
};

const model: Model = reactive(createDefaultModel());

function createDefaultModel(): Model {
  return {
    employeeUuid: '',
    employeeNo: '',
    employeeName: '',
    rewardType: 'reward',
    rewardLevel: 'company',
    docNo: '',
    occurDate: null,
    amount: 0,
    reason: '',
    salaryMonth: ''
  };
}

type RuleKey = Extract<keyof Model, 'employeeNo' | 'employeeName' | 'rewardType' | 'rewardLevel' | 'occurDate'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  employeeNo: defaultRequiredRule,
  employeeName: defaultRequiredRule,
  rewardType: defaultRequiredRule,
  rewardLevel: defaultRequiredRule,
  occurDate: defaultRequiredRule
};

function handleInitModel() {
  Object.assign(model, createDefaultModel());

  if (props.operateType === 'edit' && props.rowData) {
    Object.assign(model, props.rowData);
    // DB 列 NOT NULL DEFAULT ''，编辑回填可能拿到空串；空串会让 NDatePicker 抛 Invalid time value
    model.occurDate = props.rowData.occurDate || null;
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
      ? AddReward({ ...model, occurDate: model.occurDate ?? '' })
      : EditReward({ ...model, occurDate: model.occurDate ?? '', uuid: props.rowData?.uuid ?? '' });

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
        <NDivider :title="$t('page.hrm.reward.sectionEmployee')" title-placement="left" />
        <NGrid responsive="screen" item-responsive>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.reward.employeeNo')" path="employeeNo">
            <NInput v-model:value="model.employeeNo" :placeholder="$t('page.hrm.reward.form.employeeNo')" />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.reward.employeeName')" path="employeeName">
            <NInput v-model:value="model.employeeName" :placeholder="$t('page.hrm.reward.form.employeeName')" />
          </NFormItemGi>
        </NGrid>

        <NDivider :title="$t('page.hrm.reward.sectionReward')" title-placement="left" />
        <NGrid responsive="screen" item-responsive>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.reward.rewardType.title')" path="rewardType">
            <NSelect
              v-model:value="model.rewardType"
              :options="translateOptions(hrmRewardTypeOptions)"
              :placeholder="$t('page.hrm.reward.form.rewardType')"
            />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.reward.rewardLevel.title')" path="rewardLevel">
            <NSelect
              v-model:value="model.rewardLevel"
              :options="translateOptions(hrmRewardLevelOptions)"
              :placeholder="$t('page.hrm.reward.form.rewardLevel')"
            />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.reward.docNo')" path="docNo">
            <NInput v-model:value="model.docNo" :placeholder="$t('page.hrm.reward.form.docNo')" />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.reward.occurDate')" path="occurDate">
            <NDatePicker
              v-model:formatted-value="model.occurDate"
              type="date"
              value-format="yyyy-MM-dd"
              class="w-full"
            />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.reward.amount')" path="amount">
            <NInputNumber v-model:value="model.amount" :min="0" :precision="2" class="w-full" />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.reward.salaryMonth')" path="salaryMonth">
            <NInput v-model:value="model.salaryMonth" :placeholder="$t('page.hrm.reward.form.salaryMonth')" />
          </NFormItemGi>
          <NFormItemGi span="24" :label="$t('page.hrm.reward.reason')" path="reason">
            <NInput
              v-model:value="model.reason"
              type="textarea"
              :rows="2"
              :placeholder="$t('page.hrm.reward.form.reason')"
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
