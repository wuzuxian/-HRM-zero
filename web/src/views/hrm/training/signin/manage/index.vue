<script setup lang="tsx">
import { computed, onBeforeUnmount, ref } from 'vue';
import { NAlert, NButton, NCard, NDataTable, NDescriptions, NDescriptionsItem, NInput, NModal, NPopconfirm, NSelect, NStatistic, NTag } from 'naive-ui';
import QRCode from 'qrcode';
import {
  ExportSigninDetail,
  GetSigninDetailList,
  GetSigninStats,
  GetTrainingSessionList,
  RefreshSigninToken,
  SigninForParticipant
} from '@/service/api';
import { $t } from '@/locales';
import { hrmTrainingSessionStatusRecord } from '@/constants/business';
import { useAuth } from '@/hooks/business/auth';

const { hasAuth } = useAuth();

const canRefresh = computed(() => hasAuth('hrm:signin:refreshSigninToken'));
const canSignFor = computed(() => hasAuth('hrm:signin:signinForParticipant'));
const canExport = computed(() => hasAuth('hrm:signin:exportSigninDetail'));

// ---- 班次选择（已发布/进行中班次下拉；demo-inventory 6: 选择班次 → 应到/已签/未到） ----
const sessionOptions = ref<Api.Hrm.TrainingSession[]>([]);
const sessionOptionsLoading = ref(false);
const selectedSessionUuid = ref<string | null>(null);
const selectedSession = computed(
  () => sessionOptions.value.find(s => s.uuid === selectedSessionUuid.value) ?? null
);

const sessionSelectOptions = computed(() =>
  sessionOptions.value.map(s => ({
    label: `${s.sessionNo} ${s.sessionName}（${$t(hrmTrainingSessionStatusRecord[s.status!])}）`,
    value: s.uuid
  }))
);

async function loadSessions() {
  sessionOptionsLoading.value = true;
  const { data, error } = await GetTrainingSessionList({ current: 1, size: 100, sessionName: null, planUuid: null, status: null });
  sessionOptionsLoading.value = false;
  if (error) {
    return;
  }
  sessionOptions.value = data?.records ?? [];
}

// ---- 统计（应到/已签/未到） ----
const stats = ref<Api.Hrm.SigninStats | null>(null);
const statsLoading = ref(false);

async function loadStats() {
  if (!selectedSessionUuid.value) {
    stats.value = null;
    return;
  }
  statsLoading.value = true;
  const { data, error } = await GetSigninStats(selectedSessionUuid.value);
  statsLoading.value = false;
  if (error) {
    return;
  }
  stats.value = data ?? null;
}

// ---- 课堂动态码（短时效轮换，前端渲染二维码） ----
const tokenInfo = ref<Api.Hrm.RefreshSigninTokenResponse | null>(null);
const tokenLoading = ref(false);
const qrDataUrl = ref('');
let qrRefreshTimer: ReturnType<typeof setInterval> | null = null;

async function refreshToken(showTip = true) {
  if (!selectedSessionUuid.value) {
    return;
  }
  tokenLoading.value = true;
  const { data, error } = await RefreshSigninToken(selectedSessionUuid.value);
  tokenLoading.value = false;
  if (error) {
    return;
  }
  tokenInfo.value = data ?? null;
  if (tokenInfo.value) {
    const url = `${window.location.origin}${tokenInfo.value.signinUrl}`;
    qrDataUrl.value = await QRCode.toDataURL(url, { width: 220, margin: 1 });
  }
  if (showTip) {
    window.$message?.success($t('page.hrm.training.signin.tokenRefreshed'));
  }
  restartQrTimer();
}

function restartQrTimer() {
  stopQrTimer();
  // 剩余 TTL < 60s 时自动刷新（后端 TTL 10 分钟）
  qrRefreshTimer = setInterval(
    () => {
      refreshToken(false);
    },
    9 * 60 * 1000
  );
}

function stopQrTimer() {
  if (qrRefreshTimer) {
    clearInterval(qrRefreshTimer);
    qrRefreshTimer = null;
  }
}

onBeforeUnmount(stopQrTimer);

// ---- 签到明细表（姓名/工号/部门/签到时间/方式/操作人/原因） ----
const details = ref<Api.Hrm.SigninDetail[]>([]);
const detailsLoading = ref(false);

async function loadDetails() {
  if (!selectedSessionUuid.value) {
    details.value = [];
    return;
  }
  detailsLoading.value = true;
  const { data, error } = await GetSigninDetailList(selectedSessionUuid.value);
  detailsLoading.value = false;
  if (error) {
    return;
  }
  details.value = data?.records ?? [];
}

