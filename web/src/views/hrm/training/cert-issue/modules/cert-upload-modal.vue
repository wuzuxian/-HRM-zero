<script setup lang="ts">
import { reactive, ref, watch } from 'vue';
import type { UploadFileInfo } from 'naive-ui';
import { useLoading } from '@sa/hooks';
import { UploadCertificateFile } from '@/service/api';
import { $t } from '@/locales';

defineOptions({
  name: 'CertUploadModal'
});

interface Props {
  /** certificate uuid to upload/replace the scan file for */
  certUuid: string;
}

const props = defineProps<Props>();

interface Emits {
  (e: 'submitted'): void;
}

const emit = defineEmits<Emits>();

const visible = defineModel<boolean>('visible', {
  default: false
});

const { loading: confirmLoading, startLoading: confirmStartLoading, endLoading: confirmEndLoading } = useLoading();

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
  const file = fileList.value[0]?.file;
  if (!file) {
    window.$message?.warning($t('page.hrm.training.certificate.chooseFile'));
    return;
  }
  confirmStartLoading();
  try {
    const content = await readFileAsBase64(file);
    const { error } = await UploadCertificateFile({ uuid: props.certUuid, fileName: file.name, fileContent: content });
    if (!error) {
      window.$message?.success($t('common.updateSuccess'));
      closeModel();
      emit('submitted');
    }
  } finally {
    confirmEndLoading();
  }
}

watch(visible, () => {
  if (visible.value) {
    fileList.value = [];
  }
});
</script>

<template>
  <NModal v-model:show="visible" :title="$t('page.hrm.training.certificate.upload')" preset="card" class="w-480px">
    <NUpload v-model:file-list="fileList" :max="1" :default-upload="false" accept=".pdf,.png,.jpg,.jpeg">
      <NButton>{{ $t('page.hrm.training.certificate.chooseFile') }}</NButton>
    </NUpload>
    <template #footer>
      <NSpace justify="end" :size="16">
        <NButton @click="closeModel">{{ $t('common.cancel') }}</NButton>
        <NButton type="primary" :loading="confirmLoading" @click="handleSubmit">{{ $t('common.save') }}</NButton>
      </NSpace>
    </template>
  </NModal>
</template>

<style scoped></style>
