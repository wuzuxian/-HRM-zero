<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { useAuth } from '@/hooks/business/auth';
import {
  AddEmployee,
  CreateAndLinkEmployeeUser,
  EditEmployee,
  GetEmployeeDetail,
  GetLinkableUserList,
  LinkEmployeeUser,
  UnlinkEmployeeUser
} from '@/service/api';
import { $t } from '@/locales';
import {
  hrmEmployeeGenderOptions,
  hrmEmployeeMaritalStatusOptions,
  hrmEmployeePersonStatusOptions
} from '@/constants/business';
import { translateOptions } from '@/utils/common';

defineOptions({
  name: 'EmployeeOperateModal'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data (detail record incl. sub-tables) */
  rowData?: Api.Hrm.Employee | null;
  /** flat org options for 部门/单位选择 */
  orgOptions?: Api.Hrm.Org[];
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
const { hasAuth } = useAuth();

const activeTab = ref<'basic' | 'education' | 'work' | 'family'>('basic');

const title = computed(() => {
  const titles: Record<NaiveUI.TableOperateType, string> = {
    add: $t('page.hrm.employee.addEmployee'),
    edit: $t('page.hrm.employee.editEmployee')
  };
  return titles[props.operateType];
});

const orgSelectOptions = computed(() =>
  (props.orgOptions ?? []).map(item => ({ label: item.orgName, value: item.uuid }))
);

type Model = Omit<Api.Hrm.AddEmployeeRequest, 'educations' | 'works' | 'families'> & {
  educations: NonNullable<Api.Hrm.AddEmployeeRequest['educations']>;
  works: NonNullable<Api.Hrm.AddEmployeeRequest['works']>;
  families: NonNullable<Api.Hrm.AddEmployeeRequest['families']>;
};

const model: Model = reactive(createDefaultModel());

function createDefaultModel(): Model {
  return {
    employeeNo: '',
    name: '',
    gender: '1',
    birthday: '',
    nation: '',
    politicalStatus: '',
    maritalStatus: '',
    nativePlace: '',
    bloodType: '',
    idCard: '',
    height: '',
    weight: '',
    householdAddress: '',
    address: '',
    emergencyContact: '',
    emergencyPhone: '',
    phone: '',
    email: '',
    avatar: '',
    bankName: '',
    bankAccount: '',
    companyUuid: '',
    orgUuid: '',
    position: '',
    post: '',
    personStatus: '1',
    hireDate: '',
    formalDate: '',
    remark: '',
    educations: [],
    works: [],
    families: []
  };
}

type RuleKey = Extract<keyof Model, 'name'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  name: defaultRequiredRule
};

function handleInitModel() {
  Object.assign(model, createDefaultModel());
  activeTab.value = 'basic';

  if (props.operateType === 'edit' && props.rowData) {
    Object.assign(model, {
      ...props.rowData,
      educations: (props.rowData.educations ?? []).map((item: Api.Hrm.EmployeeEducation) => ({ ...item })),
      works: (props.rowData.works ?? []).map((item: Api.Hrm.EmployeeWork) => ({ ...item })),
      families: (props.rowData.families ?? []).map((item: Api.Hrm.EmployeeFamily) => ({ ...item }))
    });
    // 系统账号区块: 关联状态取详情行 (userUuid 空串=未关联, ADR 0007)
    linkedUserUuid.value = props.rowData.userUuid ?? '';
    linkedUsername.value = props.rowData.loginName ?? '';
    handleCandidateSearch('');
  } else {
    linkedUserUuid.value = '';
    linkedUsername.value = '';
    candidateOptions.value = [];
    selectedUserUuid.value = null;
  }
}

// ---- 系统账号区块 (HRM-51): 关联/解绑/新建并关联; 关联唯一入口在员工侧 (ADR 0007) ----
const linkedUserUuid = ref('');
const linkedUsername = ref('');

const candidateOptions = ref<{ label: string; value: string }[]>([]);
const candidateLoading = ref(false);
const selectedUserUuid = ref<string | null>(null);

async function handleCandidateSearch(query: string) {
  candidateLoading.value = true;
  const { error, data } = await GetLinkableUserList({ current: 1, size: 20, keyword: query || null });
  candidateLoading.value = false;
  if (error || !data) {
    return;
  }
  // 候选只含未关联账号 (后端排除已被占用的); 已关联的当前账号无需排除 (不在候选里)
  candidateOptions.value = (data.records ?? []).map(user => ({
    label: `${user.username} (${user.nickname})`,
    value: user.uuid
  }));
}

async function refreshAccountState() {
  const { error, data: detail } = await GetEmployeeDetail(props.rowData?.uuid ?? '');
  if (!error && detail) {
    linkedUserUuid.value = detail.employee.userUuid;
    linkedUsername.value = detail.employee.loginName;
  }
  emit('submitted');
}

const linkLoading = ref(false);

async function handleLink() {
  if (!selectedUserUuid.value) {
    window.$message?.warning($t('page.hrm.employee.systemAccountSelectFirst'));
    return;
  }
  linkLoading.value = true;
  const { error } = await LinkEmployeeUser({
    employeeUuid: props.rowData?.uuid ?? '',
    userUuid: selectedUserUuid.value
  });
  linkLoading.value = false;
  if (error) return;
  window.$message?.success($t('page.hrm.employee.systemAccountLinkSuccess'));
  selectedUserUuid.value = null;
  candidateOptions.value = [];
  await refreshAccountState();
}

const unlinkLoading = ref(false);

async function handleUnlink() {
  unlinkLoading.value = true;
  const { error } = await UnlinkEmployeeUser({ employeeUuid: props.rowData?.uuid ?? '' });
  unlinkLoading.value = false;
  if (error) return;
  window.$message?.success($t('page.hrm.employee.systemAccountUnlinkSuccess'));
  linkedUsername.value = '';
  await refreshAccountState();
}

const createForm = reactive({ username: '', password: '' });
const createLoading = ref(false);

async function handleCreateAndLink() {
  createLoading.value = true;
  const { error, data } = await CreateAndLinkEmployeeUser({
    employeeUuid: props.rowData?.uuid ?? '',
    username: createForm.username || undefined,
    password: createForm.password || undefined
  });
  createLoading.value = false;
  if (error) return;
  // 成功弹窗明示凭据 (与新增员工建档即开号同一交互, 文案共用 accountCreatedTip)
  window.$dialog?.success({
    title: $t('page.hrm.employee.systemAccountCreateTitle'),
    content: $t('page.hrm.employee.accountCreatedTip', {
      username: data?.username ?? '',
      password: data?.initialPassword ?? ''
    }),
    positiveText: $t('common.confirm')
  });
  createForm.username = '';
  createForm.password = '';
  await refreshAccountState();
}

function closeModel() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();
  confirmStartLoading();

  if (props.operateType === 'add') {
    const { error, data: added } = await AddEmployee({ ...model });
    if (error) {
      confirmEndLoading();
      return;
    }
    closeModel();
    emit('submitted');
    // 建档即开号 (spec #50): 成功提示明示登录凭据; 离职补录 (userCreated=false) 不产账号
    if (added?.userCreated && added.username) {
      window.$dialog?.success({
        title: $t('page.hrm.employee.accountCreatedTitle'),
        content: $t('page.hrm.employee.accountCreatedTip', {
          username: added.username,
          password: added.initialPassword ?? ''
        }),
        positiveText: $t('common.confirm')
      });
    } else {
      window.$message?.success($t('page.hrm.employee.addEmployeeNoAccount'));
    }
    confirmEndLoading();
    return;
  }

  const { error } = await EditEmployee({ ...model, uuid: props.rowData?.uuid ?? '' });
  if (!error) {
    window.$message?.success($t('common.updateSuccess'));
    closeModel();
    emit('submitted');
  }
  confirmEndLoading();
}

