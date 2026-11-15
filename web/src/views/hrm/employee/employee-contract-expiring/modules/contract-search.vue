<script setup lang="ts">
import { $t } from '@/locales';
import { useNaiveForm } from '@/hooks/common/form';
import { contractExpiryStateOptions } from '@/constants/business';
import { translateOptions } from '@/utils/common';

defineOptions({
  name: 'ContractSearch'
});

interface Emits {
  (e: 'reset'): void;
  (e: 'search'): void;
}

const emit = defineEmits<Emits>();

const { formRef, validate, restoreValidation } = useNaiveForm();

const model = defineModel<Api.Hrm.ContractSearchParams>('model', { required: true });

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
      <NCollapseItem :title="$t('common.search')" name="contract-search">
        <NForm ref="formRef" :model="model" label-placement="left" :label-width="90">
          <NGrid responsive="screen" item-responsive>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.employee.contractExpiring.employeeName')"
              path="employeeName"
              class="pr-24px"
            >
              <NInput
                v-model:value="model.employeeName"
                :placeholder="$t('page.hrm.employee.contractExpiring.form.employeeName')"
              />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.employee.contractExpiring.employeeNo')"
              path="employeeNo"
              class="pr-24px"
            >
              <NInput
                v-model:value="model.employeeNo"
                :placeholder="$t('page.hrm.employee.contractExpiring.form.employeeNo')"
              />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:8"
              :label="$t('page.hrm.employee.contractExpiring.endDate')"
              path="endDate"
              class="pr-24px"
            >
              <NDatePicker
                v-model:formatted-value="model.startDate"
                type="date"
                value-format="yyyy-MM-dd"
                clearable
                class="w-full"
              />
              <span class="mx-4px">~</span>
              <NDatePicker
                v-model:formatted-value="model.endDate"
                type="date"
                value-format="yyyy-MM-dd"
                clearable
                class="w-full"
              />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.employee.contractExpiring.status')"
              path="expiryState"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.expiryState"
                :placeholder="$t('page.hrm.employee.contractExpiring.form.status')"
                :options="translateOptions(contractExpiryStateOptions)"
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
