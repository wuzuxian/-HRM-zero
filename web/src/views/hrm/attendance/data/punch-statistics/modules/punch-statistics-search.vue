<script setup lang="ts">
import { ref } from 'vue';
import { $t } from '@/locales';
import { useNaiveForm } from '@/hooks/common/form';

defineOptions({
  name: 'PunchStatisticsSearch'
});

interface Emits {
  (e: 'reset'): void;
  (e: 'search'): void;
}

const emit = defineEmits<Emits>();

const { formRef, validate, restoreValidation } = useNaiveForm();

const model = defineModel<Api.Hrm.PunchStatsSearchParams>('model', { required: true });

/** 日期区间（NDatePicker value 类型为 number|number[]，桥接为格式化字符串） */
const dateRange = ref<[number, number] | null>(null);

function fmtTs(ts: number) {
  const d = new Date(ts);
  const m = `${d.getMonth() + 1}`.padStart(2, '0');
  const day = `${d.getDate()}`.padStart(2, '0');
  return `${d.getFullYear()}-${m}-${day}`;
}

function handleRangeChange(val: number | [number, number] | null) {
  if (Array.isArray(val)) {
    dateRange.value = val;
    model.value.startDate = fmtTs(val[0]);
    model.value.endDate = fmtTs(val[1]);
  } else {
    dateRange.value = null;
    model.value.startDate = undefined;
    model.value.endDate = undefined;
  }
}

const onlyAbnormalOptions = [
  { label: $t('page.hrm.attendance.punchStatistics.onlyAbnormalYes'), value: '1' }
];

async function reset() {
  dateRange.value = null;
  model.value.startDate = undefined;
  model.value.endDate = undefined;
  model.value.onlyAbnormal = undefined;
  await restoreValidation();
  emit('reset');
}

async function search() {
  await validate();
  emit('search');
}
</script>

<template>
  <NCard :bordered="false" size="small" class="card-wrapper">
    <NCollapse>
      <NCollapseItem :title="$t('common.search')" name="punch-statistics-search">
        <NForm ref="formRef" :model="model" label-placement="left" :label-width="80">
          <NGrid responsive="screen" item-responsive>
            <NFormItemGi
              span="24 s:12 m:8"
              :label="$t('page.hrm.attendance.punchStatistics.dateRange')"
              path="startDate"
              class="pr-24px"
            >
              <NDatePicker
                :value="dateRange"
                type="daterange"
                clearable
                value-format="yyyy-MM-dd"
                @update:value="handleRangeChange"
              />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.attendance.punchStatistics.name')"
              path="name"
              class="pr-24px"
            >
              <NInput v-model:value="model.name" :placeholder="$t('page.hrm.attendance.punchStatistics.form.name')" />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.attendance.punchStatistics.onlyAbnormal')"
              path="onlyAbnormal"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.onlyAbnormal"
                :options="onlyAbnormalOptions"
                :placeholder="$t('page.hrm.attendance.punchStatistics.form.onlyAbnormal')"
                clearable
              />
            </NFormItemGi>
            <NFormItemGi span="24 m:12" class="pr-24px">
              <NSpace class="w-full" justify="end">
                <NButton @click="reset">{{ $t('common.reset') }}</NButton>
                <NButton type="primary" ghost @click="search">{{ $t('common.search') }}</NButton>
              </NSpace>
            </NFormItemGi>
          </NGrid>
        </NForm>
      </NCollapseItem>
    </NCollapse>
  </NCard>
</template>
