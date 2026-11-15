<script setup lang="tsx">
import { computed, ref } from 'vue';
import { NButton, NTag } from 'naive-ui';
import { DownloadMyCertificateFile, GetMyTrainingCertificateList, GetMyTrainingSessionList } from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { hrmTrainingCertificateStatusRecord, hrmTrainingSessionStatusRecord } from '@/constants/business';
import { useAuth } from '@/hooks/business/auth';
import { useTable } from '@/hooks/common/table';

const appStore = useAppStore();
const { hasAuth } = useAuth();

const canDownload = computed(() => hasAuth('hrm:training:downloadMyCertificateFile'));

// ---- 我的班次 ----
const {
  columns: sessionColumns,
  columnChecks: sessionColumnChecks,
  data: sessionData,
  getData: getSessionData,
  loading: sessionLoading,
  mobilePagination: sessionPagination
} = useTable({
  apiFn: GetMyTrainingSessionList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10
  },
  columns: () => [
    {
      key: 'index',
      title: $t('common.index'),
      align: 'center',
      width: 64
    },
    {
      key: 'sessionNo',
      title: $t('page.hrm.training.sessionNo'),
      align: 'center',
      minWidth: 130
    },
    {
      key: 'sessionName',
      title: $t('page.hrm.training.sessionName'),
      align: 'center',
      minWidth: 140
    },
    {
      key: 'location',
      title: $t('page.hrm.training.location'),
      align: 'center',
      minWidth: 120
    },
    {
      key: 'startTime',
      title: $t('page.hrm.training.startTime'),
      align: 'center',
      minWidth: 150
    },
    {
      key: 'endTime',
      title: $t('page.hrm.training.endTime'),
      align: 'center',
      minWidth: 150
    },
    {
      key: 'hours',
      title: $t('page.hrm.training.hours'),
      align: 'center',
      width: 80,
      render: row => `${row.hours} h`
    },
    {
      key: 'status',
      title: $t('page.hrm.training.sessionStatusLabel'),
      align: 'center',
      width: 90,
      render: row => {
        if (row.status === null) {
          return null;
        }
        const tagMap: Record<Api.Hrm.TrainingSessionStatus, NaiveUI.ThemeColor> = {
          '1': 'info',
          '2': 'warning',
          '3': 'success'
        };
        return <NTag type={tagMap[row.status]}>{$t(hrmTrainingSessionStatusRecord[row.status])}</NTag>;
      }
    }
  ]
});

// ---- 我的培训证书 ----
const {
  columns: certColumns,
  columnChecks: certColumnChecks,
  data: certData,
  getData: getCertData,
  loading: certLoading,
  mobilePagination: certPagination
} = useTable({
  apiFn: GetMyTrainingCertificateList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10
  },
  columns: () => [
    {
      key: 'index',
      title: $t('common.index'),
      align: 'center',
      width: 64
    },
    {
      key: 'certName',
      title: $t('page.hrm.training.certName'),
      align: 'center',
      minWidth: 140
    },
    {
      key: 'certNo',
      title: $t('page.hrm.training.certNo'),
      align: 'center',
      minWidth: 140
    },
    {
      key: 'issueTime',
      title: $t('page.hrm.training.issueTime'),
      align: 'center',
      minWidth: 150,
      render: row => row.issueTime || '-'
    },
    {
      key: 'status',
      title: $t('page.hrm.training.certStatusLabel'),
      align: 'center',
      width: 90,
      render: row => {
        if (row.status === null) {
          return null;
        }
        const tagMap: Record<Api.Hrm.TrainingCertificateStatus, NaiveUI.ThemeColor> = {
          '0': 'warning',
          '1': 'success',
          '2': 'error'
        };
        return <NTag type={tagMap[row.status]}>{$t(hrmTrainingCertificateStatusRecord[row.status])}</NTag>;
      }
    },
    {
      key: 'operate',
      title: $t('page.hrm.training.scanFile'),
      align: 'center',
      width: 160,
      render: row =>
        row.hasFile && canDownload.value ? (
          <div class="flex-center gap-8px">
            <NButton type="primary" ghost size="small" onClick={() => handlePreview(row)}>
              {$t('page.hrm.training.preview')}
            </NButton>
            <NButton type="primary" ghost size="small" onClick={() => handleDownload(row)}>
              {$t('common.download')}
            </NButton>
          </div>
        ) : (
          $t('page.hrm.training.noFile')
        )
    }
  ]
});

