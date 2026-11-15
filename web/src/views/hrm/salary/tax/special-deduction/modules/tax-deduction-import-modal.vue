<script setup lang="ts">
import { ref, watch } from 'vue';
import type { UploadFileInfo } from 'naive-ui';
import { useLoading } from '@sa/hooks';
import { ImportTaxSpecialDeduction } from '@/service/api';
import { $t } from '@/locales';

defineOptions({
  name: 'TaxDeductionImportModal'
});

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

async function handleSubmit() {
  const fileInfo = fileList.value[0];
  const file = fileInfo?.file;
  if (!file) {
    window.$message?.warning($t('page.hrm.salary.tax.deduction.chooseFileTip'));
    return;
  }

  confirmStartLoading();
  try {
    const content = await readFileAsBase64(file);
    const { error, data } = await ImportTaxSpecialDeduction({ fileName: file.name, content });
    if (!error) {
      window.$message?.success(`${$t('page.hrm.salary.tax.deduction.importSuccess')}: ${data?.successCount ?? 0}`);
      visible.value = false;
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
  <NModal
    v-model:show="visible"
    :title="$t('page.hrm.salary.tax.deduction.importModalTitle')"
    preset="card"
    class="w-480px"
  >
    <NUpload v-model:file-list="fileList" accept=".xlsx" :max-count="1" :default-upload="false" directory-dnd>
      <NUploadDragger>
        <div class="py-24px">
          <icon-mdi-file-excel-outline class="text-48px text-primary" />
          <div class="mt-12px">{{ $t('page.hrm.salary.tax.deduction.chooseFileTip') }}</div>
        </div>
      </NUploadDragger>
    </NUpload>
    <template #footer>
      <NSpace :size="16" justify="end">
        <NButton @click="visible = false">{{ $t('common.cancel') }}</NButton>
        <NButton type="primary" :loading="confirmLoading" @click="handleSubmit">
          {{ $t('common.confirm') }}
        </NButton>
      </NSpace>
    </template>
  </NModal>
</template>

<style scoped></style>
