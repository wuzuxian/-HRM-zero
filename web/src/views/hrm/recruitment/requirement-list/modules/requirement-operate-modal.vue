<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { AddRequirement, EditRequirement, GetOrgTree } from '@/service/api';
import { $t } from '@/locales';
import { hrmRequirementPostCategoryOptions, hrmRequirementUrgencyOptions } from '@/constants/business';
import { translateOptions } from '@/utils/common';

defineOptions({
  name: 'RequirementOperateModal'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Hrm.HrmRequirement | null;
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
    add: $t('page.hrm.requirement.addRequirement'),
    edit: $t('page.hrm.requirement.editRequirement')
  };
  return titles[props.operateType];
});

type Model = Api.Hrm.AddRequirementRequest;

const model: Model = reactive(createDefaultModel());

function createDefaultModel(): Model {
  return {
    postName: '',
    postCategory: 'tech',
    headcount: 1,
    urgency: 'normal',
    expectDate: '',
    orgUuid: '',
    orgName: '',
    requirementDesc: ''
  };
}

type RuleKey = Extract<keyof Model, 'postName' | 'postCategory' | 'headcount' | 'urgency' | 'orgUuid'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  postName: defaultRequiredRule,
  postCategory: defaultRequiredRule,
  headcount: defaultRequiredRule,
  urgency: defaultRequiredRule,
  orgUuid: defaultRequiredRule
};

// ---- org tree options (所属部门 from 组织管理) ----
interface OrgOption {
  label: string;
  value: string;
  children?: OrgOption[];
}

const orgOptions = ref<OrgOption[]>([]);

function toOrgOptions(nodes: Api.Hrm.Org[]): OrgOption[] {
  return nodes.map(node => ({
    label: node.orgName,
    value: node.uuid,
    children: node.children && node.children.length > 0 ? toOrgOptions(node.children) : undefined
  }));
}

async function loadOrgOptions() {
  if (orgOptions.value.length > 0) return;
  const { data, error } = await GetOrgTree();
  if (!error && data) {
    orgOptions.value = toOrgOptions(data.records ?? []);
  }
}

function handleOrgChange(_value: string, option: OrgOption | OrgOption[] | null) {
  if (option && !Array.isArray(option)) {
    model.orgName = option.label;
  }
}

function handleInitModel() {
  Object.assign(model, createDefaultModel());

  if (props.operateType === 'edit' && props.rowData) {
    Object.assign(model, props.rowData);
  }
  loadOrgOptions();
}

function closeModel() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();
  confirmStartLoading();

  const request =
    props.operateType === 'add'
      ? AddRequirement({ ...model })
      : EditRequirement({ ...model, uuid: props.rowData?.uuid ?? '' });

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
        <NDivider :title="$t('page.hrm.requirement.sectionPost')" title-placement="left" />
        <NGrid responsive="screen" item-responsive>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.requirement.postName')" path="postName">
            <NInput v-model:value="model.postName" :placeholder="$t('page.hrm.requirement.form.postName')" />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.requirement.postCategory.title')" path="postCategory">
            <NSelect
              v-model:value="model.postCategory"
              :options="translateOptions(hrmRequirementPostCategoryOptions)"
              :placeholder="$t('page.hrm.requirement.form.postCategory')"
            />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.requirement.headcount')" path="headcount">
            <NInputNumber v-model:value="model.headcount" :min="1" :precision="0" class="w-full" />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.requirement.urgency.title')" path="urgency">
            <NSelect
              v-model:value="model.urgency"
              :options="translateOptions(hrmRequirementUrgencyOptions)"
              :placeholder="$t('page.hrm.requirement.form.urgency')"
            />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.requirement.expectDate')" path="expectDate">
            <NDatePicker
              v-model:formatted-value="model.expectDate"
              type="date"
              value-format="yyyy-MM-dd"
              class="w-full"
            />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:8" :label="$t('page.hrm.requirement.orgName')" path="orgUuid">
            <NSelect
              v-model:value="model.orgUuid"
              :options="orgOptions"
              :placeholder="$t('page.hrm.requirement.form.org')"
              @update:value="handleOrgChange"
            />
          </NFormItemGi>
          <NFormItemGi span="24" :label="$t('page.hrm.requirement.requirementDesc')" path="requirementDesc">
            <NTextarea
              v-model:value="model.requirementDesc"
              :rows="2"
              :placeholder="$t('page.hrm.requirement.form.requirementDesc')"
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