// ---- 扫描件预览 / 下载 ----
const previewVisible = ref(false);
const previewFileName = ref('');
const previewUrl = ref('');
const previewLoading = ref(false);

function mimeByFileName(fileName: string): string {
  const ext = fileName.split('.').pop()?.toLowerCase() ?? '';
  const map: Record<string, string> = {
    pdf: 'application/pdf',
    png: 'image/png',
    jpg: 'image/jpeg',
    jpeg: 'image/jpeg',
    gif: 'image/gif',
    webp: 'image/webp'
  };
  return map[ext] ?? 'application/octet-stream';
}

function base64ToBlob(content: string, mime: string): Blob {
  const raw = atob(content);
  const bytes = new Uint8Array(raw.length);
  for (let i = 0; i < raw.length; i += 1) {
    bytes[i] = raw.charCodeAt(i);
  }
  return new Blob([bytes], { type: mime });
}

async function handlePreview(row: Api.Hrm.MyTrainingCertificate) {
  previewLoading.value = true;
  const { data, error } = await DownloadMyCertificateFile(row.uuid);
  previewLoading.value = false;
  if (error) {
    return;
  }
  previewFileName.value = data.fileName;
  previewUrl.value = `data:${mimeByFileName(data.fileName)};base64,${data.content}`;
  previewVisible.value = true;
}

async function handleDownload(row: Api.Hrm.MyTrainingCertificate) {
  const { data, error } = await DownloadMyCertificateFile(row.uuid);
  if (error) {
    return;
  }
  const blob = base64ToBlob(data.content, mimeByFileName(data.fileName));
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = data.fileName;
  a.click();
  URL.revokeObjectURL(url);
}
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <NCard :title="$t('page.hrm.training.mySessionTitle')" :bordered="false" size="small" class="card-wrapper">
      <template #header-extra>
        <TableHeaderOperation
          v-model:columns="sessionColumnChecks"
          :loading="sessionLoading"
          @refresh="getSessionData"
        />
      </template>
      <NDataTable
        :columns="sessionColumns"
        :data="sessionData"
        size="small"
        :flex-height="!appStore.isMobile"
        :scroll-x="1000"
        :loading="sessionLoading"
        remote
        :row-key="row => row.uuid"
        :pagination="sessionPagination"
        class="sm:h-full"
      />
    </NCard>

    <NCard :title="$t('page.hrm.training.myCertTitle')" :bordered="false" size="small" class="card-wrapper">
      <template #header-extra>
        <TableHeaderOperation v-model:columns="certColumnChecks" :loading="certLoading" @refresh="getCertData" />
      </template>
      <NDataTable
        :columns="certColumns"
        :data="certData"
        size="small"
        :flex-height="!appStore.isMobile"
        :scroll-x="900"
        :loading="certLoading"
        remote
        :row-key="row => row.uuid"
        :pagination="certPagination"
        class="sm:h-full"
      />
    </NCard>

    <NModal v-model:show="previewVisible" class="w-80vw">
      <NCard
        :title="`${$t('page.hrm.training.preview')} - ${previewFileName}`"
        :bordered="false"
        size="small"
        closable
        @close="previewVisible = false"
      >
        <iframe v-if="previewUrl" :src="previewUrl" :title="previewFileName" class="h-70vh w-full" frameborder="0" />
      </NCard>
    </NModal>
  </div>
</template>

<style scoped></style>
