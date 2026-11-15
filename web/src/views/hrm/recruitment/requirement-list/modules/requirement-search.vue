<script setup lang="ts">
import { $t } from '@/locales';
import { useNaiveForm } from '@/hooks/common/form';
import {
  hrmBillStatusOptions,
  hrmRequirementPostCategoryOptions,
  hrmRequirementUrgencyOptions
} from '@/constants/business';
import { translateOptions } from '@/utils/common';

defineOptions({
  name: 'RequirementSearch'
});

interface Emits {
  (e: 'reset'): void;
  (e: 'search'): void;
}

const emit = defineEmits<Emits>();

const { formRef, validate, restoreValidation } = useNaiveForm();

const model = defineModel<Api.Hrm.RequirementSearchParams>('model', { required: true });

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
      <NCollapseItem :title="$t('common.search')" name="requirement-search">
        <NForm ref="formRef" :model="model" label-placement="left" :label-width="90">
          <NGrid responsive="screen" item-responsive>
            <NFormItemGi span="24 s:12 m:6" :label="$t('page.hrm.requirement.billNo')" path="billNo" class="pr-24px">
              <NInput v-model:value="model.billNo" :placeholder="$t('page.hrm.requirement.form.billNo')" />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.requirement.billStatus.title')"
              path="status"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.status"
                :placeholder="$t('page.hrm.requirement.form.status')"
                :options="translateOptions(hrmBillStatusOptions)"
                clearable
              />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.requirement.postName')"
              path="postName"
              class="pr-24px"
            >
              <NInput v-model:value="model.postName" :placeholder="$t('page.hrm.requirement.form.postName')" />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.requirement.postCategory.title')"
              path="postCategory"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.postCategory"
                :placeholder="$t('page.hrm.requirement.form.postCategory')"
                :options="translateOptions(hrmRequirementPostCategoryOptions)"
                clearable
              />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.requirement.urgency.title')"
              path="urgency"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.urgency"
                :placeholder="$t('page.hrm.requirement.form.urgency')"
                :options="translateOptions(hrmRequirementUrgencyOptions)"
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
