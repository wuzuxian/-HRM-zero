<script setup lang="tsx">
import { computed, ref } from 'vue';
import type { UploadFileInfo } from 'naive-ui';
import { NButton, NTag } from 'naive-ui';
import { DeletePunchRecord, ExportPunchRecord, GetPunchList, ImportPunchRecord, SettlePunchDay } from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { useTable } from '@/hooks/common/table';
import PunchRecordSearch from './modules/punch-record-search.vue';
import PunchRecordOperateModal from './modules/punch-record-operate-modal.vue';

defineOptions({
  name: 'PunchRecord'
});

const appStore = useAppStore();

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
  apiFn: GetPunchList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    startDate: undefined,
    endDate: undefined,
    name: undefined,
    clockInStatus: undefined
  },
  columns: () => [
    {
      key: 'index',
      title: $t('common.index'),
      align: 'center',
      width: 64
    },
    {
      key: 'employeeName',
      title: $t('page.hrm.attendance.punchRecord.employeeName'),
      align: 'center',
      width: 100
    },
    {
      key: 'username',
      title: $t('page.hrm.attendance.punchRecord.username'),
      align: 'center',
      width: 110,
      render: row => row.username || '-'
    },
    {
      key: 'punchDate',
      title: $t('page.hrm.attendance.punchRecord.punchDate'),
      align: 'center',
      width: 110
    },
    {
      key: 'weekday',
      title: $t('page.hrm.attendance.punchRecord.weekday'),
      align: 'center',
      width: 70
    },
    {
      key: 'clockInTime',
      title: $t('page.hrm.attendance.punchRecord.clockInTime'),
      align: 'center',
      width: 160,
      render: row => row.clockInTime || '-'
    },
    {
      key: 'clockInStatus',
      title: $t('page.hrm.attendance.punchRecord.clockInStatus'),
      align: 'center',
      width: 90,
      render: row => renderStatus(row.clockInStatus)
    },
    {
      key: 'clockInLocation',
      title: $t('page.hrm.attendance.punchRecord.clockInLocation'),
      align: 'center',
      minWidth: 140,
      render: row => row.clockInLocation || '-'
    },
    {
      key: 'clockOutTime',
      title: $t('page.hrm.attendance.punchRecord.clockOutTime'),
      align: 'center',
      width: 160,
      render: row => row.clockOutTime || '-'
    },
    {
      key: 'clockOutStatus',
      title: $t('page.hrm.attendance.punchRecord.clockOutStatus'),
      align: 'center',
      width: 90,
      render: row => renderStatus(row.clockOutStatus)
    },
    {
      key: 'settleStatus',
      title: $t('page.hrm.attendance.punchRecord.settleStatus'),
      align: 'center',
      width: 100,
      render: row => renderSettle(row.settleStatus)
    },
    {
      key: 'overtime',
      title: $t('page.hrm.attendance.punchRecord.overtime'),
      align: 'center',
      width: 70,
      render: row =>
        row.overtime === '1'
          ? $t('page.hrm.attendance.punchRecord.overtimeYes')
          : $t('page.hrm.attendance.punchRecord.overtimeNo')
    },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      width: 130,
      render: row => (
        <div class="flex-center gap-8px">
          <NButton v-auth={['hrm:punch:edit']} type="primary" ghost size="small" onClick={() => handleEdit(row)}>
            {$t('common.edit')}
          </NButton>
          <NButton v-auth={['hrm:punch:delete']} type="error" ghost size="small" onClick={() => handleDelete(row)}>
            {$t('common.delete')}
          </NButton>
        </div>
      )
    }
  ]
});

function renderStatus(status: string) {
  if (status === 'normal') {
    return (
      <NTag type="success" size="small">
        {$t('page.hrm.attendance.punchRecord.statusNormal')}
      </NTag>
    );
  }
  if (status === 'field') {
    return (
      <NTag type="info" size="small">
        {$t('page.hrm.attendance.punchRecord.statusField')}
      </NTag>
    );
  }
  return (
    <NTag type="error" size="small">
      {$t('page.hrm.attendance.punchRecord.statusMissing')}
    </NTag>
  );
}

