<script setup lang="ts">
import { computed, reactive, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { AddCandidate, EditCandidate } from '@/service/api';
import { $t } from '@/locales';
import { hrmCandidateEducationOptions, hrmCandidateSourceChannelOptions } from '@/constants/business';
import { translateOptions } from '@/utils/common';

defineOptions({
  name: 'CandidateOperateModal'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Hrm.HrmCandidate | null;
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
    add: $t('page.hrm.candidate.addCandidate'),
    edit: $t('page.hrm.candidate.editCandidate')
  };
  return titles[props.operateType];
});

type Model = Api.Hrm.AddCandidateRequest;

const model: Model = reactive(createDefaultModel());

function createDefaultModel(): Model {
  return {
    name: '',
    phone: '',
    email: '',
    sourceChannel: 'other',
    education: null,
    currentCompany: '',
    currentPosition: '',
    expectSalary: '',
    remark: ''
  };
}

type RuleKey = Extract<keyof Model, 'name'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  name: defaultRequiredRule
};

function handleInitModel() {
  Object.assign(model, createDefaultModel());

  if (props.operateType === 'edit' && props.rowData) {
    Object.assign(model, {
      name: props.rowData.name,
      phone: props.rowData.phone,
      email: props.rowData.email,
      sourceChannel: props.rowData.sourceChannel,
      education: props.rowData.education,
      currentCompany: props.rowData.currentCompany,
      currentPosition: props.rowData.currentPosition,
      expectSalary: props.rowData.expectSalary
    });
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
      ? AddCandidate({ ...model })
      : EditCandidate({ ...model, uuid: props.rowData?.uuid ?? '' });

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
        <NGrid responsive="screen" item-responsive>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.candidate.name')" path="name">
            <NInput v-model:value="model.name" :placeholder="$t('page.hrm.candidate.form.name')" />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.candidate.phone')" path="phone">
            <NInput v-model:value="model.phone" :placeholder="$t('page.hrm.candidate.form.phone')" />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.candidate.email')" path="email">
            <NInput v-model:value="model.email" :placeholder="$t('page.hrm.candidate.form.email')" />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.candidate.sourceChannel.title')" path="sourceChannel">
            <NSelect
              v-model:value="model.sourceChannel"
              :options="translateOptions(hrmCandidateSourceChannelOptions)"
              :placeholder="$t('page.hrm.candidate.form.sourceChannel')"
            />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.candidate.education.title')" path="education">
            <NSelect
              v-model:value="model.education"
              :options="translateOptions(hrmCandidateEducationOptions)"
              :placeholder="$t('page.hrm.candidate.form.education')"
              clearable
            />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.candidate.expectSalary')" path="expectSalary">
            <NInput v-model:value="model.expectSalary" :placeholder="$t('page.hrm.candidate.form.expectSalary')" />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.candidate.currentCompany')" path="currentCompany">
            <NInput v-model:value="model.currentCompany" :placeholder="$t('page.hrm.candidate.form.currentCompany')" />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.candidate.currentPosition')" path="currentPosition">
            <NInput
              v-model:value="model.currentPosition"
              :placeholder="$t('page.hrm.candidate.form.currentPosition')"
            />
          </NFormItemGi>
          <NFormItemGi span="24" :label="$t('page.hrm.candidate.remark')" path="remark">
            <NInput
  type="textarea" v-model:value="model.remark" :rows="2" :placeholder="$t('page.hrm.candidate.form.remark')" />
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
