<script setup lang="ts">
import { computed, reactive, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { hrmCertTemplateSourceOptions } from '@/constants/business';
import { AddCertTemplate, EditCertTemplate } from '@/service/api';
import { $t } from '@/locales';
import { translateOptions } from '@/utils/common';

defineOptions({
  name: 'CertTemplateOperateModal'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Hrm.CertTemplate | null;
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
    add: $t('page.hrm.training.certTemplate.addTemplate'),
    edit: $t('page.hrm.training.certTemplate.editTemplate')
  };
  return titles[props.operateType];
});

type Model = Api.Hrm.AddCertTemplateRequest;

const model: Model = reactive(createDefaultModel());

function createDefaultModel(): Model {
  return {
    templateName: '',
    certType: '',
    source: '1',
    noPrefix: '',
    remark: ''
  };
}

type RuleKey = Extract<keyof Model, 'templateName' | 'certType' | 'noPrefix'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  templateName: defaultRequiredRule,
  certType: defaultRequiredRule,
  noPrefix: defaultRequiredRule
};

function handleInitModel() {
  Object.assign(model, createDefaultModel());

  if (props.operateType === 'edit' && props.rowData) {
    Object.assign(model, {
      templateName: props.rowData.templateName,
      certType: props.rowData.certType,
      source: props.rowData.source,
      noPrefix: props.rowData.noPrefix,
      remark: props.rowData.remark
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
      ? AddCertTemplate({ ...model })
      : EditCertTemplate({ ...model, uuid: props.rowData?.uuid ?? '' });

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
        <NFormItemGi
          span="24 s:12 m:8"
          :label="$t('page.hrm.training.certTemplate.templateName')"
          path="templateName"
        >
          <NInput
            v-model:value="model.templateName"
            :placeholder="$t('page.hrm.training.certTemplate.form.templateName')"
          />
        </NFormItemGi>
        <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.training.certTemplate.certType')" path="certType">
          <NInput
            v-model:value="model.certType"
            :placeholder="$t('page.hrm.training.certTemplate.form.certType')"
          />
        </NFormItemGi>
        <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.training.certTemplate.noPrefix')" path="noPrefix">
          <NInput
            v-model:value="model.noPrefix"
            :placeholder="$t('page.hrm.training.certTemplate.form.noPrefix')"
          />
        </NFormItemGi>
        <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.training.certTemplate.sourceLabel')" path="source">
          <NSelect
            v-model:value="model.source"
            :options="translateOptions(hrmCertTemplateSourceOptions)"
            :placeholder="$t('page.hrm.training.certTemplate.form.source')"
          />
        </NFormItemGi>
        <NFormItemGi span="24" :label="$t('page.hrm.training.certTemplate.remark')" path="remark">
          <NInput v-model:value="model.remark" type="textarea" :rows="2" />
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
