<script setup lang="ts">
import { computed, ref } from 'vue';
import { GetLeaveTypeList } from '@/service/api';
import { $t } from '@/locales';
import { useNaiveForm } from '@/hooks/common/form';

defineOptions({
  name: 'LeaveBalanceSearch'
});

interface Emits {
  (e: 'reset'): void;
  (e: 'search'): void;
}

const emit = defineEmits<Emits>();

const { formRef, validate, restoreValidation } = useNaiveForm();

const model = defineModel<Api.Hrm.LeaveBalanceSearchParams>('model', { required: true });

interface LeaveTypeOption {
  label: string;
  value: string;
}

const leaveTypeOptions = ref<LeaveTypeOption[]>([]);

async function loadLeaveTypes() {
  if (leaveTypeOptions.value.length > 0) return;
  const { data, error } = await GetLeaveTypeList({ current: 1, size: 100 });
  if (!error && data) {
    leaveTypeOptions.value = (data.records ?? []).map(item => ({
      label: item.typeName,
      value: item.typeCode
    }));
  }
}
loadLeaveTypes();

const yearOptions = computed(() => {
  const current = new Date().getFullYear();
  const options: { label: string; value: number }[] = [];
  for (let y = current + 1; y >= current - 5; y -= 1) {
    options.push({ label: `${y}`, value: y });
  }
  return options;
});

async function reset() {
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
      <NCollapseItem :title="$t('common.search')" name="leave-balance-search">
        <NForm ref="formRef" :model="model" label-placement="left" :label-width="80">
          <NGrid responsive="screen" item-responsive>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.attendance.leaveBalance.year')"
              path="year"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.year"
                :options="yearOptions"
                :placeholder="$t('page.hrm.attendance.leaveBalance.form.year')"
                clearable
              />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.attendance.leaveBalance.name')"
              path="name"
              class="pr-24px"
            >
              <NInput v-model:value="model.name" :placeholder="$t('page.hrm.attendance.leaveBalance.form.name')" />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.attendance.leaveBalance.leaveTypeCode')"
              path="leaveTypeCode"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.leaveTypeCode"
                :options="leaveTypeOptions"
                :placeholder="$t('page.hrm.attendance.leaveBalance.form.leaveTypeCode')"
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
