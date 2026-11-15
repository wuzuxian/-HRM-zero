<script setup lang="ts">
import { ref } from 'vue';
import { GetLeaveTypeList } from '@/service/api';
import { $t } from '@/locales';
import { useNaiveForm } from '@/hooks/common/form';
import { hrmBillStatusOptions } from '@/constants/business';
import { translateOptions } from '@/utils/common';

defineOptions({
  name: 'LeaveCancelSearch'
});

interface Emits {
  (e: 'reset'): void;
  (e: 'search'): void;
}

const emit = defineEmits<Emits>();

const { formRef, validate, restoreValidation } = useNaiveForm();

const model = defineModel<Api.Hrm.LeaveSearchParams>('model', { required: true });

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
      <NCollapseItem :title="$t('common.search')" name="leave-cancel-search">
        <NForm ref="formRef" :model="model" label-placement="left" :label-width="80">
          <NGrid responsive="screen" item-responsive>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.attendance.leaveCancel.billNo')"
              path="billNo"
              class="pr-24px"
            >
              <NInput
                v-model:value="model.billNo"
                :placeholder="$t('page.hrm.attendance.leaveCancel.form.billNo')"
              />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.attendance.billStatus.title')"
              path="status"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.status"
                :placeholder="$t('page.hrm.attendance.leaveCancel.form.status')"
                :options="translateOptions(hrmBillStatusOptions)"
                clearable
              />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.attendance.leaveCancel.leaveTypeName')"
              path="leaveTypeCode"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.leaveTypeCode"
                :options="leaveTypeOptions"
                :placeholder="$t('page.hrm.attendance.leaveCancel.form.leaveTypeCode')"
                clearable
              />
            </NFormItemGi>
            <NFormItemGi span="24 m:12" class="pr-24px">
              <NSpace class="w-full" justify="end">
                <NButton @click="reset">
                  <template #icon>
                    <icon-ic-round-refresh class="text-icon" />
                  </template>
                  {{ $t('common.reset') }}
                </NButton>
                <NButton type="primary" ghost @click="search">
                  <template #icon>
                    <icon-ic-round-search class="text-icon" />
                  </template>
                  {{ $t('common.search') }}
                </NButton>
              </NSpace>
            </NFormItemGi>
          </NGrid>
        </NForm>
      </NCollapseItem>
    </NCollapse>
  </NCard>
</template>

<style scoped></style>