function renderSettle(status: string) {
  const map: Record<string, { label: string; type: 'success' | 'warning' | 'error' | 'info' | 'default' }> = {
    normal: { label: $t('page.hrm.attendance.punchRecord.settleNormal'), type: 'success' },
    late: { label: $t('page.hrm.attendance.punchRecord.settleLate'), type: 'warning' },
    early: { label: $t('page.hrm.attendance.punchRecord.settleEarly'), type: 'warning' },
    missing: { label: $t('page.hrm.attendance.punchRecord.settleMissing'), type: 'error' },
    rest: { label: $t('page.hrm.attendance.punchRecord.settleRest'), type: 'default' },
    no_rule: { label: $t('page.hrm.attendance.punchRecord.settleNoRule'), type: 'info' }
  };
  const meta = map[status];
  if (!meta) return status || '-';
  return (
    <NTag type={meta.type} size="small">
      {meta.label}
    </NTag>
  );
}

// modal state (manual, HRM 页面不用 useTableOperate)
const operateVisible = ref(false);
const operateType = ref<'add' | 'edit'>('add');
const editingRow = ref<Api.Hrm.Punch | null>(null);

function handleAdd() {
  operateType.value = 'add';
  editingRow.value = null;
  operateVisible.value = true;
}

function handleEdit(row: Api.Hrm.Punch) {
  operateType.value = 'edit';
  editingRow.value = { ...row };
  operateVisible.value = true;
}

async function handleDelete(row: Api.Hrm.Punch) {
  const { error } = await DeletePunchRecord([row.uuid]);
  if (error) return;
  window.$message?.success($t('common.deleteSuccess'));
  getData();
}

function handleBatchDeleted() {
  window.$message?.success($t('common.deleteSuccess'));
  getData();
}

function handleSubmitted() {
  getData();
}

// ---- 批量选择 ----
const checkedKeys = ref<string[]>([]);

// ---- 导出 ----
const exporting = ref(false);

async function handleExport() {
  exporting.value = true;
  const { data: exportData, error } = await ExportPunchRecord({
    startDate: searchParams.startDate,
    endDate: searchParams.endDate,
    name: searchParams.name,
    clockInStatus: searchParams.clockInStatus
  });
  exporting.value = false;
  if (error || !exportData) return;
  const binary = window.atob(exportData.content);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i += 1) {
    bytes[i] = binary.charCodeAt(i);
  }
  const blob = new Blob([bytes], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = exportData.fileName;
  link.click();
  URL.revokeObjectURL(url);
}

// ---- 导入外部考勤数据 (CSV) ----
const importVisible = ref(false);
const importFileName = ref('');
const importFileContent = ref('');
const importing = ref(false);

function handleUploadChange(options: { file: UploadFileInfo }) {
  if (options.file.file) {
    handleImportFile(options.file.file);
  }
}

function handleImportFile(file: File) {
  importFileName.value = file.name;
  const reader = new FileReader();
  reader.onload = () => {
    importFileContent.value = String(reader.result ?? '');
  };
  reader.readAsText(file);
  return false;
}

async function handleImport() {
  if (!importFileContent.value) {
    window.$message?.error($t('page.hrm.attendance.punchRecord.importFileRequired'));
    return;
  }
  importing.value = true;
  const { data: importData, error } = await ImportPunchRecord({
    fileName: importFileName.value,
    fileContent: importFileContent.value
  });
  importing.value = false;
  if (error) return;
  window.$message?.success(`${$t('page.hrm.attendance.punchRecord.importSuccess')} (${importData?.successCount ?? 0})`);
  importVisible.value = false;
  importFileName.value = '';
  importFileContent.value = '';
  getData();
}

// ---- 手动日结 ----
const settleVisible = ref(false);
const settling = ref(false);
const settleRange = ref<[number, number] | null>(null);

const settlePreviewColumns = computed(() => [
  {
    key: 'employeeName',
    title: $t('page.hrm.attendance.punchRecord.employeeName'),
    align: 'center' as const,
    width: 100
  },
  { key: 'punchDate', title: $t('page.hrm.attendance.punchRecord.punchDate'), align: 'center' as const, width: 110 },
  {
    key: 'settleStatus',
    title: $t('page.hrm.attendance.punchRecord.settleStatus'),
    align: 'center' as const,
    render: (row: Api.Hrm.PunchSettleItem) => renderSettle(row.settleStatus)
  },
  {
    key: 'lateMinutes',
    title: $t('page.hrm.attendance.punchRecord.lateMinutes'),
    align: 'center' as const,
    width: 100
  },
  {
    key: 'earlyMinutes',
    title: $t('page.hrm.attendance.punchRecord.earlyMinutes'),
    align: 'center' as const,
    width: 100
  },
  { key: 'workHours', title: $t('page.hrm.attendance.punchRecord.workHours'), align: 'center' as const, width: 100 }
]);

