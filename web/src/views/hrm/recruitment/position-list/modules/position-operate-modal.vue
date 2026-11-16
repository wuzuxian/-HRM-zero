<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { AddPosition, EditPosition, GetRequirementList } from '@/service/api';
import { $t } from '@/locales';
import { hrmPositionPostCategoryOptions } from '@/constants/business';
import { translateOptions } from '@/utils/common';

defineOptions({
  name: 'PositionOperateModal'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Hrm.HrmPosition | null;
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
    add: $t('page.hrm.position.addPosition'),
    edit: $t('page.hrm.position.editPosition')
  };
  return titles[props.operateType];
});

type Model = Omit<Api.Hrm.AddPositionRequest, 'publishDate'> & {
  // NDatePicker 的 formatted-value 不能是空串（解析 "" 抛 Invalid time value），未选择时必须为 null
  publishDate: string | null;
};

const model: Model = reactive(createDefaultModel());

function createDefaultModel(): Model {
  return {
    postName: '',
    requirementUuid: '',
    postCategory: 'tech',
    ownerUuid: '',
    ownerName: '',
    headcount: 1,
    publishDate: null,
    remark: ''
  };
}

type RuleKey = Extract<keyof Model, 'postName' | 'requirementUuid' | 'postCategory' | 'headcount'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  postName: defaultRequiredRule,
  requirementUuid: defaultRequiredRule,
  postCategory: defaultRequiredRule,
  headcount: defaultRequiredRule
};

// ---- 来源需求单 options（仅 status=approved 可引用，HRM-24 seam 契约） ----
interface RequirementOption {
  label: string;
  value: string;
}

const requirementOptions = ref<RequirementOption[]>([]);

async function loadRequirementOptions() {
  const { data, error } = await GetRequirementList({
    current: 1,
    size: 100,
    status: 'approved'
  });
  if (!error && data) {
    requirementOptions.value = (data.records ?? []).map(rec => ({
      label: `${rec.billNo} ${rec.postName}`,
      value: rec.uuid
    }));
  }
}

function handleInitModel() {
  Object.assign(model, createDefaultModel());

  if (props.operateType === 'edit' && props.rowData) {
    Object.assign(model, props.rowData);
    // DB 列 NOT NULL DEFAULT ''，编辑回填可能拿到空串；空串会让 NDatePicker 抛 Invalid time value
    model.publishDate = props.rowData.publishDate || null;
  }
  loadRequirementOptions();
}

function closeModel() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();
  confirmStartLoading();

  const request =
    props.operateType === 'add'
      ? AddPosition({ ...model, publishDate: model.publishDate ?? '' })
      : EditPosition({ ...model, publishDate: model.publishDate ?? '', uuid: props.rowData?.uuid ?? '' });

  const { error } = await request;
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
  }
});
</script>

<template>
  <NModal v-model:show="visible" :title="title" preset="card" class="w-760px">
    <NScrollbar class="max-h-480px pr-20px">
      <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="100">
        <NGrid responsive="screen" item-responsive>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.position.postName')" path="postName">
            <NInput v-model:value="model.postName" :placeholder="$t('page.hrm.position.form.postName')" />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.position.requirement')" path="requirementUuid">
            <NSelect
              v-model:value="model.requirementUuid"
              :options="requirementOptions"
              :placeholder="$t('page.hrm.position.form.requirement')"
              filterable
            />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.position.postCategory.title')" path="postCategory">
            <NSelect
              v-model:value="model.postCategory"
              :options="translateOptions(hrmPositionPostCategoryOptions)"
              :placeholder="$t('page.hrm.position.form.postCategory')"
            />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.position.ownerName')" path="ownerName">
            <NInput v-model:value="model.ownerName" :placeholder="$t('page.hrm.position.form.owner')" />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.position.headcount')" path="headcount">
            <NInputNumber v-model:value="model.headcount" :min="1" :precision="0" class="w-full" />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.position.publishDate')" path="publishDate">
            <NDatePicker
              v-model:formatted-value="model.publishDate"
              type="date"
              value-format="yyyy-MM-dd"
              class="w-full"
            />
          </NFormItemGi>
          <NFormItemGi span="24" :label="$t('page.hrm.position.remark')" path="remark">
            <NInput
              v-model:value="model.remark"
              type="textarea"
              :rows="2"
              :placeholder="$t('page.hrm.position.form.remark')"
            />
          </NFormItemGi>
        </NGrid>
      </NForm>
    </NScrollbar>
    <template #footer>
      <NSpace justify="end" :size="16">
        <NButton @click="closeModel">{{ $t('common.cancel') }}</NButton>
        <NButton type="primary" :loading="confirmLoading" @click="handleSubmit">{{ $t('common.save') }}</NButton>
      </NSpace>
    </template>
  </NModal>
</template>

<style scoped></style>
