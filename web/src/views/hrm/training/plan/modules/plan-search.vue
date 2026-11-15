<script setup lang="ts">
import { $t } from '@/locales';
import { useNaiveForm } from '@/hooks/common/form';
import {
  hrmBillStatusOptions,
  hrmTrainingPlanPublishedOptions,
  hrmTrainingPlanTypeOptions
} from '@/constants/business';
import { translateOptions } from '@/utils/common';

defineOptions({
  name: 'PlanSearch'
});

interface Emits {
  (e: 'reset'): void;
  (e: 'search'): void;
}

const emit = defineEmits<Emits>();

const { formRef, validate, restoreValidation } = useNaiveForm();

const model = defineModel<Api.Hrm.TrainingPlanSearchParams>('model', { required: true });

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
      <NCollapseItem :title="$t('common.search')" name="plan-search">
        <NForm ref="formRef" :model="model" label-placement="left" :label-width="80">
          <NGrid responsive="screen" item-responsive>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.training.plan.planName')"
              path="planName"
              class="pr-24px"
            >
              <NInput v-model:value="model.planName" :placeholder="$t('page.hrm.training.plan.form.planName')" />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.training.plan.planType')"
              path="planType"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.planType"
                :placeholder="$t('page.hrm.training.plan.form.planType')"
                :options="translateOptions(hrmTrainingPlanTypeOptions)"
                clearable
              />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.training.plan.billStatusLabel')"
              path="status"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.status"
                :placeholder="$t('page.hrm.training.plan.form.status')"
                :options="translateOptions(hrmBillStatusOptions)"
                clearable
              />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.training.plan.published')"
              path="published"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.published"
                :placeholder="$t('page.hrm.training.plan.form.published')"
                :options="translateOptions(hrmTrainingPlanPublishedOptions)"
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
