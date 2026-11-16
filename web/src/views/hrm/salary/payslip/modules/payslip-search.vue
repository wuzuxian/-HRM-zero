<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { $t } from '@/locales';
import { useNaiveForm } from '@/hooks/common/form';
import { GetSalaryBatchList } from '@/service/api';
import { translateOptions } from '@/utils/common';

defineOptions({
  name: 'SalaryPayslipSearch'
});

interface Emits {
  (e: 'reset'): void;
  (e: 'search'): void;
}

const emit = defineEmits<Emits>();

const { formRef, validate, restoreValidation } = useNaiveForm();

const model = defineModel<Api.Hrm.SalaryPayslipSearchParams>('model', { required: true });

// 工资条年月下拉（近 24 个月）与已读状态枚举
const yearMonthOptions = (() => {
  const out: CommonType.Option<string>[] = [];
  const now = new Date();
  for (let i = 0; i < 24; i += 1) {
    const d = new Date(now.getFullYear(), now.getMonth() - i, 1);
    const ym = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`;
    out.push({ label: ym, value: ym });
  }
  return out;
})();

const readStatusOptions: CommonType.Option<string>[] = [
  { label: $t('page.hrm.salary.payslip.unread'), value: '0' },
  { label: $t('page.hrm.salary.payslip.readLabel'), value: '1' }
];

// 筛选项: 薪资批次下拉（全量拉取，分页 size 放大取一页）
const batchOptions = ref<CommonType.Option<string>[]>([]);

async function loadBatchOptions() {
  const { data, error } = await GetSalaryBatchList({ current: 1, size: 500 });
  if (!error && data) {
    batchOptions.value = data.records.map(row => ({ label: `${row.batchNo} ${row.batchName}`, value: row.uuid }));
  }
}

onMounted(() => {
  loadBatchOptions();
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
      <NCollapseItem :title="$t('common.search')" name="salary-payslip-search">
        <NForm ref="formRef" :model="model" label-placement="left" :label-width="80">
          <NGrid responsive="screen" item-responsive>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.salary.payslip.batch')"
              path="batchUuid"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.batchUuid"
                :options="batchOptions"
                :placeholder="$t('page.hrm.salary.payslip.chooseBatchPlaceholder')"
                filterable
                clearable
              />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.salary.payslip.employeeName')"
              path="employeeName"
              class="pr-24px"
            >
              <NInput v-model:value="model.employeeName" :placeholder="$t('page.hrm.salary.payslip.employeeName')" />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.salary.payslip.yearMonth')"
              path="yearMonth"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.yearMonth"
                :options="yearMonthOptions"
                :placeholder="$t('page.hrm.salary.payslip.yearMonth')"
                filterable
                clearable
              />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.salary.payslip.readStatus')"
              path="readStatus"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.readStatus"
                :options="translateOptions(readStatusOptions)"
                :placeholder="$t('page.hrm.salary.payslip.readStatus')"
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
