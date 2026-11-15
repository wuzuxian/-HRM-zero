<script setup lang="ts">
import { reactive, ref, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { GetCertTemplateList, GetTrainingSessionList, BatchIssueCertificates } from '@/service/api';
import { $t } from '@/locales';

defineOptions({
  name: 'CertBatchIssueModal'
});

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

type Model = {
  sessionUuid: string;
  templateUuid: string;
};

const model: Model = reactive({ sessionUuid: '', templateUuid: '' });

const rules: Record<keyof Model, App.Global.FormRule> = {
  sessionUuid: defaultRequiredRule,
  templateUuid: defaultRequiredRule
};

interface Option {
  label: string;
  value: string;
}

const sessionOptions = ref<Option[]>([]);
const templateOptions = ref<Option[]>([]);

async function loadOptions() {
  const [sessionRes, templateRes] = await Promise.all([
    GetTrainingSessionList({ current: 1, size: 100, status: '3' }),
    GetCertTemplateList({ current: 1, size: 100, status: '1' })
  ]);
  if (!sessionRes.error && sessionRes.data) {
    sessionOptions.value = (sessionRes.data.records ?? []).map(s => ({
      label: `${s.sessionName}（${s.sessionNo}）`,
      value: s.uuid
    }));
  }
  if (!templateRes.error && templateRes.data) {
    templateOptions.value = (templateRes.data.records ?? []).map(tpl => ({
      label: `${tpl.templateName}（${tpl.noPrefix}）`,
      value: tpl.uuid
    }));
  }
}

function closeModel() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();
  confirmStartLoading();
  const { data, error } = await BatchIssueCertificates({
    sessionUuid: model.sessionUuid,
    templateUuid: model.templateUuid
  });
  confirmEndLoading();
  if (error) {
    return;
  }
  window.$message?.success(
    `${$t('page.hrm.training.certificate.batchIssueDone')}: ${data?.issuedCount ?? 0} / ${$t(
      'page.hrm.training.certificate.batchIssueSkipped'
    )}: ${data?.skippedCount ?? 0}`
  );
  closeModel();
  emit('submitted');
}

watch(visible, () => {
  if (visible.value) {
    model.sessionUuid = '';
    model.templateUuid = '';
    restoreValidation();
    loadOptions();
  }
});
</script>

<template>
  <NModal
    v-model:show="visible"
    :title="$t('page.hrm.training.certificate.batchIssue')"
    preset="card"
    class="w-560px"
  >
    <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="100">
      <NFormItemGi :label="$t('page.hrm.training.certificate.session')" path="sessionUuid">
        <NSelect
          v-model:value="model.sessionUuid"
          :options="sessionOptions"
          :placeholder="$t('page.hrm.training.certificate.form.session')"
          filterable
        />
      </NFormItemGi>
      <NFormItemGi :label="$t('page.hrm.training.certificate.template')" path="templateUuid">
        <NSelect
          v-model:value="model.templateUuid"
          :options="templateOptions"
          :placeholder="$t('page.hrm.training.certificate.form.template')"
          filterable
        />
      </NFormItemGi>
      <NAlert type="info" :show-icon="true" class="mt-8px">
        {{ $t('page.hrm.training.certificate.batchIssueTip') }}
      </NAlert>
    </NForm>
    <template #footer>
      <NSpace justify="end" :size="16">
        <NButton @click="closeModel">{{ $t('common.cancel') }}</NButton>
        <NButton type="primary" :loading="confirmLoading" @click="handleSubmit">{{ $t('common.confirm') }}</NButton>
      </NSpace>
    </template>
  </NModal>
</template>

<style scoped></style>
