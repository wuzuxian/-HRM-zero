<script setup lang="ts">
import { ref } from 'vue';
import { $t } from '@/locales';
import { useNaiveForm } from '@/hooks/common/form';

defineOptions({
  name: 'PunchRecordSearch'
});

interface Emits {
  (e: 'reset'): void;
  (e: 'search'): void;
}

const emit = defineEmits<Emits>();

const { formRef, validate, restoreValidation } = useNaiveForm();

const model = defineModel<Api.Hrm.PunchSearchParams>('model', { required: true });

const clockStatusOptions = [
  { label: $t('page.hrm.attendance.punchRecord.statusNormal'), value: 'normal' },
  { label: $t('page.hrm.attendance.punchRecord.statusField'), value: 'field' },
  { label: $t('page.hrm.attendance.punchRecord.statusMissing'), value: 'missing' }
];

/** 日期区间（NDatePicker value 类型为 number|number[]，桥接为格式化字符串） */
const dateRange = ref<[number, number] | null>(null);

function handleRangeChange(val: number | [number, number] | null) {
  if (Array.isArray(val)) {
    dateRange.value = val;
    const fmt = (ts: number) => {
      const d = new Date(ts);
      const m = `${d.getMonth() + 1}`.padStart(2, '0');
      const day = `${d.getDate()}`.padStart(2, '0');
      return `${d.getFullYear()}-${m}-${day}`;
    };
    model.value.startDate = fmt(val[0]);
    model.value.endDate = fmt(val[1]);
  } else {
    dateRange.value = null;
    model.value.startDate = undefined;
    model.value.endDate = undefined;
  }
}

async function reset() {
  dateRange.value = null;
  model.value.startDate = undefined;
  model.value.endDate = undefined;
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
      <NCollapseItem :title="$t('common.search')" name="punch-record-search">
        <NForm ref="formRef" :model="model" label-placement="left" :label-width="80">
          <NGrid responsive="screen" item-responsive>
            <NFormItemGi
              span="24 s:12 m:8"
              :label="$t('page.hrm.attendance.punchRecord.dateRange')"
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
              :label="$t('page.hrm.attendance.punchRecord.name')"
              path="name"
              class="pr-24px"
            >
              <NInput v-model:value="model.name" :placeholder="$t('page.hrm.attendance.punchRecord.form.name')" />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.attendance.punchRecord.clockInStatus')"
              path="clockInStatus"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.clockInStatus"
                :options="clockStatusOptions"
                :placeholder="$t('page.hrm.attendance.punchRecord.form.clockInStatus')"
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
