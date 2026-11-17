<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { AddAttendanceGroup, EditAttendanceGroup, GetEmployeeList, GetOrgTree } from '@/service/api';
import { $t } from '@/locales';
import { hrmAttendanceStatusOptions } from '@/constants/business';

defineOptions({
  name: 'AttendanceGroupOperateModal'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Hrm.AttendanceGroup | null;
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
    add: $t('page.hrm.attendance.group.addGroup'),
    edit: $t('page.hrm.attendance.group.editGroup')
  };
  return titles[props.operateType];
});

/** 班制 options */
const groupTypeOptions: CommonType.Option<string>[] = [
  { value: '1', label: 'page.hrm.attendance.group.groupType.fixed' },
  { value: '2', label: 'page.hrm.attendance.group.groupType.scheduling' },
  { value: '3', label: 'page.hrm.attendance.group.groupType.flexible' }
];

/** 考勤日 options (周一至周日) */
const workDayOptions = computed(() =>
  Array.from({ length: 7 }, (_, i) => ({
    value: i + 1,
    label: $t(`page.hrm.attendance.group.weekday.${i + 1}` as App.I18n.I18nKey)
  }))
);

type Model = {
  groupName: string;
  groupType: Api.Hrm.AttendanceGroupType | '';
  workDays: number[];
  startTime: string;
  endTime: string;
  flexibleMinutes: number;
  orgUuids: string[];
  memberUuids: string[];
  managers: string[];
  status: Api.Hrm.AttendanceStatus;
  remark: string;
};

const model: Model = reactive(createDefaultModel());

function createDefaultModel(): Model {
  return {
    groupName: '',
    groupType: '1',
    workDays: [1, 2, 3, 4, 5],
    startTime: '09:00',
    endTime: '18:00',
    flexibleMinutes: 0,
    orgUuids: [],
    memberUuids: [],
    managers: [],
    status: '1',
    remark: ''
  };
}

type RuleKey = Extract<keyof Model, 'groupName' | 'groupType'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  groupName: defaultRequiredRule,
  groupType: defaultRequiredRule
};

/** 固定班制必须携带考勤时间 (与后端校验口径一致) */
const timeRequired = computed(() => model.groupType === '1');

// ---- 适用范围 options: 部门(树拍平) + 指定人员 ----
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

const employeeOptions = ref<{ label: string; value: string }[]>([]);

async function loadEmployeeOptions() {
  const { error, data } = await GetEmployeeList({ current: 1, size: 500 });
  if (!error && data) {
    employeeOptions.value = data.records.map(row => ({
      label: `${row.name}（${row.employeeNo}）`,
      value: row.uuid
    }));
  }
}

watch(visible, v => {
  if (v) {
    loadOrgOptions();
    loadEmployeeOptions();
  }
});

function handleInitModel() {
  Object.assign(model, createDefaultModel());

  if (props.operateType === 'edit' && props.rowData) {
    Object.assign(model, {
      ...props.rowData,
      workDays: props.rowData.workDays ? [...props.rowData.workDays] : [],
      orgUuids: props.rowData.orgUuids ? [...props.rowData.orgUuids] : [],
      memberUuids: props.rowData.memberUuids ? [...props.rowData.memberUuids] : [],
      managers: props.rowData.managers ? [...props.rowData.managers] : [],
      status: props.rowData.status ?? '1'
    });
  }
}

function closeModel() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();
  confirmStartLoading();

  const payload: Api.Hrm.AddAttendanceGroupRequest = {
    groupName: model.groupName,
    groupType: model.groupType,
    workDays: model.workDays,
    startTime: timeRequired.value ? model.startTime : model.startTime || undefined,
    endTime: timeRequired.value ? model.endTime : model.endTime || undefined,
    flexibleMinutes: model.flexibleMinutes,
    orgUuids: model.orgUuids,
    memberUuids: model.memberUuids,
    managers: model.managers,
    status: model.status,
    remark: model.remark
  };

  const request =
    props.operateType === 'add'
      ? AddAttendanceGroup(payload)
      : EditAttendanceGroup({ ...payload, uuid: props.rowData?.uuid ?? '', status: model.status });

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
  <NModal v-model:show="visible" :title="title" preset="card" class="w-700px">
    <NScrollbar class="h-480px pr-20px">
      <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="110">
        <NFormItem :label="$t('page.hrm.attendance.group.groupName')" path="groupName">
          <NInput v-model:value="model.groupName" :placeholder="$t('page.hrm.attendance.group.form.groupName')" />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.group.groupType.title')" path="groupType">
          <NRadioGroup v-model:value="model.groupType">
            <NRadio v-for="item in groupTypeOptions" :key="item.value" :value="item.value" :label="$t(item.label)" />
          </NRadioGroup>
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
        <NFormItem
          :label="$t('page.hrm.attendance.group.startTime')"
          path="startTime"
          :rule="timeRequired ? defaultRequiredRule : undefined"
        >
          <NTimePicker
            :formatted-value="model.startTime || null"
            format="HH:mm"
            value-format="HH:mm"
            :disabled="model.groupType === '3'"
            @update:formatted-value="(v: string | null) => (model.startTime = v ?? '')"
          />
        </NFormItem>
        <NFormItem
          :label="$t('page.hrm.attendance.group.endTime')"
          path="endTime"
          :rule="timeRequired ? defaultRequiredRule : undefined"
        >
          <NTimePicker
            :formatted-value="model.endTime || null"
            format="HH:mm"
            value-format="HH:mm"
            :disabled="model.groupType === '3'"
            @update:formatted-value="(v: string | null) => (model.endTime = v ?? '')"
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.group.flexibleMinutes')" path="flexibleMinutes">
          <NInputNumber v-model:value="model.flexibleMinutes" :min="0" class="w-full" />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.group.orgScope')">
          <NSelect
            v-model:value="model.orgUuids"
            multiple
            filterable
            clearable
            :options="orgTreeOptions"
            :placeholder="$t('page.hrm.attendance.group.form.orgScope')"
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.group.memberScope')">
          <NSelect
            v-model:value="model.memberUuids"
            multiple
            filterable
            clearable
            :options="employeeOptions"
            :placeholder="$t('page.hrm.attendance.group.form.memberScope')"
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.group.managers')">
          <NSelect
            v-model:value="model.managers"
            tag
            multiple
            filterable
            :options="[]"
            :placeholder="$t('page.hrm.attendance.group.form.managers')"
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.group.status')" path="status">
          <NRadioGroup v-model:value="model.status">
            <NRadio
              v-for="item in hrmAttendanceStatusOptions"
              :key="item.value"
              :value="item.value"
              :label="$t(item.label)"
            />
          </NRadioGroup>
        </NFormItem>
        <NFormItem :label="$t('page.hrm.attendance.group.remark')" path="remark">
          <NInput
            v-model:value="model.remark"
            type="textarea"
            :placeholder="$t('page.hrm.attendance.group.form.remark')"
          />
        </NFormItem>
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
