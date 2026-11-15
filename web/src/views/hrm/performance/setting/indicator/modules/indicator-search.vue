<script setup lang="ts">
import { $t } from '@/locales';
import {
  perfDataSourceOptions,
  perfIndicatorTypeOptions,
  perfScoreRuleOptions,
  perfStatusOptions
} from '@/constants/hrm-performance';
import { translateOptions } from '@/utils/common';

defineOptions({
  name: 'PerfIndicatorSearch'
});

interface Emits {
  (e: 'reset'): void;
  (e: 'search'): void;
}

const emit = defineEmits<Emits>();

const model = defineModel<Api.Hrm.Performance.IndicatorSearchParams>('model', { required: true });
</script>

<template>
  <NCard :bordered="false" size="small" class="card-wrapper">
    <NCollapse>
      <NCollapseItem :title="$t('common.search')" name="perf-indicator-search">
        <NForm :model="model" label-placement="left" :label-width="80">
          <NGrid responsive="screen" item-responsive>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.performance.indicator.code')"
              path="code"
              class="pr-24px"
            >
              <NInput v-model:value="model.code" :placeholder="$t('page.hrm.performance.indicator.form.code')" />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.performance.indicator.name')"
              path="name"
              class="pr-24px"
            >
              <NInput v-model:value="model.name" :placeholder="$t('page.hrm.performance.indicator.form.name')" />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.performance.indicator.indicatorType')"
              path="indicatorType"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.indicatorType"
                :placeholder="$t('page.hrm.performance.indicator.form.indicatorType')"
                :options="translateOptions(perfIndicatorTypeOptions)"
                clearable
              />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.performance.indicator.scoreRule')"
              path="scoreRule"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.scoreRule"
                :placeholder="$t('page.hrm.performance.indicator.form.scoreRule')"
                :options="translateOptions(perfScoreRuleOptions)"
                clearable
              />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.performance.indicator.dataSource')"
              path="dataSource"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.dataSource"
                :placeholder="$t('page.hrm.performance.indicator.form.dataSource')"
                :options="translateOptions(perfDataSourceOptions)"
                clearable
              />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.performance.indicator.status')"
              path="status"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.status"
                :placeholder="$t('page.hrm.performance.indicator.form.status')"
                :options="translateOptions(perfStatusOptions)"
                clearable
              />
            </NFormItemGi>
            <NFormItemGi span="24 m:12" class="pr-24px">
              <NSpace class="w-full" justify="end">
                <NButton @click="() => emit('reset')">
                  <template #icon>
                    <icon-ic-round-refresh class="text-icon" />
                  </template>
                  {{ $t('common.reset') }}
                </NButton>
                <NButton type="primary" ghost @click="() => emit('search')">
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
