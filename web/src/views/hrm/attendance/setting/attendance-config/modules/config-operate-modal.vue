<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { GetOrgTree, SaveAttendanceConfig } from '@/service/api';
import { $t } from '@/locales';

defineOptions({
  name: 'AttendanceConfigOperateModal'
});

interface Props {
  /** the edit row data (null = 新增/快速开始) */
  rowData?: Api.Hrm.AttendanceConfig | null;
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

const isEdit = computed(() => Boolean(props.rowData?.uuid));

const title = computed(() =>
  isEdit.value ? $t('page.hrm.attendance.config.editConfig') : $t('page.hrm.attendance.config.quickStart')
);

const fallbackHint = $t('page.hrm.attendance.config.fallbackHint');

type Model = {
  orgUuid: string;
  startTime: string;
  endTime: string;
  workDays: number[];
  remark: string;
};

const model: Model = reactive(createDefaultModel());

function createDefaultModel(): Model {
  return {
    orgUuid: '',
    startTime: '09:00',
    endTime: '18:00',
    workDays: [1, 2, 3, 4, 5],
    remark: ''
  };
}

const workDayOptions = computed(() =>
  Array.from({ length: 7 }, (_, i) => ({
    value: i + 1,
    label: $t(`page.hrm.attendance.group.weekday.${i + 1}` as App.I18n.I18nKey)
  }))
);

// org tree options
const orgTreeOptions = ref<{ label: string; value: string; children?: { label: string; value: string }[] }[]>([]);

async function loadOrgOptions() {
  const { error, data: tree } = await GetOrgTree();
  if (!error && tree) {
    const toOption = (
      node: Api.Hrm.Org
    ): { label: string; value: string; children?: { label: string; value: string }[] } => ({
      label: node.orgName,
      value: node.uuid,
      children: node.children?.length ? node.children.map(toOption) : undefined
    });
    orgTreeOptions.value = tree.records.map(toOption);
  }
}

type RuleKey = Extract<keyof Model, 'orgUuid'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  orgUuid: defaultRequiredRule
};

function handleInitModel() {
  Object.assign(model, createDefaultModel());

  if (props.rowData) {
    Object.assign(model, {
      orgUuid: props.rowData.orgUuid,
      startTime: props.rowData.startTime,
      endTime: props.rowData.endTime,
      workDays: props.rowData.workDays ? [...props.rowData.workDays] : [],
      remark: props.rowData.remark
    });
  }
}

function closeModel() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();
  confirmStartLoading();

  const { error } = await SaveAttendanceConfig({
    orgUuid: model.orgUuid,
    startTime: model.startTime,
    endTime: model.endTime,
    workDays: model.workDays,
    remark: model.remark
  });
  if (!error) {
    window.$message?.success($t('common.updateSuccess'));
    closeModel();
    emit('submitted');
  }
  confirmEndLoading();
}

watch(visible, () => {
  if (visible.value) {
    handleInitModel();
    loadOrgOptions();
    restoreValidation();
  }
});
</script>

<template>
  <NModal v-model:show="visible" :title="title" preset="card" class="w-600px">
    <NAlert type="info" class="mb-16px" :show-icon="true">{{ fallbackHint }}</NAlert>
    <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="110">
      <NFormItem :label="$t('page.hrm.attendance.config.org')" path="orgUuid">
        <NTreeSelect
          v-model:value="model.orgUuid"
          :options="orgTreeOptions"
          filterable
          clearable
          default-expand-all
          :disabled="isEdit"
          :placeholder="$t('page.hrm.attendance.config.form.org')"
        />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.attendance.group.startTime')" path="startTime">
        <NTimePicker
          :value="model.startTime as unknown as number"
          format="HH:mm"
          value-format="HH:mm"
          @update:value="(v: string | number | null) => (model.startTime = String(v ?? ''))"
        />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.attendance.group.endTime')" path="endTime">
        <NTimePicker
          :value="model.endTime as unknown as number"
          format="HH:mm"
          value-format="HH:mm"
          @update:value="(v: string | number | null) => (model.endTime = String(v ?? ''))"
        />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.attendance.group.workDays')" path="workDays">
        <NSelect
          v-model:value="model.workDays"
          multiple
          clearable
          :options="workDayOptions"
          :placeholder="$t('page.hrm.attendance.group.form.workDays')"
        />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.attendance.group.remark')" path="remark">
        <NInput
          v-model:value="model.remark"
          type="textarea"
          :placeholder="$t('page.hrm.attendance.group.form.remark')"
        />
      </NFormItem>
    </NForm>
    <template #footer>
      <NSpace :size="16" justify="end">
        <NButton @click="closeModel">{{ $t('common.cancel') }}</NButton>
        <NButton type="primary" :loading="confirmLoading" @click="handleSubmit">{{ $t('common.confirm') }}</NButton>
      </NSpace>
    </template>
  </NModal>
</template>

<style scoped></style>
