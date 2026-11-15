<script setup lang="tsx">
import { computed, ref } from 'vue';
import type { UploadFileInfo } from 'naive-ui';
import { NButton } from 'naive-ui';
import {
  CopyScheduleWeek,
  ExportScheduleBoard,
  GenerateScheduleFromCycle,
  GetScheduleMatrix,
  GetShiftList,
  ImportScheduleBoard,
  PublishScheduleMonth,
  SaveScheduleCell
} from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import ScheduleBoardSearch from './modules/schedule-board-search.vue';
import ScheduleGenerateModal from './modules/schedule-generate-modal.vue';
import ScheduleCopyWeekModal from './modules/schedule-copy-week-modal.vue';
import ScheduleImportModal from './modules/schedule-import-modal.vue';

defineOptions({
  name: 'ScheduleBoard'
});

const appStore = useAppStore();

// ---- 查询状态（考勤组 + 月份） ----
const groupUuid = ref('');
const month = ref('');
const boardLoading = ref(false);

type MatrixRow = Api.Hrm.ScheduleMatrixRow;

const board = ref<Api.Hrm.ScheduleMatrixResponse | null>(null);
const rows = ref<MatrixRow[]>([]);
const monthDays = ref<string[]>([]);

// 班次列表（渲染单元格缩写 + 弹窗选择用）
const shiftList = ref<Api.Hrm.Shift[]>([]);
const shiftNameMap = computed(() => {
  const map: Record<string, string> = {};
  for (const s of shiftList.value) {
    map[s.uuid] = s.shiftName;
  }
  return map;
});

async function loadShifts() {
  const { data, error } = await GetShiftList({ current: 1, size: 100 });
  if (error || !data) return;
  shiftList.value = data.records;
}

function shiftAbbr(shiftUuid: string) {
  const name = shiftNameMap.value[shiftUuid] ?? '';
  return name ? name.slice(0, 2) : '';
}

function shiftColor(shiftUuid: string) {
  // 依 uuid 稳定取色（演示用：不同班次不同底色）
  const colors = ['#18a058', '#2080f0', '#f0a020', '#d03050', '#722ed1', '#13c2c2'];
  let sum = 0;
  for (const ch of shiftUuid) sum = (sum + ch.charCodeAt(0)) % colors.length;
  return colors[sum]!;
}

async function loadBoard() {
  if (!groupUuid.value || !month.value) {
    window.$message?.warning($t('page.hrm.attendance.scheduleBoard.selectGroupFirst'));
    return;
  }
  boardLoading.value = true;
  const filters = searchParams.value;
  const { data, error } = await GetScheduleMatrix({
    groupUuid: groupUuid.value,
    month: month.value,
    name: filters.name || undefined,
    employeeNo: filters.employeeNo || undefined,
    orgName: filters.orgName || undefined
  });
  boardLoading.value = false;
  if (error || !data) return;
  board.value = data;
  rows.value = data.rows;
  monthDays.value = data.monthDays;
}

// ---- 搜索筛选 ----
const searchParams = ref({ name: '', employeeNo: '', orgName: '' });

function handleSearch() {
  loadBoard();
}

// ---- 矩阵列（当月每日一列，key 用真实字段 cells 的下标不行 → 用 date 字符串作 key 前缀列） ----
// NDataTable 列 key 不能是虚拟值；列数组按 monthDays 动态生成，key = 日期字符串。
// 行数据对象被映射为 { employeeName, employeeNo, orgName, publishedCount, uuid, [date]: shiftUuid } 形态。
type BoardTableData = {
  uuid: string;
  employeeNo: string;
  employeeName: string;
  orgName: string;
  publishedCount: number;
} & Record<string, string | number>;

const tableData = computed<BoardTableData[]>(() =>
  rows.value.map(row => {
    const record: BoardTableData = {
      uuid: row.employeeUuid,
      employeeNo: row.employeeNo,
      employeeName: row.employeeName,
      orgName: row.orgName,
      publishedCount: row.publishedCount
    };
    row.cells.forEach(cell => {
      record[cell.date] = cell.shiftUuid;
    });
    return record;
  })
);

// 单元格点击选择班次（点击循环切换：休 → 班次1 → 班次2 → … → 休）
function nextShift(current: string): string {
  const enabled = shiftList.value.filter(s => s.status === '1');
  if (enabled.length === 0) return current;
  const idx = enabled.findIndex(s => s.uuid === current);
  return enabled[(idx + 1) % enabled.length]!.uuid;
}

async function handleCellClick(row: BoardTableData, date: string) {
  if (publishedDays.value.has(`${row.uuid}:${date}`)) {
    window.$message?.warning($t('page.hrm.attendance.scheduleBoard.publishedLocked'));
    return;
  }
  const current = (row[date] as string) ?? '';
  const target = nextShift(current);
  const { error } = await SaveScheduleCell({
    groupUuid: groupUuid.value,
    employeeUuid: row.uuid,
    date,
    shiftUuid: target
  });
  if (error) return;
  row[date] = target;
}

const publishedDays = computed(() => {
  const set = new Set<string>();
  for (const row of rows.value) {
    for (const cell of row.cells) {
      if (cell.published === '1' && cell.shiftUuid) {
        set.add(`${row.employeeUuid}:${cell.date}`);
      }
    }
  }
  return set;
});

type BoardColumn = import('naive-ui').DataTableColumns<BoardTableData>[number];

