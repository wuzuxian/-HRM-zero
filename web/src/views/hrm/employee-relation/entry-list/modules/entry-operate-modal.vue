<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import type { FormInst } from 'naive-ui';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { AddEntry, EditEntry } from '@/service/api';
import { $t } from '@/locales';
import { hrmEmployeeGenderOptions, hrmEmployeeMaritalStatusOptions } from '@/constants/business';
import { translateOptions } from '@/utils/common';

defineOptions({
  name: 'EntryOperateModal'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Hrm.HrmEntryBill | null;
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
    add: $t('page.hrm.employeeRelation.entry.addEntry'),
    edit: $t('page.hrm.employeeRelation.entry.editEntry')
  };
  return titles[props.operateType];
});

type Model = Api.Hrm.AddEntryRequest;

const model: Model = reactive(createDefaultModel());

function createDefaultModel(): Model {
  return {
    name: '',
    phone: '',
    email: '',
    gender: '',
    birthday: '',
    nation: '',
    politicalStatus: '',
    maritalStatus: '',
    cultureLevel: '',
    nativePlace: '',
    address: '',
    householdAddress: '',
    emergencyContact: '',
    emergencyPhone: '',
    bankName: '',
    bankAccount: '',
    avatar: '',
    position: '',
    post: '',
    probationMonths: 0,
    orgUuid: '',
    companyUuid: '',
    hireDate: '',
    expectFormalDate: '',
    attachmentDesc: '',
    remark: '',
    works: [],
    educations: [],
    families: []
  };
}

type RuleKey = Extract<keyof Model, 'name' | 'hireDate'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  name: defaultRequiredRule,
  hireDate: defaultRequiredRule
};

// ---- 三子表 (工作经历/教育经历/家庭成员) 可增删行 ----

function addRow<T extends Record<string, unknown>>(rows: T[], empty: () => T) {
  rows.push(empty());
}

function removeRow<T>(rows: T[], idx: number) {
  rows.splice(idx, 1);
}

function addWork() {
  addRow(model.works, () => ({ company: '', position: '', startDate: '', endDate: '', remark: '' }));
}

function addEducation() {
  addRow(model.educations, () => ({ school: '', major: '', education: '', startDate: '', endDate: '', remark: '' }));
}

function addFamily() {
  addRow(model.families, () => ({ name: '', relation: '', phone: '', workUnit: '', remark: '' }));
}

function handleInitModel() {
  Object.assign(model, createDefaultModel());

  if (props.operateType === 'edit' && props.rowData) {
    const { works, educations, families, ...rest } = props.rowData;
    Object.assign(model, rest);
    model.works = (works ?? []).map(item => ({ ...item }));
    model.educations = (educations ?? []).map(item => ({ ...item }));
    model.families = (families ?? []).map(item => ({ ...item }));
  }
}

function closeModel() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();
  confirmStartLoading();

  // 空行剔除后再提交
  const payload: Api.Hrm.AddEntryRequest = {
    ...model,
    works: model.works.filter(item => item.company || item.position),
    educations: model.educations.filter(item => item.school || item.major),
    families: model.families.filter(item => item.name || item.relation)
  };

  const request =
    props.operateType === 'add' ? AddEntry(payload) : EditEntry({ ...payload, uuid: props.rowData?.uuid ?? '' });

  const { error } = await request;
  if (!error) {
    window.$message?.success(props.operateType === 'add' ? $t('common.addSuccess') : $t('common.updateSuccess'));
    closeModel();
    emit('submitted');
  }
  confirmEndLoading();
}

const activeTab = ref('basic');

watch(visible, () => {
  if (visible.value) {
    handleInitModel();
    restoreValidation();
    activeTab.value = 'basic';
  }
});
</script>

