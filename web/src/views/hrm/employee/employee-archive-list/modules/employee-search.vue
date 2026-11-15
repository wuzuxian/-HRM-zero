<script setup lang="ts">
import { computed } from 'vue';
import { $t } from '@/locales';
import { useNaiveForm } from '@/hooks/common/form';
import { hrmEmployeePersonStatusOptions } from '@/constants/business';
import { translateOptions } from '@/utils/common';

defineOptions({
  name: 'EmployeeSearch'
});

interface Props {
  /** flat org options (uuid/orgName) for 部门筛选 */
  orgOptions?: Api.Hrm.Org[];
}

const props = defineProps<Props>();

interface Emits {
  (e: 'reset'): void;
  (e: 'search'): void;
}

const emit = defineEmits<Emits>();

const { formRef, validate, restoreValidation } = useNaiveForm();

const model = defineModel<Api.Hrm.EmployeeSearchParams>('model', { required: true });

const orgSelectOptions = computed(() =>
  (props.orgOptions ?? []).map(item => ({ label: item.orgName, value: item.uuid }))
);

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
      <NCollapseItem :title="$t('common.search')" name="employee-search">
        <NForm ref="formRef" :model="model" label-placement="left" :label-width="80">
          <NGrid responsive="screen" item-responsive>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.employee.employeeNo')"
              path="employeeNo"
              class="pr-24px"
            >
              <NInput
                v-model:value="model.employeeNo"
                :placeholder="$t('page.hrm.employee.form.employeeNo')"
                clearable
              />
            </NFormItemGi>
            <NFormItemGi span="24 s:12 m:6" :label="$t('page.hrm.employee.name')" path="name" class="pr-24px">
              <NInput v-model:value="model.name" :placeholder="$t('page.hrm.employee.form.name')" clearable />
            </NFormItemGi>
            <NFormItemGi span="24 s:12 m:6" :label="$t('page.hrm.employee.orgName')" path="orgUuid" class="pr-24px">
              <NSelect
                v-model:value="model.orgUuid"
                :placeholder="$t('page.hrm.employee.form.org')"
                :options="orgSelectOptions"
                clearable
                filterable
              />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.employee.personStatusLabel')"
              path="personStatus"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.personStatus"
                :placeholder="$t('page.hrm.employee.form.personStatus')"
                :options="translateOptions(hrmEmployeePersonStatusOptions)"
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
