<script setup lang="ts">
import { ref } from 'vue';
import { $t } from '@/locales';
import { useNaiveForm } from '@/hooks/common/form';
import { GetSalarySubjectList } from '@/service/api';
import { translateOptions } from '@/utils/common';

defineOptions({
  name: 'TaxOpeningSearch'
});

interface Emits {
  (e: 'reset'): void;
  (e: 'search'): void;
}

const emit = defineEmits<Emits>();

const { formRef, validate, restoreValidation } = useNaiveForm();

const model = defineModel<Api.Hrm.TaxOpeningSearchParams>('model', { required: true });

const yearOptions = (() => {
  const now = new Date().getFullYear();
  return [now + 1, now, now - 1, now - 2].map(y => ({ label: String(y), value: String(y) }));
})();

const subjectOptions = ref<{ label: string; value: string }[]>([]);

async function loadSubjectOptions() {
  if (subjectOptions.value.length > 0) return;
  const { data, error } = await GetSalarySubjectList({ current: 1, size: 200 });
  if (!error && data) {
    subjectOptions.value = (data.records ?? []).map(rec => ({
      label: rec.subjectName,
      value: rec.uuid
    }));
  }
}

loadSubjectOptions();

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
      <NCollapseItem :title="$t('common.search')" name="tax-opening-search">
        <NForm ref="formRef" :model="model" label-placement="left" :label-width="80">
          <NGrid responsive="screen" item-responsive>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.salary.tax.opening.employeeName')"
              path="employeeName"
              class="pr-24px"
            >
              <NInput
                v-model:value="model.employeeName"
                :placeholder="$t('page.hrm.salary.tax.opening.form.employeeName')"
              />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.salary.tax.opening.subjectName')"
              path="subjectUuid"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.subjectUuid"
                :options="subjectOptions"
                :placeholder="$t('page.hrm.salary.tax.opening.form.subjectName')"
                clearable
                filterable
              />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.salary.tax.opening.taxYear')"
              path="taxYear"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.taxYear"
                :options="yearOptions"
                :placeholder="$t('page.hrm.salary.tax.opening.form.taxYear')"
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
