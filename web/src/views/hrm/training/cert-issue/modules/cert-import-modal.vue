<script setup lang="ts">
import { reactive, ref, watch } from 'vue';
import type { UploadFileInfo } from 'naive-ui';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { ImportExternalCertificate, GetUserList } from '@/service/api';
import { $t } from '@/locales';

defineOptions({
  name: 'CertImportModal'
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

type Model = Api.Hrm.ImportExternalCertificateRequest;

const model: Model = reactive(createDefaultModel());

function createDefaultModel(): Model {
  return {
    userUuid: '',
    certName: '',
    certType: '',
    certNo: ''
  };
}

const rules: Partial<Record<keyof Model, App.Global.FormRule>> = {
  userUuid: defaultRequiredRule,
  certName: defaultRequiredRule,
  certType: defaultRequiredRule,
  certNo: defaultRequiredRule
};

// ---- 学员远程搜索（与培训班次选人同一模式） ----
interface UserOption {
  label: string;
  value: string;
}

const userOptions = ref<UserOption[]>([]);
const userSearching = ref(false);

async function searchUser(query: string) {
  if (!query) {
    userOptions.value = [];
    return;
  }
  userSearching.value = true;
  const { data, error } = await GetUserList({ current: 1, size: 20, username: query });
  userSearching.value = false;
  if (!error && data) {
    userOptions.value = (data.records ?? []).map(u => ({
      label: u.nickName ? `${u.nickName}（${u.username}）` : u.username,
      value: u.uuid
    }));
  }
}

// ---- 扫描件（可选） ----
const fileList = ref<UploadFileInfo[]>([]);

function readFileAsBase64(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => {
      const buffer = reader.result as ArrayBuffer;
      const bytes = new Uint8Array(buffer);
      let binary = '';
      const chunkSize = 0x8000;
      for (let i = 0; i < bytes.length; i += chunkSize) {
        binary += String.fromCharCode(...bytes.subarray(i, i + chunkSize));
      }
      resolve(btoa(binary));
    };
    reader.onerror = () => reject(new Error('read file failed'));
    reader.readAsArrayBuffer(file);
  });
}

function closeModel() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();
  confirmStartLoading();
  try {
    let fileName = '';
    let fileContent = '';
    const file = fileList.value[0]?.file;
    if (file) {
      fileName = file.name;
      fileContent = await readFileAsBase64(file);
    }
    const { error } = await ImportExternalCertificate({ ...model, fileName, fileContent });
    if (!error) {
      window.$message?.success($t('common.addSuccess'));
      closeModel();
      emit('submitted');
    }
  } finally {
    confirmEndLoading();
  }
}

watch(visible, () => {
  if (visible.value) {
    Object.assign(model, createDefaultModel());
    fileList.value = [];
    userOptions.value = [];
    restoreValidation();
  }
});
</script>

<template>
  <NModal
    v-model:show="visible"
    :title="$t('page.hrm.training.certificate.importExternal')"
    preset="card"
    class="w-560px"
  >
    <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="100">
      <NFormItem :label="$t('page.hrm.training.certificate.student')" path="userUuid">
        <NSelect
          v-model:value="model.userUuid"
          :options="userOptions"
          :placeholder="$t('page.hrm.training.certificate.form.student')"
          filterable
          remote
          :loading="userSearching"
          @search="searchUser"
        />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.training.certName')" path="certName">
        <NInput v-model:value="model.certName" :placeholder="$t('page.hrm.training.certificate.form.certName')" />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.training.certType')" path="certType">
        <NInput v-model:value="model.certType" :placeholder="$t('page.hrm.training.certificate.form.certType')" />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.training.certNo')" path="certNo">
        <NInput v-model:value="model.certNo" :placeholder="$t('page.hrm.training.certificate.form.certNo')" />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.training.scanFile')" path="file">
        <NUpload v-model:file-list="fileList" :max="1" :default-upload="false" accept=".pdf,.png,.jpg,.jpeg">
          <NButton>{{ $t('page.hrm.training.certificate.chooseFile') }}</NButton>
        </NUpload>
      </NFormItem>
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
