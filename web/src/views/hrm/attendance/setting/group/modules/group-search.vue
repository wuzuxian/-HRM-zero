<script setup lang="ts">
import { $t } from '@/locales';
import { useNaiveForm } from '@/hooks/common/form';
import { hrmAttendanceStatusOptions } from '@/constants/business';
import { translateOptions } from '@/utils/common';

defineOptions({
  name: 'AttendanceGroupSearch'
});

interface Emits {
  (e: 'reset'): void;
  (e: 'search'): void;
}

const emit = defineEmits<Emits>();

const { formRef, validate, restoreValidation } = useNaiveForm();

const model = defineModel<Api.Hrm.AttendanceGroupSearchParams>('model', { required: true });

/** 班制 options (固定班制/排班制/自由工时) */
const groupTypeOptions: CommonType.Option<string>[] = [
  { value: '1', label: 'page.hrm.attendance.group.groupType.fixed' },
  { value: '2', label: 'page.hrm.attendance.group.groupType.scheduling' },
  { value: '3', label: 'page.hrm.attendance.group.groupType.flexible' }
];

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
      <NCollapseItem :title="$t('common.search')" name="group-search">
        <NForm ref="formRef" :model="model" label-placement="left" :label-width="80">
          <NGrid responsive="screen" item-responsive>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.attendance.group.groupName')"
              path="groupName"
              class="pr-24px"
            >
              <NInput v-model:value="model.groupName" :placeholder="$t('page.hrm.attendance.group.form.groupName')" />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.attendance.group.groupType.title')"
              path="groupType"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.groupType"
                :placeholder="$t('page.hrm.attendance.group.form.groupType')"
                :options="translateOptions(groupTypeOptions)"
                clearable
              />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.attendance.group.status')"
              path="status"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.status"
                :placeholder="$t('page.hrm.attendance.group.form.status')"
                :options="translateOptions(hrmAttendanceStatusOptions)"
                clearable
              />
            </NFormItemGi>
            <NFormItemGi span="24 m:6" class="pr-24px">
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
