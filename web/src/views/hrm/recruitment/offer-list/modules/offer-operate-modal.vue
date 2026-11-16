<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { AddOffer, EditOffer, GetCandidateList, GetPositionList } from '@/service/api';
import { $t } from '@/locales';
import { hrmOfferRankLevelOptions } from '@/constants/business';
import { translateOptions } from '@/utils/common';

defineOptions({
  name: 'OfferOperateModal'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Hrm.HrmOffer | null;
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
    add: $t('page.hrm.offer.addOffer'),
    edit: $t('page.hrm.offer.editOffer')
  };
  return titles[props.operateType];
});

type Model = Api.Hrm.AddOfferRequest;

const model: Model = reactive(createDefaultModel());

function createDefaultModel(): Model {
  return {
    candidateUuid: '',
    positionUuid: '',
    planEntryDate: '',
    probationMonths: 0,
    rankLevel: 'junior',
    reporterName: ''
  };
}

type RuleKey = Extract<keyof Model, 'candidateUuid' | 'positionUuid' | 'planEntryDate'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  candidateUuid: defaultRequiredRule,
  positionUuid: defaultRequiredRule,
  planEntryDate: defaultRequiredRule
};

// ---- 候选人 options（仅 status=offer 已发Offer 可发起录用, HRM-26 seam 契约） ----
interface PickerOption {
  label: string;
  value: string;
}

const candidateOptions = ref<PickerOption[]>([]);

async function loadCandidateOptions() {
  const { data, error } = await GetCandidateList({
    current: 1,
    size: 100,
    status: 'offer'
  });
  if (!error && data) {
    candidateOptions.value = (data.records ?? []).map(rec => ({
      label: rec.name,
      value: rec.uuid
    }));
  }
}

// ---- 职位 options（仅 status=open 招聘中 可录用, HRM-25 seam 契约） ----
const positionOptions = ref<PickerOption[]>([]);

async function loadPositionOptions() {
  const { data, error } = await GetPositionList({
    current: 1,
    size: 100,
    status: 'open'
  });
  if (!error && data) {
    positionOptions.value = (data.records ?? []).map(rec => ({
      label: rec.postName,
      value: rec.uuid
    }));
  }
}

function handleInitModel() {
  Object.assign(model, createDefaultModel());

  if (props.operateType === 'edit' && props.rowData) {
    Object.assign(model, {
      candidateUuid: props.rowData.candidateUuid,
      positionUuid: props.rowData.positionUuid,
      planEntryDate: props.rowData.planEntryDate,
      probationMonths: props.rowData.probationMonths,
      rankLevel: props.rowData.rankLevel,
      reporterName: props.rowData.reporterName
    });
  }
  loadCandidateOptions();
  loadPositionOptions();
}

function closeModel() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();
  confirmStartLoading();

  const request =
    props.operateType === 'add'
      ? AddOffer({ ...model })
      : EditOffer({ ...model, uuid: props.rowData?.uuid ?? '' });

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
      <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="110">
        <NGrid responsive="screen" item-responsive>
          <NFormItemGi span="24 s:12" :label="$t('page.hrm.offer.candidate')" path="candidateUuid">
            <NSelect
              v-model:value="model.candidateUuid"
              :options="candidateOptions"
              :placeholder="$t('page.hrm.offer.form.candidate')"
              filterable
            />
          </NFormItemGi>
          <NFormItemGi span="24 s:12" :label="$t('page.hrm.offer.position')" path="positionUuid">
            <NSelect
              v-model:value="model.positionUuid"
              :options="positionOptions"
              :placeholder="$t('page.hrm.offer.form.position')"
              filterable
            />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.offer.planEntryDate')" path="planEntryDate">
            <NDatePicker
              v-model:formatted-value="model.planEntryDate"
              type="date"
              value-format="yyyy-MM-dd"
              class="w-full"
            />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.offer.probationMonths')" path="probationMonths">
            <NInputNumber v-model:value="model.probationMonths" :min="0" :precision="0" class="w-full" />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.offer.rankLevel.title')" path="rankLevel">
            <NSelect
              v-model:value="model.rankLevel"
              :options="translateOptions(hrmOfferRankLevelOptions)"
              :placeholder="$t('page.hrm.offer.form.rankLevel')"
            />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.offer.reporterName')" path="reporterName">
            <NInput v-model:value="model.reporterName" :placeholder="$t('page.hrm.offer.form.reporter')" />
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