const detailColumns = computed(() => [
  { key: 'index', title: $t('common.index'), align: 'center' as const, width: 56, render: (_: Api.Hrm.SigninDetail, i: number) => i + 1 },
  { key: 'nickName', title: $t('page.hrm.training.signin.nickName'), align: 'center' as const, minWidth: 90 },
  { key: 'username', title: $t('page.hrm.training.signin.username'), align: 'center' as const, minWidth: 100 },
  { key: 'employeeNo', title: $t('page.hrm.training.signin.employeeNo'), align: 'center' as const, width: 100, render: (row: Api.Hrm.SigninDetail) => row.employeeNo || '-' },
  { key: 'orgName', title: $t('page.hrm.training.signin.orgName'), align: 'center' as const, minWidth: 110, render: (row: Api.Hrm.SigninDetail) => row.orgName || '-' },
  { key: 'signInTime', title: $t('page.hrm.training.signin.signInTime'), align: 'center' as const, width: 160, render: (row: Api.Hrm.SigninDetail) => row.signInTime || '-' },
  {
    key: 'signInMethod',
    title: $t('page.hrm.training.signin.signInMethod'),
    align: 'center' as const,
    width: 110,
    render: (row: Api.Hrm.SigninDetail) => renderMethod(row.signInMethod)
  },
  { key: 'signInByName', title: $t('page.hrm.training.signin.signInByName'), align: 'center' as const, width: 110, render: (row: Api.Hrm.SigninDetail) => row.signInByName || '-' },
  { key: 'signInReason', title: $t('page.hrm.training.signin.signInReason'), align: 'center' as const, minWidth: 140, render: (row: Api.Hrm.SigninDetail) => row.signInReason || '-' },
  {
    key: 'operate',
    title: $t('common.operate'),
    align: 'center' as const,
    width: 90,
    fixed: 'right' as const,
    render: (row: Api.Hrm.SigninDetail) => renderRowOperate(row)
  }
]);

function renderMethod(method: Api.Hrm.SigninMethod | null) {
  if (!method) {
    return <NTag type="error">{$t('page.hrm.training.signin.methodAbsent')}</NTag>;
  }
  if (method === '1') {
    return <NTag type="success">{$t('page.hrm.training.signin.methodSelf')}</NTag>;
  }
  return <NTag type="warning">{$t('page.hrm.training.signin.methodFor')}</NTag>;
}

function renderRowOperate(row: Api.Hrm.SigninDetail) {
  // 未到名单行 → 补签入口（demo-inventory 6: 管理员补签）
  if (!row.signInTime && canSignFor.value) {
    return (
      <NButton type="primary" ghost size="small" onClick={() => openSignFor(row)}>
        {$t('page.hrm.training.signin.signFor')}
      </NButton>
    );
  }
  return null;
}

// ---- 补签弹窗（确认弹窗：方式=管理员补签，记录 操作人/原因） ----
const signForVisible = ref(false);
const signForRow = ref<Api.Hrm.SigninDetail | null>(null);
const signForReason = ref('');
const signForSubmitting = ref(false);

function openSignFor(row: Api.Hrm.SigninDetail) {
  signForRow.value = row;
  signForReason.value = '';
  signForVisible.value = true;
}

async function submitSignFor() {
  if (!signForRow.value) {
    return;
  }
  signForSubmitting.value = true;
  const { error } = await SigninForParticipant({ uuid: signForRow.value.uuid, reason: signForReason.value });
  signForSubmitting.value = false;
  if (error) {
    return;
  }
  signForVisible.value = false;
  window.$message?.success($t('page.hrm.training.signin.signForSuccess'));
  loadDetails();
  loadStats();
}

