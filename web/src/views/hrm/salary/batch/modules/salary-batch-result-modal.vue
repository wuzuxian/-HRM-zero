<script setup lang="tsx">
import { ref, watch } from 'vue';
import { NTag } from 'naive-ui';
import { GetSalaryBatchResult } from '@/service/api';
import { $t } from '@/locales';

defineOptions({
  name: 'SalaryBatchResultModal'
});

interface Props {
  /** batch uuid to load */
  batchUuid: string;
}

const props = defineProps<Props>();

const visible = defineModel<boolean>('visible', {
  default: false
});

const loading = ref(false);
const batch = ref<Api.Hrm.SalaryBatch | null>(null);
const items = ref<Api.Hrm.SalaryBatchItem[]>([]);

const itemClassTag = (cls: string) => {
  const map: Record<string, 'success' | 'warning' | 'info'> = {
    earning: 'success',
    deduction: 'warning',
    cost: 'info'
  };
  return map[cls] ?? 'default';
};

const resultColumns = [
  {
    title: $t('page.hrm.salary.batch.itemName'),
    key: 'itemName',
    minWidth: 120,
    render: (row: Api.Hrm.SalaryBatchItem) => `${row.itemName}（${row.itemCode}）`
  },
  {
    title: $t('page.hrm.salary.batch.itemClass'),
    key: 'itemClass',
    width: 90,
    render: (row: Api.Hrm.SalaryBatchItem) => (
      <NTag size="small" type={itemClassTag(row.itemClass)} bordered={false}>
        {$t(`page.hrm.salary.batch.itemClassEnum.${row.itemClass}`)}
      </NTag>
    )
  },
  {
    title: $t('page.hrm.salary.batch.amount'),
    key: 'amount',
    width: 120,
    align: 'right' as const
  },
  {
    title: $t('page.hrm.salary.batch.source'),
    key: 'source',
    width: 110,
    render: (row: Api.Hrm.SalaryBatchItem) => {
      const key = `page.hrm.salary.batch.sourceEnum.${row.source}` as App.I18n.I18nKey;
      return $t(key);
    }
  },
  {
    title: $t('page.hrm.salary.batch.sourceDetail'),
    key: 'sourceDetail',
    minWidth: 200,
    render: (row: Api.Hrm.SalaryBatchItem) => row.sourceDetail || '-'
  }
];

/** 按员工分组（工号升序、组内按计算顺序），分组表头 + 项目明细行 */
const grouped = ref<
  { key: string; employeeNo: string; employeeName: string; orgName: string; rows: Api.Hrm.SalaryBatchItem[] }[]
>([]);

function regroup() {
  const map = new Map<
    string,
    { key: string; employeeNo: string; employeeName: string; orgName: string; rows: Api.Hrm.SalaryBatchItem[] }
  >();
  for (const it of items.value) {
    const key = it.employeeUuid;
    if (!map.has(key)) {
      map.set(key, {
        key,
        employeeNo: it.employeeNo,
        employeeName: it.employeeName,
        orgName: it.orgName,
        rows: []
      });
    }
    map.get(key)!.rows.push(it);
  }
  grouped.value = [...map.values()];
}

async function load() {
  if (!props.batchUuid) return;
  loading.value = true;
  const { data, error } = await GetSalaryBatchResult(props.batchUuid);
  loading.value = false;
  if (error || !data) return;
  batch.value = data.batch;
  items.value = data.items;
  regroup();
}

watch(visible, val => {
  if (val) {
    items.value = [];
    grouped.value = [];
    batch.value = null;
    load();
  }
});
</script>

<template>
  <NModal v-model:show="visible" :title="$t('page.hrm.salary.batch.resultTitle')" preset="card" class="w-860px">
    <NSpin :show="loading">
      <NDescriptions v-if="batch" :column="4" size="small" label-placement="left" bordered>
        <NDescriptionsItem :label="$t('page.hrm.salary.batch.batchNo')">{{ batch.batchNo }}</NDescriptionsItem>
        <NDescriptionsItem :label="$t('page.hrm.salary.batch.batchName')">{{ batch.batchName }}</NDescriptionsItem>
        <NDescriptionsItem :label="$t('page.hrm.salary.batch.subject')">{{ batch.subjectName }}</NDescriptionsItem>
        <NDescriptionsItem :label="$t('page.hrm.salary.batch.period')">
          {{ batch.batchYear }}-{{ `${batch.batchMonth}`.padStart(2, '0') }}
        </NDescriptionsItem>
        <NDescriptionsItem :label="$t('page.hrm.salary.batch.headcount')">{{ batch.headcount }}</NDescriptionsItem>
        <NDescriptionsItem :label="$t('page.hrm.salary.batch.earningTotal')">{{ batch.earningTotal }}</NDescriptionsItem>
        <NDescriptionsItem :label="$t('page.hrm.salary.batch.deductionTotal')">{{ batch.deductionTotal }}</NDescriptionsItem>
        <NDescriptionsItem :label="$t('page.hrm.salary.batch.netTotal')">{{ batch.netTotal }}</NDescriptionsItem>
      </NDescriptions>
      <NAlert v-if="batch?.warnings" type="warning" :bordered="false" class="mt-8px whitespace-pre-wrap">
        {{ batch.warnings }}
      </NAlert>

      <div v-for="group in grouped" :key="group.key" class="mt-12px">
        <div class="mb-4px font-bold">
          {{ group.employeeName }}（{{ group.employeeNo }}）
          <span v-if="group.orgName" class="text-12px text-gray-400">{{ group.orgName }}</span>
        </div>
        <NDataTable
          size="small"
          :single-line="false"
          :columns="resultColumns"
          :data="group.rows"
          :pagination="false"
          :row-key="(row: Api.Hrm.SalaryBatchItem) => row.uuid"
        />
      </div>

      <NEmpty
        v-if="!loading && grouped.length === 0"
        :description="$t('page.hrm.salary.batch.noResult')"
        class="py-24px"
      />
    </NSpin>
  </NModal>
</template>
