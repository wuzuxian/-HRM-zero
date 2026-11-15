<script setup lang="ts">
import { $t } from '@/locales';
import { useNaiveForm } from '@/hooks/common/form';
import { hrmTaxDeductTypeOptions, hrmTaxDeductionSourceOptions } from '@/constants/business';
import { translateOptions } from '@/utils/common';

defineOptions({
  name: 'TaxDeductionSearch'
});

interface Emits {
  (e: 'reset'): void;
  (e: 'search'): void;
}

const emit = defineEmits<Emits>();

const { formRef, validate, restoreValidation } = useNaiveForm();

const model = defineModel<Api.Hrm.TaxSpecialDeductionSearchParams>('model', { required: true });

const yearOptions = (() => {
  const now = new Date().getFullYear();
  return [now + 1, now, now - 1, now - 2].map(y => ({ label: String(y), value: String(y) }));
})();

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
      <NCollapseItem :title="$t('common.search')" name="tax-deduction-search">
        <NForm ref="formRef" :model="model" label-placement="left" :label-width="80">
          <NGrid responsive="screen" item-responsive>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.salary.tax.deduction.employeeName')"
              path="employeeName"
              class="pr-24px"
            >
              <NInput
                v-model:value="model.employeeName"
                :placeholder="$t('page.hrm.salary.tax.deduction.form.employeeName')"
              />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.salary.tax.deduction.deductYear')"
              path="deductYear"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.deductYear"
                :options="yearOptions"
                :placeholder="$t('page.hrm.salary.tax.deduction.form.deductYear')"
                clearable
              />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.salary.tax.deduction.deductType')"
              path="deductType"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.deductType"
                :options="translateOptions(hrmTaxDeductTypeOptions)"
                :placeholder="$t('page.hrm.salary.tax.deduction.form.deductYear')"
                clearable
              />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.salary.tax.deduction.source')"
              path="source"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.source"
                :options="translateOptions(hrmTaxDeductionSourceOptions)"
                :placeholder="$t('page.hrm.salary.tax.deduction.form.employeeName')"
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
