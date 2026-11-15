<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { AddPerfTemplate, EditPerfTemplate, GetAllPerfIndicators, GetPerfTemplateDetail } from '@/service/api';
import { $t } from '@/locales';
import { perfStatusOptions } from '@/constants/hrm-performance';
import { translateOptions } from '@/utils/common';

defineOptions({
  name: 'PerfTemplateOperateModal'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Hrm.Performance.Template | null;
}

const props = defineProps<Props>();

interface Emits {
  (e: 'submitted'): void;
}

const emit = defineEmits<Emits>();

const visible = defineModel<boolean>('visible', {
  default: false
});

const { formRef, validate, restoreValidation } = useNaiveForm();
const { defaultRequiredRule } = useFormRules();
const { loading: confirmLoading, startLoading: confirmStartLoading, endLoading: confirmEndLoading } = useLoading();

const title = computed(() => {
  const titles: Record<NaiveUI.TableOperateType, string> = {
    add: $t('page.hrm.performance.template.addTemplate'),
    edit: $t('page.hrm.performance.template.editTemplate')
  };
  return titles[props.operateType];
});

/** 指标权重配置行（编辑草稿） */
type ItemDraft = {
  indicatorUuid: string | null;
  weight: number | null;
};

type Model = {
  name: string;
  status: Api.Hrm.Performance.CommonStatus;
  remark: string;
  items: ItemDraft[];
};

const model: Model = reactive(createDefaultModel());

function createDefaultModel(): Model {
  return {
    name: '',
    status: '1',
    remark: '',
    items: []
  };
}

type RuleKey = Extract<keyof Model, 'name'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  name: defaultRequiredRule
};

/** 指标库选项（模板内配置指标） */
const indicatorOptions = ref<CommonType.Option<string>[]>([]);

async function getIndicatorOptions() {
  const { error, data } = await GetAllPerfIndicators();
  if (!error) {
    indicatorOptions.value = data.records.map(item => ({
      label: `${item.code} ${item.name}`,
      value: item.uuid
    }));
  }
}

/** 权重之和（百分比）— 模板权重校验：必须等于 100% */
const weightSum = computed(() => {
  const sum = model.items.reduce((acc, item) => acc + (item.weight ?? 0), 0);
  return Math.round(sum * 100) / 100;
});

const weightSumValid = computed(() => model.items.length === 0 || weightSum.value === 100);

function addItem() {
  model.items.push({ indicatorUuid: null, weight: null });
}

function removeItem(index: number) {
  model.items.splice(index, 1);
}

function handleInitModel() {
  Object.assign(model, createDefaultModel());

  if (props.operateType === 'edit' && props.rowData?.uuid) {
    GetPerfTemplateDetail(props.rowData.uuid).then(({ error, data }) => {
      if (!error && data) {
        model.name = data.name;
        model.status = (data.status ?? '1') as Api.Hrm.Performance.CommonStatus;
        model.remark = data.remark;
        model.items = data.items.map(item => ({
          indicatorUuid: item.indicatorUuid,
          weight: item.weight
        }));
      }
    });
  }
}

function closeModel() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();

  if (!weightSumValid.value) {
    window.$message?.error($t('page.hrm.performance.template.weightSumError'));
    return;
  }
  // 指标权重配置行必须完整：已选指标 + 权重大于 0
  for (const item of model.items) {
    if (!item.indicatorUuid || item.weight === null || item.weight <= 0) {
      window.$message?.error($t('page.hrm.performance.template.itemIncomplete'));
      return;
    }
  }

  const items: Api.Hrm.Performance.TemplateItemRequest[] = model.items.map(item => ({
    indicatorUuid: item.indicatorUuid as string,
    weight: item.weight as number
  }));

  confirmStartLoading();
  const { error } =
    props.operateType === 'add'
      ? await AddPerfTemplate({
          name: model.name,
          status: model.status,
          remark: model.remark,
          items
        })
      : await EditPerfTemplate({
          uuid: props.rowData?.uuid,
          name: model.name,
          status: model.status,
          remark: model.remark,
          items
        });
  if (!error) {
    window.$message?.success(props.operateType === 'add' ? $t('common.addSuccess') : $t('common.updateSuccess'));
    closeModel();
    emit('submitted');
  }
  confirmEndLoading();
}

watch(visible, () => {
  if (visible.value) {
    handleInitModel();
    restoreValidation();
    getIndicatorOptions();
  }
});
</script>

<template>
  <NModal v-model:show="visible" :title="title" preset="card" class="w-720px">
    <NScrollbar class="h-480px pr-20px">
      <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="100">
        <NFormItem :label="$t('page.hrm.performance.template.name')" path="name">
          <NInput v-model:value="model.name" :placeholder="$t('page.hrm.performance.template.form.name')" />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.performance.template.status')" path="status">
          <NRadioGroup v-model:value="model.status">
            <NRadio
              v-for="item in translateOptions(perfStatusOptions)"
              :key="item.value"
              :value="item.value"
              :label="item.label"
            />
          </NRadioGroup>
        </NFormItem>
        <NFormItem :label="$t('page.hrm.performance.template.remark')" path="remark">
          <NInput v-model:value="model.remark" :placeholder="$t('page.hrm.performance.template.form.remark')" />
        </NFormItem>

        <NDivider :title="$t('page.hrm.performance.template.configItems')" title-placement="left" />
        <NAlert v-if="model.items.length > 0 && !weightSumValid" type="warning" class="mb-12px">
          {{ $t('page.hrm.performance.template.weightSumError') }}（{{ weightSum }}%）
        </NAlert>
        <div v-for="(item, index) in model.items" :key="index" class="mb-12px flex items-center gap-12px">
          <NSelect
            v-model:value="item.indicatorUuid"
            class="flex-1"
            :placeholder="$t('page.hrm.performance.template.indicatorPlaceholder')"
            :options="indicatorOptions"
            filterable
          />
          <NInputNumber v-model:value="item.weight" :min="0" :max="100" :precision="2" class="w-160px">
            <template #suffix>%</template>
          </NInputNumber>
          <NButton quaternary type="error" @click="() => removeItem(index)">
            {{ $t('common.delete') }}
          </NButton>
        </div>
        <NButton size="small" dashed block @click="addItem">
          {{ $t('page.hrm.performance.template.addIndicatorItem') }}
        </NButton>
        <div class="mt-12px text-right">
          {{ $t('page.hrm.performance.template.weightSumLabel') }}:
          <span :class="weightSumValid ? 'text-success' : 'text-error'">{{ weightSum }}%</span>
        </div>
      </NForm>
    </NScrollbar>
    <template #footer>
      <NSpace :size="16" justify="end">
        <NButton @click="closeModel">{{ $t('common.cancel') }}</NButton>
        <NButton type="primary" :loading="confirmLoading" @click="handleSubmit">{{ $t('common.confirm') }}</NButton>
      </NSpace>
    </template>
  </NModal>
</template>

<style scoped></style>
