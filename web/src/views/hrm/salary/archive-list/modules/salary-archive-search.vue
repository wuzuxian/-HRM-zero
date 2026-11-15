<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { $t } from '@/locales';
import { useNaiveForm } from '@/hooks/common/form';
import { hrmSalaryStatusOptions } from '@/constants/business';
import { GetSalarySchemeList, GetSalarySubjectList } from '@/service/api';
import { translateOptions } from '@/utils/common';

defineOptions({
  name: 'SalaryArchiveSearch'
});

interface Emits {
  (e: 'reset'): void;
  (e: 'search'): void;
}

const emit = defineEmits<Emits>();

const { formRef, validate, restoreValidation } = useNaiveForm();

const model = defineModel<Api.Hrm.SalaryArchiveSearchParams>('model', { required: true });

// 筛选项: 薪酬主体 / 薪酬方案 下拉（全量拉取，分页 size 放大取一页）
const subjectOptions = ref<CommonType.Option<string>[]>([]);
const schemeOptions = ref<CommonType.Option<string>[]>([]);

async function loadFilterOptions() {
  const { data: subjectData, error: subjectError } = await GetSalarySubjectList({ current: 1, size: 500 });
  if (!subjectError && subjectData) {
    subjectOptions.value = subjectData.records.map(row => ({ label: row.subjectName, value: row.uuid }));
  }
  const { data: schemeData, error: schemeError } = await GetSalarySchemeList({ current: 1, size: 500 });
  if (!schemeError && schemeData) {
    schemeOptions.value = schemeData.records.map(row => ({ label: row.schemeName, value: row.uuid }));
  }
}

onMounted(() => {
  loadFilterOptions();
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
      <NCollapseItem :title="$t('common.search')" name="salary-archive-search">
        <NForm ref="formRef" :model="model" label-placement="left" :label-width="80">
          <NGrid responsive="screen" item-responsive>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.salary.archive.employeeName')"
              path="employeeName"
              class="pr-24px"
            >
              <NInput
                v-model:value="model.employeeName"
                :placeholder="$t('page.hrm.salary.archive.form.employeeName')"
              />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.salary.archive.employeeNo')"
              path="employeeNo"
              class="pr-24px"
            >
              <NInput
                v-model:value="model.employeeNo"
                :placeholder="$t('page.hrm.salary.archive.form.employeeNo')"
              />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.salary.archive.subject')"
              path="subjectUuid"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.subjectUuid"
                :options="subjectOptions"
                :placeholder="$t('page.hrm.salary.archive.form.subject')"
                filterable
                clearable
              />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.salary.archive.scheme')"
              path="schemeUuid"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.schemeUuid"
                :options="schemeOptions"
                :placeholder="$t('page.hrm.salary.archive.form.scheme')"
                filterable
                clearable
              />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.salary.archive.status')"
              path="status"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.status"
                :placeholder="$t('page.hrm.salary.archive.form.status')"
                :options="translateOptions(hrmSalaryStatusOptions)"
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