// ---- 导出签到明细 xlsx（base64 → 下载） ----
async function handleExport() {
  if (!selectedSessionUuid.value) {
    return;
  }
  const { data, error } = await ExportSigninDetail(selectedSessionUuid.value);
  if (error || !data) {
    return;
  }
  const raw = atob(data.content);
  const bytes = new Uint8Array(raw.length);
  for (let i = 0; i < raw.length; i += 1) {
    bytes[i] = raw.charCodeAt(i);
  }
  const url = URL.createObjectURL(new Blob([bytes], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' }));
  const a = document.createElement('a');
  a.href = url;
  a.download = data.fileName;
  a.click();
  URL.revokeObjectURL(url);
}

// ---- session 切换 → 全量刷新 ----
function handleSessionChange(uuid: string | null) {
  selectedSessionUuid.value = uuid;
  tokenInfo.value = null;
  qrDataUrl.value = '';
  stopQrTimer();
  loadStats();
  loadDetails();
  if (uuid && selectedSession.value && selectedSession.value.status !== '3') {
    // 已发布/进行中班次自动签发动态码
    refreshToken(false);
  }
}

function handleSessionLoaded() {
  loadSessions().then(() => {
    if (!selectedSessionUuid.value && sessionOptions.value.length > 0) {
      const first = sessionOptions.value[0];
      if (first) {
        handleSessionChange(first.uuid);
      }
    }
  });
}

handleSessionLoaded();

// ---- 未到名单（补签入口聚合视图） ----
const absentRows = computed(() => details.value.filter(d => !d.signInTime));
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <NCard :title="$t('page.hrm.training.signin.title')" :bordered="false" size="small" class="card-wrapper">
      <div class="flex flex-wrap items-center gap-12px">
        <span class="font-medium">{{ $t('page.hrm.training.signin.selectSession') }}</span>
        <NSelect
          v-model:value="selectedSessionUuid"
          class="w-360px"
          :options="sessionSelectOptions"
          :loading="sessionOptionsLoading"
          :placeholder="$t('page.hrm.training.signin.sessionPlaceholder')"
          clearable
          filterable
          @update:value="handleSessionChange"
        />
        <NButton v-if="selectedSessionUuid" :loading="detailsLoading || statsLoading" @click="handleSessionChange(selectedSessionUuid)">
          {{ $t('page.hrm.training.signin.refreshData') }}
        </NButton>
      </div>
    </NCard>

    <template v-if="selectedSessionUuid">
      <div class="grid grid-cols-3 gap-16px lt-sm:grid-cols-1">
        <NCard size="small" class="card-wrapper">
          <NStatistic :label="$t('page.hrm.training.signin.expected')" :value="stats?.expected ?? 0" />
        </NCard>
        <NCard size="small" class="card-wrapper">
          <NStatistic :label="$t('page.hrm.training.signin.signed')" :value="stats?.signed ?? 0" />
        </NCard>
        <NCard size="small" class="card-wrapper">
          <NStatistic :label="$t('page.hrm.training.signin.absent')" :value="stats?.absent ?? 0" />
        </NCard>
      </div>

      <NCard :title="$t('page.hrm.training.signin.qrTitle')" :bordered="false" size="small" class="card-wrapper">
        <div class="flex flex-wrap items-center gap-24px">
          <div v-if="qrDataUrl" class="flex-col-center gap-8px">
            <img :src="qrDataUrl" :alt="$t('page.hrm.training.signin.qrTitle')" class="h-220px w-220px" />
            <span class="text-12px text-gray-500">{{ $t('page.hrm.training.signin.qrExpireTip', { expire: tokenInfo?.expireTime }) }}</span>
          </div>
          <div class="flex-col gap-12px">
            <NPopconfirm v-if="canRefresh" @positive-click="refreshToken()">
              <template #trigger>
                <NButton type="primary" :loading="tokenLoading">{{ $t('page.hrm.training.signin.refreshToken') }}</NButton>
              </template>
              {{ $t('page.hrm.training.signin.refreshTokenConfirm') }}
            </NPopconfirm>
            <NAlert type="info" :show-icon="true" class="max-w-420px">
              {{ $t('page.hrm.training.signin.qrTip') }}
            </NAlert>
          </div>
        </div>
      </NCard>

      <NCard :title="$t('page.hrm.training.signin.detailTitle')" :bordered="false" size="small" class="card-wrapper">
        <template #header-extra>
          <NPopconfirm v-if="canSignFor && absentRows.length > 0" @positive-click="openSignFor(absentRows[0]!)">
            <template #trigger>
              <NButton size="small" type="warning" ghost>{{ $t('page.hrm.training.signin.signForFirstAbsent') }}</NButton>
            </template>
            {{ $t('page.hrm.training.signin.signForFirstAbsentTip', { name: absentRows[0]?.nickName }) }}
          </NPopconfirm>
          <NButton v-if="canExport" size="small" type="primary" ghost @click="handleExport">
            {{ $t('page.hrm.training.signin.export') }}
          </NButton>
        </template>
        <NDataTable
          :columns="detailColumns"
          :data="details"
          size="small"
          :scroll-x="1100"
          :loading="detailsLoading"
          :row-key="(row: Api.Hrm.SigninDetail) => row.uuid"
          :pagination="false"
        />
      </NCard>
    </template>

    <NModal v-model:show="signForVisible" preset="card" :title="$t('page.hrm.training.signin.signForTitle')" class="w-460px">
      <NDescriptions v-if="signForRow" :column="1" size="small" bordered>
        <NDescriptionsItem :label="$t('page.hrm.training.signin.nickName')">{{ signForRow.nickName }}</NDescriptionsItem>
        <NDescriptionsItem :label="$t('page.hrm.training.signin.username')">{{ signForRow.username }}</NDescriptionsItem>
        <NDescriptionsItem :label="$t('page.hrm.training.signin.employeeNo')">{{ signForRow.employeeNo || '-' }}</NDescriptionsItem>
        <NDescriptionsItem :label="$t('page.hrm.training.signin.orgName')">{{ signForRow.orgName || '-' }}</NDescriptionsItem>
      </NDescriptions>
      <div class="mt-16px flex-col gap-8px">
        <span>{{ $t('page.hrm.training.signin.signForReasonLabel') }}</span>
        <NInput v-model:value="signForReason" type="textarea" :rows="2" :placeholder="$t('page.hrm.training.signin.signForReasonPlaceholder')" />
      </div>
      <div class="mt-16px flex justify-end gap-8px">
        <NButton @click="signForVisible = false">{{ $t('common.cancel') }}</NButton>
        <NPopconfirm @positive-click="submitSignFor">
          <template #trigger>
            <NButton type="primary" :loading="signForSubmitting">{{ $t('page.hrm.training.signin.signForConfirm') }}</NButton>
          </template>
          {{ $t('page.hrm.training.signin.signForConfirmTip') }}
        </NPopconfirm>
      </div>
    </NModal>
  </div>
</template>

<style scoped></style>