// ---- 从表行编辑 ----
function addEducation() {
  model.educations.push({ school: '', major: '', education: '', startDate: '', endDate: '', remark: '' });
}
function removeEducation(index: number) {
  model.educations.splice(index, 1);
}
function addWork() {
  model.works.push({ company: '', position: '', startDate: '', endDate: '', remark: '' });
}
function removeWork(index: number) {
  model.works.splice(index, 1);
}
function addFamily() {
  model.families.push({ name: '', relation: '', phone: '', workUnit: '', remark: '' });
}
function removeFamily(index: number) {
  model.families.splice(index, 1);
}

watch(visible, () => {
  if (visible.value) {
    handleInitModel();
    restoreValidation();
  }
});

// ---- date pickers bind string (yyyy-MM-dd) model fields ----
function toDateNumber(value: string): number | null {
  if (!value) return null;
  const ts = Date.parse(`${value}T00:00:00`);
  return Number.isNaN(ts) ? null : ts;
}

function fromDateNumber(value: number | null): string {
  if (!value) return '';
  const d = new Date(value);
  const mm = String(d.getMonth() + 1).padStart(2, '0');
  const dd = String(d.getDate()).padStart(2, '0');
  return `${d.getFullYear()}-${mm}-${dd}`;
}

const hireDateValue = computed(() => toDateNumber(model.hireDate ?? ''));
function onHireDateChange(v: number | null) {
  model.hireDate = fromDateNumber(v);
}