<template>
  <NModal v-model:show="visible" :title="title" preset="card" class="w-860px">
    <NScrollbar class="max-h-520px pr-20px">
      <NTabs v-model:value="activeTab" type="line">
        <!-- Tab 1: 单据信息-基本信息 -->
        <NTabPane name="basic" :tab="$t('page.hrm.employeeRelation.entry.tabBasic')">
          <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="110">
            <NDivider :title="$t('page.hrm.employeeRelation.entry.sectionBasic')" title-placement="left" />
            <NGrid responsive="screen" item-responsive>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employeeRelation.entry.name')" path="name">
                <NInput v-model:value="model.name" :placeholder="$t('page.hrm.employeeRelation.entry.form.name')" />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employeeRelation.entry.gender')" path="gender">
                <NSelect
                  v-model:value="model.gender"
                  :options="translateOptions(hrmEmployeeGenderOptions)"
                  :placeholder="$t('page.hrm.employeeRelation.entry.form.gender')"
                  clearable
                />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employeeRelation.entry.birthday')" path="birthday">
                <NDatePicker
                  v-model:formatted-value="model.birthday"
                  type="date"
                  value-format="yyyy-MM-dd"
                  class="w-full"
                />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employeeRelation.entry.phone')" path="phone">
                <NInput v-model:value="model.phone" :placeholder="$t('page.hrm.employeeRelation.entry.form.phone')" />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employeeRelation.entry.email')" path="email">
                <NInput v-model:value="model.email" :placeholder="$t('page.hrm.employeeRelation.entry.form.email')" />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employeeRelation.entry.nation')" path="nation">
                <NInput v-model:value="model.nation" :placeholder="$t('page.hrm.employeeRelation.entry.form.nation')" />
              </NFormItemGi>
              <NFormItemGi
                span="24 s:12 m:8"
                :label="$t('page.hrm.employeeRelation.entry.politicalStatus')"
                path="politicalStatus"
              >
                <NInput
                  v-model:value="model.politicalStatus"
                  :placeholder="$t('page.hrm.employeeRelation.entry.form.politicalStatus')"
                />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employeeRelation.entry.maritalStatus')" path="maritalStatus">
                <NSelect
                  v-model:value="model.maritalStatus"
                  :options="translateOptions(hrmEmployeeMaritalStatusOptions)"
                  :placeholder="$t('page.hrm.employeeRelation.entry.form.maritalStatus')"
                  clearable
                />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employeeRelation.entry.cultureLevel')" path="cultureLevel">
                <NInput
                  v-model:value="model.cultureLevel"
                  :placeholder="$t('page.hrm.employeeRelation.entry.form.cultureLevel')"
                />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employeeRelation.entry.nativePlace')" path="nativePlace">
                <NInput v-model:value="model.nativePlace" :placeholder="$t('page.hrm.employeeRelation.entry.form.nativePlace')" />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employeeRelation.entry.address')" path="address">
                <NInput v-model:value="model.address" :placeholder="$t('page.hrm.employeeRelation.entry.form.address')" />
              </NFormItemGi>
              <NFormItemGi
                span="24 s:12 m:8"
                :label="$t('page.hrm.employeeRelation.entry.householdAddress')"
                path="householdAddress"
              >
                <NInput
                  v-model:value="model.householdAddress"
                  :placeholder="$t('page.hrm.employeeRelation.entry.form.householdAddress')"
                />
              </NFormItemGi>
              <NFormItemGi
                span="24 s:12 m:8"
                :label="$t('page.hrm.employeeRelation.entry.emergencyContact')"
                path="emergencyContact"
              >
                <NInput
                  v-model:value="model.emergencyContact"
                  :placeholder="$t('page.hrm.employeeRelation.entry.form.emergencyContact')"
                />
              </NFormItemGi>
              <NFormItemGi
                span="24 s:12 m:8"
                :label="$t('page.hrm.employeeRelation.entry.emergencyPhone')"
                path="emergencyPhone"
              >
                <NInput
                  v-model:value="model.emergencyPhone"
                  :placeholder="$t('page.hrm.employeeRelation.entry.form.emergencyPhone')"
                />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employeeRelation.entry.bankName')" path="bankName">
                <NInput v-model:value="model.bankName" :placeholder="$t('page.hrm.employeeRelation.entry.form.bankName')" />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employeeRelation.entry.bankAccount')" path="bankAccount">
                <NInput v-model:value="model.bankAccount" :placeholder="$t('page.hrm.employeeRelation.entry.form.bankAccount')" />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employeeRelation.entry.avatar')" path="avatar">
                <NInput v-model:value="model.avatar" :placeholder="$t('page.hrm.employeeRelation.entry.form.avatar')" />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employeeRelation.entry.position')" path="position">
                <NInput v-model:value="model.position" :placeholder="$t('page.hrm.employeeRelation.entry.form.position')" />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employeeRelation.entry.post')" path="post">
                <NInput v-model:value="model.post" :placeholder="$t('page.hrm.employeeRelation.entry.form.post')" />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employeeRelation.entry.orgUuid')" path="orgUuid">
                <NInput v-model:value="model.orgUuid" :placeholder="$t('page.hrm.employeeRelation.entry.form.orgUuid')" />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employeeRelation.entry.companyUuid')" path="companyUuid">
                <NInput v-model:value="model.companyUuid" :placeholder="$t('page.hrm.employeeRelation.entry.form.companyUuid')" />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employeeRelation.entry.hireDate')" path="hireDate">
                <NDatePicker
                  v-model:formatted-value="model.hireDate"
                  type="date"
                  value-format="yyyy-MM-dd"
                  class="w-full"
                />
              </NFormItemGi>
              <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.employeeRelation.entry.probationMonths')" path="probationMonths">
                <NInputNumber v-model:value="model.probationMonths" :min="0" class="w-full" />
              </NFormItemGi>
              <NFormItemGi
                span="24 s:12 m:8"
                :label="$t('page.hrm.employeeRelation.entry.expectFormalDate')"
                path="expectFormalDate"
              >
                <NDatePicker
                  v-model:formatted-value="model.expectFormalDate"
                  type="date"
                  value-format="yyyy-MM-dd"
                  class="w-full"
                />
              </NFormItemGi>
              <NFormItemGi span="24" :label="$t('page.hrm.employeeRelation.entry.remark')" path="remark">
                <NTextarea v-model:value="model.remark" :rows="2" :placeholder="$t('page.hrm.employeeRelation.entry.form.remark')" />
              </NFormItemGi>
            </NGrid>
          </NForm>
        </NTabPane>

        <!-- Tab 2: 工作经历 (子表可增行) -->
        <NTabPane name="works" :tab="$t('page.hrm.employeeRelation.entry.tabWorks')">
          <div class="mb-8px">
            <NButton size="small" type="primary" ghost @click="addWork">
              {{ $t('page.hrm.employeeRelation.entry.addRow') }}
            </NButton>
          </div>
          <NEmpty
            v-if="model.works.length === 0"
            :description="$t('page.hrm.employeeRelation.entry.noRows')"
            class="py-24px"
          />
          <NCard v-for="(work, idx) in model.works" :key="idx" size="small" class="mb-8px">
            <NGrid responsive="screen" item-responsive>
              <NGi span="24 s:12 m:6">
                <NFormItem :label="$t('page.hrm.employeeRelation.entry.workCompany')" label-placement="left" :label-width="80">
                  <NInput v-model:value="work.company" :placeholder="$t('page.hrm.employeeRelation.entry.form.workCompany')" />
                </NFormItem>
              </NGi>
              <NGi span="24 s:12 m:5">
                <NFormItem :label="$t('page.hrm.employeeRelation.entry.workPosition')" label-placement="left" :label-width="80">
                  <NInput v-model:value="work.position" :placeholder="$t('page.hrm.employeeRelation.entry.form.workPosition')" />
                </NFormItem>
              </NGi>
              <NGi span="24 s:12 m:5">
                <NFormItem :label="$t('page.hrm.employeeRelation.entry.startDate')" label-placement="left" :label-width="80">
                  <NDatePicker
                    v-model:formatted-value="work.startDate"
                    type="date"
                    value-format="yyyy-MM-dd"
                    class="w-full"
                  />
                </NFormItem>
              </NGi>
              <NGi span="24 s:12 m:5">
                <NFormItem :label="$t('page.hrm.employeeRelation.entry.endDate')" label-placement="left" :label-width="80">
                  <NDatePicker
                    v-model:formatted-value="work.endDate"
                    type="date"
                    value-format="yyyy-MM-dd"
                    class="w-full"
                  />
                </NFormItem>
              </NGi>
              <NGi span="24 s:12 m:3">
                <NButton size="small" type="error" ghost @click="removeRow(model.works, idx)">
                  {{ $t('common.delete') }}
                </NButton>
              </NGi>
            </NGrid>
          </NCard>
        </NTabPane>

        <!-- Tab 3: 教育经历 (子表可增行) -->
        <NTabPane name="educations" :tab="$t('page.hrm.employeeRelation.entry.tabEducations')">
          <div class="mb-8px">
            <NButton size="small" type="primary" ghost @click="addEducation">
              {{ $t('page.hrm.employeeRelation.entry.addRow') }}
            </NButton>
          </div>
          <NEmpty
            v-if="model.educations.length === 0"
            :description="$t('page.hrm.employeeRelation.entry.noRows')"
            class="py-24px"
          />
          <NCard v-for="(edu, idx) in model.educations" :key="idx" size="small" class="mb-8px">
            <NGrid responsive="screen" item-responsive>
              <NGi span="24 s:12 m:6">
                <NFormItem :label="$t('page.hrm.employeeRelation.entry.school')" label-placement="left" :label-width="80">
                  <NInput v-model:value="edu.school" :placeholder="$t('page.hrm.employeeRelation.entry.form.school')" />
                </NFormItem>
              </NGi>
              <NGi span="24 s:12 m:5">
                <NFormItem :label="$t('page.hrm.employeeRelation.entry.major')" label-placement="left" :label-width="80">
                  <NInput v-model:value="edu.major" :placeholder="$t('page.hrm.employeeRelation.entry.form.major')" />
                </NFormItem>
              </NGi>
              <NGi span="24 s:12 m:4">
                <NFormItem :label="$t('page.hrm.employeeRelation.entry.education')" label-placement="left" :label-width="80">
                  <NInput v-model:value="edu.education" :placeholder="$t('page.hrm.employeeRelation.entry.form.education')" />
                </NFormItem>
              </NGi>
              <NGi span="24 s:12 m:4">
                <NFormItem :label="$t('page.hrm.employeeRelation.entry.startDate')" label-placement="left" :label-width="80">
                  <NDatePicker
                    v-model:formatted-value="edu.startDate"
                    type="date"
                    value-format="yyyy-MM-dd"
                    class="w-full"
                  />
                </NFormItem>
              </NGi>
              <NGi span="24 s:12 m:4">
                <NFormItem :label="$t('page.hrm.employeeRelation.entry.endDate')" label-placement="left" :label-width="80">
                  <NDatePicker
                    v-model:formatted-value="edu.endDate"
                    type="date"
                    value-format="yyyy-MM-dd"
                    class="w-full"
                  />
                </NFormItem>
              </NGi>
              <NGi span="24 s:12 m:1">
                <NButton size="small" type="error" ghost @click="removeRow(model.educations, idx)">
                  {{ $t('common.delete') }}
                </NButton>
              </NGi>
            </NGrid>
          </NCard>
        </NTabPane>

        <!-- Tab 4: 家庭成员 (子表可增行) -->
        <NTabPane name="families" :tab="$t('page.hrm.employeeRelation.entry.tabFamilies')">
          <div class="mb-8px">
            <NButton size="small" type="primary" ghost @click="addFamily">
              {{ $t('page.hrm.employeeRelation.entry.addRow') }}
            </NButton>
          </div>
          <NEmpty
            v-if="model.families.length === 0"
            :description="$t('page.hrm.employeeRelation.entry.noRows')"
            class="py-24px"
          />
          <NCard v-for="(family, idx) in model.families" :key="idx" size="small" class="mb-8px">
            <NGrid responsive="screen" item-responsive>
              <NGi span="24 s:12 m:5">
                <NFormItem :label="$t('page.hrm.employeeRelation.entry.familyName')" label-placement="left" :label-width="80">
                  <NInput v-model:value="family.name" :placeholder="$t('page.hrm.employeeRelation.entry.form.familyName')" />
                </NFormItem>
              </NGi>
              <NGi span="24 s:12 m:5">
                <NFormItem :label="$t('page.hrm.employeeRelation.entry.relation')" label-placement="left" :label-width="80">
                  <NInput v-model:value="family.relation" :placeholder="$t('page.hrm.employeeRelation.entry.form.relation')" />
                </NFormItem>
              </NGi>
              <NGi span="24 s:12 m:6">
                <NFormItem :label="$t('page.hrm.employeeRelation.entry.familyPhone')" label-placement="left" :label-width="80">
                  <NInput v-model:value="family.phone" :placeholder="$t('page.hrm.employeeRelation.entry.form.familyPhone')" />
                </NFormItem>
              </NGi>
              <NGi span="24 s:12 m:6">
                <NFormItem :label="$t('page.hrm.employeeRelation.entry.workUnit')" label-placement="left" :label-width="80">
                  <NInput v-model:value="family.workUnit" :placeholder="$t('page.hrm.employeeRelation.entry.form.workUnit')" />
                </NFormItem>
              </NGi>
              <NGi span="24 s:12 m:2">
                <NButton size="small" type="error" ghost @click="removeRow(model.families, idx)">
                  {{ $t('common.delete') }}
                </NButton>
              </NGi>
            </NGrid>
          </NCard>
        </NTabPane>

        <!-- Tab 5: 附件信息 -->
        <NTabPane name="attachments" :tab="$t('page.hrm.employeeRelation.entry.tabAttachments')">
          <NForm label-placement="left" :label-width="110">
            <NFormItem :label="$t('page.hrm.employeeRelation.entry.attachmentDesc')" path="attachmentDesc">
              <NTextarea
                v-model:value="model.attachmentDesc"
                :rows="4"
                :placeholder="$t('page.hrm.employeeRelation.entry.form.attachmentDesc')"
              />
            </NFormItem>
          </NForm>
        </NTabPane>
      </NTabs>
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
