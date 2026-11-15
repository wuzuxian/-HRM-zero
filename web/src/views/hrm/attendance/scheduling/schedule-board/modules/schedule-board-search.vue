<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { GetAttendanceGroupList, GetShiftList } from '@/service/api';
import { $t } from '@/locales';

defineOptions({
  name: 'ScheduleBoardSearch'
});

const emit = defineEmits<{
  (e: 'search'): void;
  (e: 'load'): void;
}>();

const groupUuid = defineModel<string>('groupUuid');
const month = defineModel<string>('month');
const model = defineModel<{ name: string; employeeNo: string; orgName: string }>('model');

const groupOptions = ref<{ label: string; value: string }[]>([]);
const shiftCount = ref(0);

async function loadGroups() {
  const { data, error } = await GetAttendanceGroupList({ current: 1, size: 100, status: '1' });
  if (error || !data) return;
  groupOptions.value = data.records.map(g => ({ label: g.groupName, value: g.uuid }));
}

onMounted(async () => {
  await loadGroups();
  // 默认月份 = 当前月
  if (!month.value) {
    const d = new Date();
    month.value = `${d.getFullYear()}-${`${d.getMonth() + 1}`.padStart(2, '0')}`;
  }
  emit('load');
});

function handleGroupChange() {
  emit('load');
}

function handleMonthChange() {
  emit('load');
}

function handleReset() {
  model.value = { name: '', employeeNo: '', orgName: '' };
  emit('search');
}
</script>

<template>
  <NCard :bordered="false" size="small" class="card-wrapper">
    <NDivider class="!mb-0 !mt-0" title-placement="left">
      {{ $t('page.hrm.attendance.scheduleBoard.searchTitle') }}
    </NDivider>
    <NForm label-placement="left" :label-width="80" class="flex flex-wrap items-center gap-8px">
      <NFormItem :label="$t('page.hrm.attendance.scheduleBoard.group')">
        <NSelect
          v-model:value="groupUuid"
          :options="groupOptions"
          :placeholder="$t('page.hrm.attendance.scheduleBoard.groupPlaceholder')"
          class="w-180px"
          @update:value="handleGroupChange"
        />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.attendance.scheduleBoard.month')">
        <NDatePicker
          v-model:formatted-value="month"
          type="month"
          :clearable="false"
          value-format="yyyy-MM"
          class="w-140px"
          @update:formatted-value="handleMonthChange"
        />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.attendance.scheduleBoard.employeeName')">
        <NInput v-model:value="model!.name" :placeholder="$t('page.hrm.attendance.scheduleBoard.employeeName')" class="w-120px" clearable />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.attendance.scheduleBoard.employeeNo')">
        <NInput v-model:value="model!.employeeNo" :placeholder="$t('page.hrm.attendance.scheduleBoard.employeeNo')" class="w-120px" clearable />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.attendance.scheduleBoard.orgName')">
        <NInput v-model:value="model!.orgName" :placeholder="$t('page.hrm.attendance.scheduleBoard.orgName')" class="w-120px" clearable />
      </NFormItem>
      <NButton class="ml-8px" type="primary" @click="emit('search')">
        {{ $t('common.search') }}
      </NButton>
      <NButton class="ml-8px" @click="handleReset">
        {{ $t('common.reset') }}
      </NButton>
    </NForm>
  </NCard>
</template>
