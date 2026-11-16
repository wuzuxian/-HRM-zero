<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { $t } from '@/locales';
import { useNaiveForm } from '@/hooks/common/form';
import { hrmSalaryStatusOptions } from '@/constants/business';
import { GetSalarySubjectList } from '@/service/api';
import { translateOptions } from '@/utils/common';

defineOptions({
  name: 'SalaryBatchSearch'
});

interface Emits {
  (e: 'reset'): void;
  (e: 'search'): void;
}

const emit = defineEmits<Emits>();

const { formRef, validate, restoreValidation } = useNaiveForm();

const model = defineModel<Api.Hrm.SalaryBatchSearchParams>('model', { required: true });

const subjectOptions = ref<CommonType.Option<string>[]>([]);

async function loadFilterOptions() {
  const { data: subjectData, error: subjectError } = await GetSalarySubjectList({ current: 1, size: 500 });
  if (!subjectError && subjectData) {
    subjectOptions.value = subjectData.records.map(row => ({ label: row.subjectName, value: row.uuid }));
  }
}

onMounted(() => {
  loadFilterOptions();
});

const yearOptions = Array.from({ length: 7 }, (_, i) => {
  const y = new Date().getFullYear() - 5 + i;
  return { label: `${y}`, value: y };
});
const monthOptions = Array.from({ length: 12 }, (_, i) => ({ label: `${i + 1}`, value: i + 1 }));

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
      <NCollapseItem :title="$t('common.search')" name="salary-batch-search">
        <NForm ref="formRef" :model="model" label-placement="left" :label-width="80">
          <NGrid responsive="screen" item-responsive>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.salary.batch.batchName')"
              path="batchName"
              class="pr-24px"
            >
              <NInput v-model:value="model.batchName" :placeholder="$t('page.hrm.salary.batch.form.batchName')" />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.salary.batch.subject')"
              path="subjectUuid"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.subjectUuid"
                :options="subjectOptions"
                :placeholder="$t('page.hrm.salary.batch.form.subject')"
                filterable
                clearable
              />
            </NFormItemGi>
            <NFormItemGi span="24 s:12 m:6" :label="$t('page.hrm.salary.batch.year')" path="year" class="pr-24px">
              <NSelect v-model:value="model.year" :options="yearOptions" clearable />
            </NFormItemGi>
            <NFormItemGi span="24 s:12 m:6" :label="$t('page.hrm.salary.batch.month')" path="month" class="pr-24px">
              <NSelect v-model:value="model.month" :options="monthOptions" clearable />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.salary.batch.trialStatus')"
              path="trialStatus"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.trialStatus"
                :options="translateOptions(hrmSalaryStatusOptions)"
                :placeholder="$t('page.hrm.salary.batch.form.trialStatus')"
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