const boardColumns = computed<import('naive-ui').DataTableColumns<BoardTableData>>(() => {
  const fixed: BoardColumn[] = [
    {
      key: 'employeeNo',
      title: $t('page.hrm.attendance.scheduleBoard.employeeNo'),
      align: 'center' as const,
      width: 90,
      fixed: 'left' as const
    },
    {
      key: 'employeeName',
      title: $t('page.hrm.attendance.scheduleBoard.employeeName'),
      align: 'center' as const,
      width: 90,
      fixed: 'left' as const,
      render: (row: BoardTableData) => <span class="whitespace-nowrap">{row.employeeName}</span>
    },
    {
      key: 'orgName',
      title: $t('page.hrm.attendance.scheduleBoard.orgName'),
      align: 'center' as const,
      width: 90,
      fixed: 'left' as const,
      render: (row: BoardTableData) => row.orgName || '-'
    }
  ];
  const dayCols: BoardColumn[] = monthDays.value.map(date => {
    const dayNum = Number(date.slice(8, 10));
    const weekday = new Date(`${date}T00:00:00`).getDay();
    const isWeekend = weekday === 0 || weekday === 6;
    return {
      key: date,
      title: () => (
        <div class={isWeekend ? 'text-red-500' : ''}>
          <div>{dayNum}</div>
        </div>
      ),
      align: 'center' as const,
      width: 52,
      render: (row: BoardTableData) => {
        const shiftUuid = (row[date] as string) ?? '';
        if (!shiftUuid) return <span class="text-gray-300">休</span>;
        const isPublished = publishedDays.value.has(`${row.uuid}:${date}`);
        return (
          <span
            class="inline-block cursor-pointer rounded px-4px py-2px text-xs text-white"
            style={{ backgroundColor: shiftColor(shiftUuid) }}
            onClick={() => handleCellClick(row, date)}
          >
            {shiftAbbr(shiftUuid)}
            {isPublished ? '·已发布' : ''}
          </span>
        );
      }
    };
  });
  const last: BoardColumn[] = [
    {
      key: 'publishedCount',
      title: $t('page.hrm.attendance.scheduleBoard.publishedCount'),
      align: 'center' as const,
      width: 80,
      fixed: 'right' as const,
      render: (row: BoardTableData) => `${row.publishedCount}`
    }
  ];
  return [...fixed, ...dayCols, ...last];
});

// ---- 周期生成 / 复制周 / 导入 / 导出 / 发布 ----
const generateVisible = ref(false);
const copyWeekVisible = ref(false);
const importVisible = ref(false);
const importing = ref(false);
const exporting = ref(false);
const publishing = ref(false);

async function handleExport() {
  exporting.value = true;
  const { data, error } = await ExportScheduleBoard({ groupUuid: groupUuid.value, month: month.value });
  exporting.value = false;
  if (error || !data) return;
  const binary = window.atob(data.content);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i += 1) {
    bytes[i] = binary.charCodeAt(i);
  }
  const blob = new Blob([bytes], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = data.fileName;
  link.click();
  URL.revokeObjectURL(url);
}

async function handlePublish() {
  publishing.value = true;
  const { data, error } = await PublishScheduleMonth({ groupUuid: groupUuid.value, month: month.value });
  publishing.value = false;
  if (error) return;
  window.$message?.success($t('page.hrm.attendance.scheduleBoard.publishSuccess', { n: data?.publishedCount ?? 0 }));
  loadBoard();
}

function handleUploaded() {
  loadBoard();
}
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <ScheduleBoardSearch
      v-model:group-uuid="groupUuid"
      v-model:month="month"
      v-model:model="searchParams"
      @search="handleSearch"
      @load="loadBoard"
    />
    <NCard
      :title="$t('page.hrm.attendance.scheduleBoard.title')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <div class="flex items-center gap-8px">
          <NButton v-auth="'hrm:scheduleboard:generate'" type="primary" @click="generateVisible = true">
            {{ $t('page.hrm.attendance.scheduleBoard.generate') }}
          </NButton>
          <NButton v-auth="'hrm:scheduleboard:copyweek'" @click="copyWeekVisible = true">
            {{ $t('page.hrm.attendance.scheduleBoard.copyWeek') }}
          </NButton>
          <NButton v-auth="'hrm:scheduleboard:import'" @click="importVisible = true">
            {{ $t('page.hrm.attendance.scheduleBoard.import') }}
          </NButton>
          <NButton v-auth="'hrm:scheduleboard:export'" :loading="exporting" @click="handleExport">
            {{ $t('page.hrm.attendance.scheduleBoard.export') }}
          </NButton>
          <NButton v-auth="'hrm:scheduleboard:publish'" type="warning" ghost :loading="publishing" @click="handlePublish">
            {{ $t('page.hrm.attendance.scheduleBoard.publish') }}
          </NButton>
        </div>
      </template>
      <div v-if="board" class="mb-8px flex items-center gap-12px text-sm text-gray-500">
        <span>{{ $t('page.hrm.attendance.scheduleBoard.groupLabel', { name: board.groupName }) }}</span>
        <span>{{ $t('page.hrm.attendance.scheduleBoard.cellHint') }}</span>
      </div>
      <NDataTable
        :columns="boardColumns"
        :data="tableData"
        size="small"
        :loading="boardLoading"
        :scroll-x="Math.max(360, 280 + monthDays.length * 52 + 80)"
        :max-height="480"
        :bordered="true"
        :single-line="false"
      />
    </NCard>
    <ScheduleGenerateModal
      v-model:visible="generateVisible"
      :group-uuid="groupUuid"
      :month="month"
      @submitted="loadBoard"
    />
    <ScheduleCopyWeekModal
      v-model:visible="copyWeekVisible"
      :group-uuid="groupUuid"
      :month="month"
      @submitted="loadBoard"
    />
    <ScheduleImportModal
      v-model:visible="importVisible"
      :group-uuid="groupUuid"
      :month="month"
      @submitted="handleUploaded"
    />
  </div>
</template>