const settleResult = ref<Api.Hrm.PunchSettleItem[]>([]);

function handleOpenSettle() {
  settleResult.value = [];
  settleRange.value = null;
  settleVisible.value = true;
}

async function handleSettle() {
  if (!settleRange.value) {
    window.$message?.error($t('page.hrm.attendance.punchRecord.settleRangeRequired'));
    return;
  }
  settling.value = true;
  const fmtTs = (ts: number) => {
    const d = new Date(ts);
    const m = `${d.getMonth() + 1}`.padStart(2, '0');
    const day = `${d.getDate()}`.padStart(2, '0');
    return `${d.getFullYear()}-${m}-${day}`;
  };
  const { data: settleData, error } = await SettlePunchDay({
    startDate: fmtTs(settleRange.value[0]),
    endDate: fmtTs(settleRange.value[1])
  });
  settling.value = false;
  if (error) return;
  settleResult.value = settleData?.records ?? [];
  window.$message?.success($t('page.hrm.attendance.punchRecord.settleSuccess'));
  getData();
}
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <PunchRecordSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <NCard
      :title="$t('page.hrm.attendance.punchRecord.title')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <div class="flex items-center gap-8px">
          <NButton v-auth="'hrm:punch:add'" type="primary" @click="handleAdd">
            {{ $t('page.hrm.attendance.punchRecord.addTitle') }}
          </NButton>
          <NButton v-auth="'hrm:punch:export'" :loading="exporting" @click="handleExport">
            {{ $t('page.hrm.attendance.punchRecord.exportTitle') }}
          </NButton>
          <NButton v-auth="'hrm:punch:import'" @click="importVisible = true">
            {{ $t('page.hrm.attendance.punchRecord.importTitle') }}
          </NButton>
          <NButton v-auth="'hrm:punch:settle'" type="warning" ghost @click="handleOpenSettle">
            {{ $t('page.hrm.attendance.punchRecord.settleTitle') }}
          </NButton>
          <TableHeaderOperation
            v-model:columns="columnChecks"
            :loading="loading"
            :disabled-delete="checkedKeys.length === 0"
            @delete="handleBatchDeleted"
            @refresh="getData"
          />
        </div>
      </template>
      <NDataTable
        v-model:checked-row-keys="checkedKeys"
        :columns="columns"
        :data="data"
        size="small"
        :flex-height="!appStore.isMobile"
        :scroll-x="1400"
        :loading="loading"
        remote
        :pagination="mobilePagination"
      />
    </NCard>
    <PunchRecordOperateModal
      v-model:visible="operateVisible"
      :operate-type="operateType"
      :row-data="editingRow"
      @submitted="handleSubmitted"
    />
    <NModal
      v-model:show="importVisible"
      :title="$t('page.hrm.attendance.punchRecord.importTitle')"
      preset="card"
      class="w-520px"
    >
      <NUpload :max="1" accept=".csv" :default-upload="false" @change="handleUploadChange">
        <NUploadDragger>
          <div class="py-16px">{{ $t('page.hrm.attendance.punchRecord.importHint') }}</div>
        </NUploadDragger>
      </NUpload>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="importVisible = false">{{ $t('common.cancel') }}</NButton>
          <NButton type="primary" :loading="importing" @click="handleImport">
            {{ $t('common.confirm') }}
          </NButton>
        </NSpace>
      </template>
    </NModal>
    <NModal
      v-model:show="settleVisible"
      :title="$t('page.hrm.attendance.punchRecord.settleTitle')"
      preset="card"
      class="w-720px"
    >
      <NForm label-placement="left" :label-width="100">
        <NFormItem :label="$t('page.hrm.attendance.punchRecord.dateRange')" path="settleRange">
          <NDatePicker v-model:value="settleRange" type="daterange" clearable value-format="yyyy-MM-dd" />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.punchRecord.settleScope')">
          <span class="text-gray-500">{{ $t('page.hrm.attendance.punchRecord.settleScopeHint') }}</span>
        </NFormItem>
      </NForm>
      <NDataTable
        v-if="settleResult.length > 0"
        :columns="settlePreviewColumns"
        :data="settleResult"
        size="small"
        :max-height="320"
      />
      <template #footer>
        <NSpace justify="end">
          <NButton @click="settleVisible = false">{{ $t('common.close') }}</NButton>
          <NButton type="primary" :loading="settling" @click="handleSettle">
            {{ $t('common.confirm') }}
          </NButton>
        </NSpace>
      </template>
    </NModal>
  </div>
</template>