const formalDateValue = computed(() => toDateNumber(model.formalDate ?? ''));
function onFormalDateChange(v: number | null) {
  model.formalDate = fromDateNumber(v);
}
</script>

<template>
  <NModal v-model:show="visible" :title="title" preset="card" class="w-860px">
    <NScrollbar class="h-480px pr-20px">
      <NTabs v-model:value="activeTab" type="line">
        <NTabPane name="basic" :tab="$t('page.hrm.employee.tabBasic')">
          <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="110">
            <NGrid responsive="screen" item-responsive>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employee.employeeNo')" path="employeeNo">
                <NInput
                  v-model:value="model.employeeNo"
                  :placeholder="$t('page.hrm.employee.form.employeeNoAuto')"
                  :disabled="operateType === 'edit'"
                  clearable
                />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employee.name')" path="name">
                <NInput v-model:value="model.name" :placeholder="$t('page.hrm.employee.form.name')" />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employee.genderLabel')" path="gender">
                <NRadioGroup v-model:value="model.gender">
                  <NRadio
                    v-for="item in hrmEmployeeGenderOptions"
                    :key="item.value"
                    :value="item.value"
                    :label="$t(item.label)"
                  />
                </NRadioGroup>
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employee.personStatusLabel')" path="personStatus">
                <NSelect
                  v-model:value="model.personStatus"
                  :options="translateOptions(hrmEmployeePersonStatusOptions)"
                  :placeholder="$t('page.hrm.employee.form.personStatus')"
                />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employee.hireDate')" path="hireDate">
                <NDatePicker
                  v-model:value="hireDateValue"
                  type="date"
                  class="w-full"
                  @update:value="onHireDateChange"
                />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employee.formalDate')" path="formalDate">
                <NDatePicker
                  v-model:value="formalDateValue"
                  type="date"
                  class="w-full"
                  @update:value="onFormalDateChange"
                />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employee.orgName')" path="orgUuid">
                <NSelect
                  v-model:value="model.orgUuid"
                  :options="orgSelectOptions"
                  :placeholder="$t('page.hrm.employee.form.org')"
                  clearable
                  filterable
                />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employee.companyName')" path="companyUuid">
                <NSelect
                  v-model:value="model.companyUuid"
                  :options="orgSelectOptions"
                  :placeholder="$t('page.hrm.employee.form.company')"
                  clearable
                  filterable
                />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employee.position')" path="position">
                <NInput v-model:value="model.position" :placeholder="$t('page.hrm.employee.form.position')" />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employee.post')" path="post">
                <NInput v-model:value="model.post" :placeholder="$t('page.hrm.employee.form.post')" />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employee.phone')" path="phone">
                <NInput v-model:value="model.phone" :placeholder="$t('page.hrm.employee.form.phone')" />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employee.email')" path="email">
                <NInput v-model:value="model.email" :placeholder="$t('page.hrm.employee.form.email')" />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employee.birthday')" path="birthday">
                <NInput v-model:value="model.birthday" :placeholder="$t('page.hrm.employee.form.birthday')" />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employee.nation')" path="nation">
                <NInput v-model:value="model.nation" :placeholder="$t('page.hrm.employee.form.nation')" />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employee.politicalStatus')" path="politicalStatus">
                <NInput
                  v-model:value="model.politicalStatus"
                  :placeholder="$t('page.hrm.employee.form.politicalStatus')"
                />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employee.maritalStatusLabel')" path="maritalStatus">
                <NSelect
                  v-model:value="model.maritalStatus"
                  :options="translateOptions(hrmEmployeeMaritalStatusOptions)"
                  :placeholder="$t('page.hrm.employee.form.maritalStatus')"
                  clearable
                />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employee.nativePlace')" path="nativePlace">
                <NInput v-model:value="model.nativePlace" :placeholder="$t('page.hrm.employee.form.nativePlace')" />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employee.bloodType')" path="bloodType">
                <NInput v-model:value="model.bloodType" :placeholder="$t('page.hrm.employee.form.bloodType')" />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employee.idCard')" path="idCard">
                <NInput v-model:value="model.idCard" :placeholder="$t('page.hrm.employee.form.idCard')" />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employee.householdAddress')" path="householdAddress">
                <NInput
                  v-model:value="model.householdAddress"
                  :placeholder="$t('page.hrm.employee.form.householdAddress')"
                />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employee.address')" path="address">
                <NInput v-model:value="model.address" :placeholder="$t('page.hrm.employee.form.address')" />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employee.emergencyContact')" path="emergencyContact">
                <NInput
                  v-model:value="model.emergencyContact"
                  :placeholder="$t('page.hrm.employee.form.emergencyContact')"
                />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employee.emergencyPhone')" path="emergencyPhone">
                <NInput
                  v-model:value="model.emergencyPhone"
                  :placeholder="$t('page.hrm.employee.form.emergencyPhone')"
                />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employee.bankName')" path="bankName">
                <NInput v-model:value="model.bankName" :placeholder="$t('page.hrm.employee.form.bankName')" />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employee.bankAccount')" path="bankAccount">
                <NInput v-model:value="model.bankAccount" :placeholder="$t('page.hrm.employee.form.bankAccount')" />
              </NFormItemGi>
              <NFormItemGi span="24 m:16" :label="$t('page.hrm.employee.remark')" path="remark">
                <NInput
                  v-model:value="model.remark"
                  type="textarea"
                  :placeholder="$t('page.hrm.employee.form.remark')"
                />
              </NFormItemGi>
            </NGrid>
          </NForm>
        </NTabPane>
        <NTabPane name="account" :tab="$t('page.hrm.employee.systemAccountTab')">
          <NSpace v-if="operateType === 'edit'" vertical :size="16">
            <NTag :type="linkedUserUuid ? 'success' : 'warning'" size="small">
              {{
                linkedUserUuid
                  ? `${$t('page.hrm.employee.systemAccountLinked')}: ${linkedUsername}`
                  : $t('page.hrm.employee.systemAccountUnlinked')
              }}
            </NTag>
            <NCard
              v-if="hasAuth('hrm:employee:linkUser')"
              size="small"
              :title="$t('page.hrm.employee.systemAccountLinkButton')"
            >
              <NSelect
                v-model:value="selectedUserUuid"
                remote
                filterable
                clearable
                :options="candidateOptions"
                :loading="candidateLoading"
                :placeholder="$t('page.hrm.employee.systemAccountSearchPlaceholder')"
                @search="handleCandidateSearch"
              />
              <NButton class="mt-12px" type="primary" ghost :loading="linkLoading" @click="handleLink">
                {{ $t('page.hrm.employee.systemAccountLinkButton') }}
              </NButton>
            </NCard>
            <NCard
              size="small"
              :title="$t('page.hrm.employee.systemAccountCreateTitle')"
            >
              <template #header-extra>
                <NPopconfirm v-if="linkedUserUuid && hasAuth('hrm:employee:unlinkUser')" @positive-click="handleUnlink">
                  <template #trigger>
                    <NButton type="error" ghost size="small" :loading="unlinkLoading">
                      {{ $t('page.hrm.employee.systemAccountUnlinkButton') }}
                    </NButton>
                  </template>
                  {{ $t('page.hrm.employee.systemAccountUnlinkConfirm') }}
                </NPopconfirm>
              </template>
              <NSpace v-if="hasAuth('hrm:employee:createAndLinkUser')" vertical>
                <NInput
                  v-model:value="createForm.username"
                  :placeholder="$t('page.hrm.employee.systemAccountCreateUsername')"
                />
                <NInput
                  v-model:value="createForm.password"
                  :placeholder="$t('page.hrm.employee.systemAccountCreatePassword')"
                />
                <NButton type="primary" ghost :loading="createLoading" @click="handleCreateAndLink">
                  {{ $t('page.hrm.employee.systemAccountCreateButton') }}
                </NButton>
              </NSpace>
            </NCard>
          </NSpace>
          <NEmpty v-else :description="$t('page.hrm.employee.systemAccountAddModeTip')" />
        </NTabPane>
        <NTabPane name="education" :tab="$t('page.hrm.employee.tabEducation')">
          <NSpace vertical :size="12">
            <NButton size="small" type="primary" ghost @click="addEducation">
              {{ $t('page.hrm.employee.addEducationRow') }}
            </NButton>
            <NEmpty v-if="model.educations.length === 0" :description="$t('page.hrm.employee.noEducationRows')" />
            <NCard v-for="(item, index) in model.educations" :key="index" size="small">
              <NGrid responsive="screen" item-responsive>
                <NFormItemGi span="24 s:12 m:6" :label="$t('page.hrm.employee.eduSchool')">
                  <NInput v-model:value="item.school" />
                </NFormItemGi>
                <NFormItemGi span="24 s:12 m:6" :label="$t('page.hrm.employee.eduMajor')">
                  <NInput v-model:value="item.major" />
                </NFormItemGi>
                <NFormItemGi span="24 s:12 m:6" :label="$t('page.hrm.employee.eduEducation')">
                  <NInput v-model:value="item.education" />
                </NFormItemGi>
                <NFormItemGi span="24 s:12 m:6" :label="$t('page.hrm.employee.eduStartDate')">
                  <NInput v-model:value="item.startDate" placeholder="2012-09-01" />
                </NFormItemGi>
                <NFormItemGi span="24 s:12 m:6" :label="$t('page.hrm.employee.eduEndDate')">
                  <NInput v-model:value="item.endDate" placeholder="2016-06-30" />
                </NFormItemGi>
                <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employee.remark')">
                  <NInput v-model:value="item.remark" />
                </NFormItemGi>
                <NFormItemGi span="24 s:12 m:4">
                  <NButton size="small" type="error" ghost @click="removeEducation(index)">
                    {{ $t('common.delete') }}
                  </NButton>
                </NFormItemGi>
              </NGrid>
            </NCard>
          </NSpace>
        </NTabPane>
        <NTabPane name="work" :tab="$t('page.hrm.employee.tabWork')">
          <NSpace vertical :size="12">
            <NButton size="small" type="primary" ghost @click="addWork">
              {{ $t('page.hrm.employee.addWorkRow') }}
            </NButton>
            <NEmpty v-if="model.works.length === 0" :description="$t('page.hrm.employee.noWorkRows')" />
            <NCard v-for="(item, index) in model.works" :key="index" size="small">
              <NGrid responsive="screen" item-responsive>
                <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employee.workCompany')">
                  <NInput v-model:value="item.company" />
                </NFormItemGi>
                <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employee.workPosition')">
                  <NInput v-model:value="item.position" />
                </NFormItemGi>
                <NFormItemGi span="24 s:12 m:6" :label="$t('page.hrm.employee.workStartDate')">
                  <NInput v-model:value="item.startDate" placeholder="2018-07-01" />
                </NFormItemGi>
                <NFormItemGi span="24 s:12 m:6" :label="$t('page.hrm.employee.workEndDate')">
                  <NInput v-model:value="item.endDate" placeholder="2020-12-31" />
                </NFormItemGi>
                <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employee.remark')">
                  <NInput v-model:value="item.remark" />
                </NFormItemGi>
                <NFormItemGi span="24 s:12 m:4">
                  <NButton size="small" type="error" ghost @click="removeWork(index)">
                    {{ $t('common.delete') }}
                  </NButton>
                </NFormItemGi>
              </NGrid>
            </NCard>
          </NSpace>
        </NTabPane>
        <NTabPane name="family" :tab="$t('page.hrm.employee.tabFamily')">
          <NSpace vertical :size="12">
            <NButton size="small" type="primary" ghost @click="addFamily">
              {{ $t('page.hrm.employee.addFamilyRow') }}
            </NButton>
            <NEmpty v-if="model.families.length === 0" :description="$t('page.hrm.employee.noFamilyRows')" />
            <NCard v-for="(item, index) in model.families" :key="index" size="small">
              <NGrid responsive="screen" item-responsive>
                <NFormItemGi span="24 s:12 m:6" :label="$t('page.hrm.employee.familyName')">
                  <NInput v-model:value="item.name" />
                </NFormItemGi>
                <NFormItemGi span="24 s:12 m:6" :label="$t('page.hrm.employee.familyRelation')">
                  <NInput v-model:value="item.relation" />
                </NFormItemGi>
                <NFormItemGi span="24 s:12 m:6" :label="$t('page.hrm.employee.familyPhone')">
                  <NInput v-model:value="item.phone" />
                </NFormItemGi>
                <NFormItemGi span="24 s:12 m:6" :label="$t('page.hrm.employee.familyWorkUnit')">
                  <NInput v-model:value="item.workUnit" />
                </NFormItemGi>
                <NFormItemGi span="24 m:8" :label="$t('page.hrm.employee.remark')">
                  <NInput v-model:value="item.remark" />
                </NFormItemGi>
                <NFormItemGi span="24 s:12 m:4">
                  <NButton size="small" type="error" ghost @click="removeFamily(index)">
                    {{ $t('common.delete') }}
                  </NButton>
                </NFormItemGi>
              </NGrid>
            </NCard>
          </NSpace>
        </NTabPane>
      </NTabs>
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
