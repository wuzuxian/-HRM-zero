<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { AddTrainingSession, EditTrainingSession, GetTrainingPlanList } from '@/service/api';
import { $t } from '@/locales';

defineOptions({
  name: 'SessionOperateModal'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Hrm.TrainingSession | null;
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
    add: $t('page.hrm.training.session.addSession'),
    edit: $t('page.hrm.training.session.editSession')
  };
  return titles[props.operateType];
});

type Model = Omit<Api.Hrm.AddTrainingSessionRequest, 'startTime' | 'endTime'> & {
  // NDatePicker 的 formatted-value 不能是空串（解析 "" 抛 Invalid time value），未选择时必须为 null
  startTime: string | null;
  endTime: string | null;
};

const model: Model = reactive(createDefaultModel());

function createDefaultModel(): Model {
  return {
    sessionName: '',
    planUuid: '',
    lecturer: '',
    location: '',
    hours: 0,
    startTime: null,
    endTime: null,
    remark: ''
  };
}

type RuleKey = Extract<keyof Model, 'sessionName' | 'planUuid' | 'startTime' | 'endTime'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  sessionName: defaultRequiredRule,
  planUuid: defaultRequiredRule,
  startTime: defaultRequiredRule,
  endTime: defaultRequiredRule
};

// ---- published plan options (发布计划后可建班次) ----
interface PlanOption {
  label: string;
  value: string;
}

const planOptions = ref<PlanOption[]>([]);
const planOptionsLoaded = ref(false);

async function loadPlanOptions() {
  if (planOptionsLoaded.value) return;
  const { data, error } = await GetTrainingPlanList({ current: 1, size: 100, published: '1' });
  if (!error && data) {
    planOptions.value = (data.records ?? []).map(p => ({
      label: `${p.planName}（${p.billNo}）`,
      value: p.uuid
    }));
    planOptionsLoaded.value = true;
  }
}

function handleInitModel() {
  Object.assign(model, createDefaultModel());

  if (props.operateType === 'edit' && props.rowData) {
    Object.assign(model, {
      sessionName: props.rowData.sessionName,
      planUuid: props.rowData.planUuid,
      lecturer: props.rowData.lecturer,
      location: props.rowData.location,
      hours: props.rowData.hours,
      // DB 列 NOT NULL DEFAULT ''，编辑回填可能拿到空串；空串会让 NDatePicker 抛 Invalid time value
      startTime: props.rowData.startTime || null,
      endTime: props.rowData.endTime || null,
      remark: props.rowData.remark
    });
  }
  loadPlanOptions();
}

function closeModel() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();
  // validate() 已拦截 null；此比较仅对非空时间串有意义，null 参与比较会得到 false 不触发拦截
  if (model.startTime && model.endTime && model.endTime <= model.startTime) {
    window.$message?.error($t('page.hrm.training.session.form.endTimeBeforeStart'));
    return;
  }
  confirmStartLoading();

  const request =
    props.operateType === 'add'
      ? AddTrainingSession({ ...model, startTime: model.startTime ?? '', endTime: model.endTime ?? '' })
      : EditTrainingSession({
          ...model,
          startTime: model.startTime ?? '',
          endTime: model.endTime ?? '',
          uuid: props.rowData?.uuid ?? ''
        });

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
  <NModal v-model:show="visible" :title="title" preset="card" class="w-680px">
    <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="100">
      <NGrid responsive="screen" item-responsive>
        <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.training.session.sessionName')" path="sessionName">
          <NInput v-model:value="model.sessionName" :placeholder="$t('page.hrm.training.session.form.sessionName')" />
        </NFormItemGi>
        <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.training.session.planName')" path="planUuid">
          <NSelect
            v-model:value="model.planUuid"
            :options="planOptions"
            :placeholder="$t('page.hrm.training.session.form.planUuid')"
            filterable
          />
        </NFormItemGi>
        <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.training.lecturer')" path="lecturer">
          <NInput v-model:value="model.lecturer" :placeholder="$t('page.hrm.training.session.form.lecturer')" />
        </NFormItemGi>
        <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.training.location')" path="location">
          <NInput v-model:value="model.location" :placeholder="$t('page.hrm.training.session.form.location')" />
        </NFormItemGi>
        <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.training.hours')" path="hours">
          <NInputNumber v-model:value="model.hours" :min="0" :precision="1" class="w-full" />
        </NFormItemGi>
        <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.training.startTime')" path="startTime">
          <NDatePicker
            v-model:formatted-value="model.startTime"
            type="datetime"
            value-format="yyyy-MM-dd HH:mm:ss"
            class="w-full"
          />
        </NFormItemGi>
        <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.training.endTime')" path="endTime">
          <NDatePicker
            v-model:formatted-value="model.endTime"
            type="datetime"
            value-format="yyyy-MM-dd HH:mm:ss"
            class="w-full"
          />
        </NFormItemGi>
        <NFormItemGi span="24" :label="$t('page.hrm.training.session.remark')" path="remark">
          <NInput v-model:value="model.remark" type="textarea" :rows="2" />
        </NFormItemGi>
      </NGrid>
    </NForm>
    <template #footer>
      <NSpace justify="end" :size="16">
        <NButton @click="closeModel">{{ $t('common.cancel') }}</NButton>
        <NButton type="primary" :loading="confirmLoading" @click="handleSubmit">{{ $t('common.save') }}</NButton>
      </NSpace>
    </template>
  </NModal>
</template>

<style scoped></style>
