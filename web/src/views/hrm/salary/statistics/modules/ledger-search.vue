<script setup lang="ts">
import { $t } from '@/locales';
import { useNaiveForm } from '@/hooks/common/form';

defineOptions({
  name: 'SalaryLedgerSearch'
});

interface Emits {
  (e: 'reset'): void;
  (e: 'search'): void;
}

const emit = defineEmits<Emits>();

const { formRef, validate, restoreValidation } = useNaiveForm();

const model = defineModel<Api.Hrm.SalaryLedgerSearchParams>('model', { required: true });

const onlyAbnormalOptions = [
  { label: $t('common.yesOrNo.yes'), value: true },
  { label: $t('common.yesOrNo.no'), value: false }
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
      <NCollapseItem :title="$t('common.search')" name="salary-ledger-search">
        <NForm ref="formRef" :model="model" label-placement="left" :label-width="80">
          <NGrid responsive="screen" item-responsive>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.salary.statistics.batch')"
              path="batchUuid"
              class="pr-24px"
            >
              <NInput v-model:value="model.batchUuid" :placeholder="$t('page.hrm.salary.statistics.batchUuidTip')" />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.salary.statistics.employeeName')"
              path="employeeName"
              class="pr-24px"
            >
              <NInput v-model:value="model.employeeName" :placeholder="$t('page.hrm.salary.statistics.employeeName')" />
            </NFormItemGi>
            <NFormItemGi span="24 s:12 m:6" :label="$t('page.hrm.salary.statistics.org')" path="orgName" class="pr-24px">
              <NInput v-model:value="model.orgName" :placeholder="$t('page.hrm.salary.statistics.org')" />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.salary.statistics.onlyAbnormal')"
              path="onlyAbnormal"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.onlyAbnormal"
                :options="onlyAbnormalOptions"
                :placeholder="$t('page.hrm.salary.statistics.onlyAbnormal')"
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
