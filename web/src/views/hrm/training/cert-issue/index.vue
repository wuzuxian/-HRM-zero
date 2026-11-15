<script setup lang="tsx">
import { computed, reactive, ref } from 'vue';
import { NButton, NPopconfirm, NTag } from 'naive-ui';
import {
  DownloadCertificateFile,
  GetCertificateList,
  IssueCertificates,
  RevokeCertificates,
  UploadCertificateFile
} from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import {
  hrmCertificateSourceRecord,
  hrmTrainingCertificateStatusRecord
} from '@/constants/business';
import { useTable } from '@/hooks/common/table';
import { useAuth } from '@/hooks/business/auth';
import CertBatchIssueModal from './modules/cert-batch-issue-modal.vue';
import CertImportModal from './modules/cert-import-modal.vue';
import CertUploadModal from './modules/cert-upload-modal.vue';

const appStore = useAppStore();
const { hasAuth } = useAuth();

const canBatchIssue = computed(() => hasAuth('hrm:certificate:batchIssue'));
const canImport = computed(() => hasAuth('hrm:certificate:importExternal'));
const canIssue = computed(() => hasAuth('hrm:certificate:issue'));
const canRevoke = computed(() => hasAuth('hrm:certificate:revoke'));
const canUpload = computed(() => hasAuth('hrm:certificate:uploadFile'));
const canDownload = computed(() => hasAuth('hrm:certificate:downloadFile'));

type LoadingStatus = Record<string, boolean>;
const issueLoadingStatus = reactive<LoadingStatus>({});
const revokeLoadingStatus = reactive<LoadingStatus>({});

const {
  columns,
  columnChecks,
  data,
  getData,
  getDataByPage,
  loading,
  mobilePagination,
  searchParams,
  resetSearchParams
} = useTable({
  apiFn: GetCertificateList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    certNo: null,
    certName: null,
    certType: null,
    status: null,
    source: null,
    sessionUuid: null
  },
  columns: () => [
    {
      key: 'index',
      title: $t('common.index'),
      align: 'center',
      width: 64
    },
    {
      key: 'nickName',
      title: $t('page.hrm.training.certificate.student'),
      align: 'center',
      width: 110,
      render: row => (row.nickName ? `${row.nickName}（${row.username}）` : row.username || '-')
    },
    {
      key: 'certNo',
      title: $t('page.hrm.training.certNo'),
      align: 'center',
      width: 150
    },
    {
      key: 'certName',
      title: $t('page.hrm.training.certName'),
      align: 'center',
      minWidth: 130
    },
    {
      key: 'certType',
      title: $t('page.hrm.training.certTemplate.certType'),
      align: 'center',
      width: 100,
      render: row => row.certType || '-'
    },
    {
      key: 'templateName',
      title: $t('page.hrm.training.certificate.template'),
      align: 'center',
      width: 120,
      render: row => row.templateName || '-'
    },
    {
      key: 'source',
      title: $t('page.hrm.training.certificate.sourceLabel'),
      align: 'center',
      width: 100,
      render: row => {
        if (row.source === null) {
          return null;
        }
        return $t(hrmCertificateSourceRecord[row.source]);
      }
    },
    {
      key: 'sessionName',
      title: $t('page.hrm.training.certificate.session'),
      align: 'center',
      width: 130,
      render: row => row.sessionName || '-'
    },
    {
      key: 'issueTime',
      title: $t('page.hrm.training.issueTime'),
      align: 'center',
      width: 150,
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
      key: 'fileName',
      title: $t('page.hrm.training.scanFile'),
      align: 'center',
      width: 200,
      fixed: 'right',
      render: row => renderFileOperate(row)
    }
  ]
});

// ---- 扫描件预览 / 下载（与我的培训同一实现） ----
const previewVisible = ref(false);
const previewFileName = ref('');
const previewUrl = ref('');

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

async function handlePreview(row: Api.Hrm.TrainingCertificate) {
  const { data, error } = await DownloadCertificateFile(row.uuid);
  if (error) {
    return;
  }
  previewFileName.value = data.fileName;
  previewUrl.value = `data:${mimeByFileName(data.fileName)};base64,${data.content}`;
  previewVisible.value = true;
}

