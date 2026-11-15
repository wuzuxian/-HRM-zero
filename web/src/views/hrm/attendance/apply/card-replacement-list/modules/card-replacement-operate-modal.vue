<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { AddCardReplacement, EditCardReplacement, GetEmployeeList } from '@/service/api';
import { $t } from '@/locales';

defineOptions({
  name: 'CardReplacementOperateModal'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Hrm.HrmCardReplacement | null;
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
    add: $t('page.hrm.attendance.cardReplacement.addTitle'),
    edit: $t('page.hrm.attendance.cardReplacement.editTitle')
  };
  return titles[props.operateType];
});

// ---- 员工 options（选择员工 → 带出档案快照） ----
interface EmployeeOption {
  label: string;
  value: string;
}

const employeeOptions = ref<EmployeeOption[]>([]);
const employeeMap = ref<Map<string, Api.Hrm.Employee>>(new Map());

async function loadEmployeeOptions() {
  if (employeeOptions.value.length > 0) return;
  const { data, error } = await GetEmployeeList({ current: 1, size: 200 });
  if (!error && data) {
    const options: EmployeeOption[] = [];
    const map = new Map<string, Api.Hrm.Employee>();
    for (const rec of data.records ?? []) {
      options.push({ label: `${rec.employeeNo} ${rec.name}`, value: rec.uuid });
      map.set(rec.uuid, rec);
    }
    employeeOptions.value = options;
    employeeMap.value = map;
  }
}

/** 选择员工 → 带出档案快照（工号/姓名/部门） */
function handleEmployeeChange(uuid: string) {
  const emp = employeeMap.value.get(uuid);
  if (!emp) return;
  model.employeeNo = emp.employeeNo;
  model.employeeName = emp.name;
  model.orgUuid = emp.orgUuid;
  model.orgName = (emp as unknown as { orgName?: string }).orgName ?? '';
}

// ---- 补卡类型 options ----
const cardTypeOptions = computed(() => [
  { label: $t('page.hrm.attendance.cardReplacement.clockIn'), value: 'clockIn' },
  { label: $t('page.hrm.attendance.cardReplacement.clockOut'), value: 'clockOut' }
]);

type Model = Api.Hrm.AddCardReplacementRequest;

const model: Model = reactive(createDefaultModel());

function createDefaultModel(): Model {
  return {
    employeeUuid: '',
    employeeNo: '',
    employeeName: '',
    orgUuid: '',
    orgName: '',
    punchDate: '',
    cardType: 'clockIn',
    punchTime: '',
    reason: ''
  };
}

/** 补卡日期（NDatePicker type="date" 单值） */
const punchDateTs = ref<number | null>(null);
/** 补卡时间（NDatePicker type="datetime" 单值） */
const punchTimeTs = ref<number | null>(null);

type RuleKey = Extract<keyof Model, 'employeeUuid' | 'cardType'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  employeeUuid: defaultRequiredRule,
  cardType: defaultRequiredRule
};

function handleInitModel() {
  Object.assign(model, createDefaultModel());
  punchDateTs.value = null;
  punchTimeTs.value = null;

  if (props.operateType === 'edit' && props.rowData) {
    Object.assign(model, props.rowData);
    const d = new Date(props.rowData.punchDate.replace(' ', 'T'));
    if (!Number.isNaN(d.getTime())) {
      punchDateTs.value = d.getTime();
    }
    const t = new Date(props.rowData.punchTime.replace(' ', 'T'));
    if (!Number.isNaN(t.getTime())) {
      punchTimeTs.value = t.getTime();
    }
  }
  loadEmployeeOptions();
}

function dateToBillDate(ts: number) {
  const d = new Date(ts);
  const pad = (n: number) => `${n}`.padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
}

function tsToBillTime(ts: number) {
  const d = new Date(ts);
  const pad = (n: number) => `${n}`.padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:00`;
}

function closeModel() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();
  if (!punchDateTs.value) {
    window.$message?.error($t('page.hrm.attendance.cardReplacement.form.punchDateRequired'));
    return;
  }
  if (!punchTimeTs.value) {
    window.$message?.error($t('page.hrm.attendance.cardReplacement.form.punchTimeRequired'));
    return;
  }
  confirmStartLoading();
  model.punchDate = dateToBillDate(punchDateTs.value);
  model.punchTime = tsToBillTime(punchTimeTs.value);

  const request =
    props.operateType === 'add'
      ? AddCardReplacement({ ...model })
      : EditCardReplacement({ ...model, uuid: props.rowData?.uuid ?? '' });

  const { error } = await request;
  confirmEndLoading();
  if (!error) {
    window.$message?.success(props.operateType === 'add' ? $t('common.addSuccess') : $t('common.updateSuccess'));
    closeModel();
    emit('submitted');
  }
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
    <NScrollbar class="max-h-480px pr-20px">
      <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="100">
        <NGrid responsive="screen" item-responsive>
          <NFormItemGi
            span="24 s:12"
            :label="$t('page.hrm.attendance.cardReplacement.selectEmployee')"
            path="employeeUuid"
          >
            <NSelect
              v-model:value="model.employeeUuid"
              :options="employeeOptions"
              filterable
              :disabled="operateType === 'edit'"
              :placeholder="$t('page.hrm.attendance.cardReplacement.form.selectEmployee')"
              @update:value="handleEmployeeChange"
            />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:8"
            :label="$t('page.hrm.attendance.cardReplacement.employeeNo')"
            path="employeeNo"
          >
            <NInput v-model:value="model.employeeNo" disabled />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:8"
            :label="$t('page.hrm.attendance.cardReplacement.employeeName')"
            path="employeeName"
          >
            <NInput v-model:value="model.employeeName" disabled />
          </NFormItemGi>
        </NGrid>
        <NGrid responsive="screen" item-responsive>
          <NFormItemGi
            span="24 s:12 m:8"
            :label="$t('page.hrm.attendance.cardReplacement.punchDate')"
            path="punchDate"
          >
            <NDatePicker
              v-model:value="punchDateTs"
              type="date"
              clearable
              :placeholder="$t('page.hrm.attendance.cardReplacement.form.punchDate')"
            />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:8"
            :label="$t('page.hrm.attendance.cardReplacement.cardType')"
            path="cardType"
          >
            <NSelect
              v-model:value="model.cardType"
              :options="cardTypeOptions"
              :placeholder="$t('page.hrm.attendance.cardReplacement.form.cardType')"
            />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:8"
            :label="$t('page.hrm.attendance.cardReplacement.punchTime')"
            path="punchTime"
          >
            <NDatePicker
              v-model:value="punchTimeTs"
              type="datetime"
              clearable
              :placeholder="$t('page.hrm.attendance.cardReplacement.form.punchTime')"
            />
          </NFormItemGi>
        </NGrid>
        <NFormItem :label="$t('page.hrm.attendance.cardReplacement.reason')" path="reason">
          <NInput
            v-model:value="model.reason"
            type="textarea"
            :rows="2"
            :placeholder="$t('page.hrm.attendance.cardReplacement.form.reason')"
          />
        </NFormItem>
      </NForm>
    </NScrollbar>
    <template #footer>
      <NSpace justify="end">
        <NButton @click="closeModel">{{ $t('common.cancel') }}</NButton>
        <NButton type="primary" :loading="confirmLoading" @click="handleSubmit">{{ $t('common.confirm') }}</NButton>
      </NSpace>
    </template>
  </NModal>
</template>

<style scoped></style>
