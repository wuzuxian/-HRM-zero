<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import type { UploadFileInfo } from 'naive-ui';
import { ImportScheduleBoard } from '@/service/api';
import { $t } from '@/locales';

defineOptions({
  name: 'ScheduleImportModal'
});

const props = defineProps<{
  visible: boolean;
  groupUuid: string;
  month: string;
}>();

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void;
  (e: 'submitted'): void;
}>();

const visible = computed({
  get: () => props.visible,
  set: value => emit('update:visible', value)
});

const fileName = ref('');
const fileContent = ref('');
const importing = ref(false);

watch(
  () => props.visible,
  value => {
    if (value) {
      fileName.value = '';
      fileContent.value = '';
    }
  }
);

function handleUploadChange(options: { file: UploadFileInfo }) {
  if (options.file.file) {
    const file = options.file.file;
    fileName.value = file.name;
    const reader = new FileReader();
    reader.onload = () => {
      // xlsx 是二进制 → base64
      const bytes = new Uint8Array(reader.result as ArrayBuffer);
      let binary = '';
      for (let i = 0; i < bytes.length; i += 1) {
        binary += String.fromCharCode(bytes[i]!);
      }
      fileContent.value = window.btoa(binary);
    };
    reader.readAsArrayBuffer(file);
  }
}

async function handleSubmit() {
  if (!fileContent.value) {
    window.$message?.error($t('page.hrm.attendance.scheduleBoard.importFileRequired'));
    return;
  }
  importing.value = true;
  const { data, error } = await ImportScheduleBoard({
    groupUuid: props.groupUuid,
    month: props.month,
    fileName: fileName.value,
    fileContent: fileContent.value
  });
  importing.value = false;
  if (error) return;
  window.$message?.success($t('page.hrm.attendance.scheduleBoard.importSuccess', { n: data?.successCount ?? 0 }));
  visible.value = false;
  emit('submitted');
}
</script>

<template>
  <NModal v-model:show="visible" :title="$t('page.hrm.attendance.scheduleBoard.import')" preset="card" class="w-520px">
    <NUpload :max="1" accept=".xlsx" :default-upload="false" @change="handleUploadChange">
      <NUploadDragger>
        <div class="py-16px whitespace-pre-line">{{ $t('page.hrm.attendance.scheduleBoard.importHint') }}</div>
      </NUploadDragger>
    </NUpload>
    <template #footer>
      <NSpace justify="end">
        <NButton @click="visible = false">{{ $t('common.cancel') }}</NButton>
        <NButton type="primary" :loading="importing" @click="handleSubmit">{{ $t('common.confirm') }}</NButton>
      </NSpace>
    </template>
  </NModal>
</template>