async function handleDownload(row: Api.Hrm.TrainingCertificate) {
  const { data, error } = await DownloadCertificateFile(row.uuid);
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

function renderFileOperate(row: Api.Hrm.TrainingCertificate) {
  const fileOps =
    row.hasFile && canDownload.value ? (
      <div class="flex-center gap-4px">
        <NButton type="primary" ghost size="small" onClick={() => handlePreview(row)}>
          {$t('page.hrm.training.preview')}
        </NButton>
        <NButton type="primary" ghost size="small" onClick={() => handleDownload(row)}>
          {$t('common.download')}
        </NButton>
      </div>
    ) : (
      <span>{$t('page.hrm.training.noFile')}</span>
    );
  return (
    <div class="flex-center flex-wrap gap-4px">
      {fileOps}
      {canUpload.value && (
        <NButton type="info" ghost size="small" onClick={() => handleUpload(row)}>
          {$t('page.hrm.training.certificate.upload')}
        </NButton>
      )}
      {row.status === '1' && canRevoke.value && (
        <NPopconfirm onPositiveClick={() => handleRevoke(row.uuid)}>
          {{
            default: () => $t('page.hrm.training.certificate.confirmRevoke'),
            trigger: () => (
              <NButton loading={revokeLoadingStatus[row.uuid]} type="error" ghost size="small">
                {$t('page.hrm.training.certificate.revoke')}
              </NButton>
            )
          }}
        </NPopconfirm>
      )}
      {row.status === '0' && canIssue.value && (
        <NPopconfirm onPositiveClick={() => handleIssue(row.uuid)}>
          {{
            default: () => $t('page.hrm.training.certificate.confirmIssue'),
            trigger: () => (
              <NButton loading={issueLoadingStatus[row.uuid]} type="warning" ghost size="small">
                {$t('page.hrm.training.certificate.issue')}
              </NButton>
            )
          }}
        </NPopconfirm>
      )}
    </div>
  );
}

async function handleIssue(uuid: string) {
  issueLoadingStatus[uuid] = true;
  const { error } = await IssueCertificates([uuid]);
  issueLoadingStatus[uuid] = false;
  if (error) {
    return;
  }
  window.$message?.success($t('common.updateSuccess'));
  getData();
}

async function handleRevoke(uuid: string) {
  revokeLoadingStatus[uuid] = true;
  const { error } = await RevokeCertificates([uuid]);
  revokeLoadingStatus[uuid] = false;
  if (error) {
    return;
  }
  window.$message?.success($t('common.updateSuccess'));
  getData();
}

// ---- 批量发证 / 导入 / 上传 modals ----
const batchIssueVisible = ref(false);
const importVisible = ref(false);
const uploadVisible = ref(false);
const uploadCertUuid = ref('');

function handleBatchIssue() {
  batchIssueVisible.value = true;
}

function handleImport() {
  importVisible.value = true;
}

function handleUpload(row: Api.Hrm.TrainingCertificate) {
  uploadCertUuid.value = row.uuid;
  uploadVisible.value = true;
}
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <CertSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <NCard
      :title="$t('page.hrm.training.certificate.title')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <NSpace>
          <NButton v-if="canBatchIssue" type="primary" @click="handleBatchIssue">
            {{ $t('page.hrm.training.certificate.batchIssue') }}
          </NButton>
          <NButton v-if="canImport" type="primary" ghost @click="handleImport">
            {{ $t('page.hrm.training.certificate.importExternal') }}
          </NButton>
          <TableHeaderOperation
            v-model:columns="columnChecks"
            :loading="loading"
            @refresh="getData"
          />
        </NSpace>
      </template>
      <NDataTable
        :columns="columns"
        :data="data"
        size="small"
        :flex-height="!appStore.isMobile"
        :scroll-x="1600"
        :loading="loading"
        remote
        :row-key="row => row.uuid"
        :pagination="mobilePagination"
        class="sm:h-full"
      />
      <CertBatchIssueModal v-model:visible="batchIssueVisible" @submitted="getData" />
      <CertImportModal v-model:visible="importVisible" @submitted="getData" />
      <CertUploadModal v-model:visible="uploadVisible" :cert-uuid="uploadCertUuid" @submitted="getData" />
      <NModal v-model:show="previewVisible" :title="previewFileName" preset="card" class="w-720px">
        <iframe v-if="previewUrl" :src="previewUrl" class="h-560px w-full" />
      </NModal>
    </NCard>
  </div>
</template>

<style scoped></style>
