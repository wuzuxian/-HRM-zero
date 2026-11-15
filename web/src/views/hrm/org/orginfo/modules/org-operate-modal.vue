<script setup lang="ts">
import { computed, reactive, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { AddOrg, AddOrgChild, EditOrg } from '@/service/api';
import { $t } from '@/locales';
import { hrmOrgStatusOptions, hrmOrgTypeOptions } from '@/constants/business';
import { translateOptions } from '@/utils/common';

defineOptions({
  name: 'OrgOperateModal'
});

export type OperateType = 'add' | 'addChild' | 'edit';

interface Props {
  /** the type of operation */
  operateType: OperateType;
  /** the edited org, or the parent org when adding a child */
  rowData?: Api.Hrm.Org | null;
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
  const titles: Record<OperateType, string> = {
    add: $t('page.hrm.org.addOrg'),
    addChild: $t('page.hrm.org.addChildOrg'),
    edit: $t('page.hrm.org.editOrg')
  };
  return titles[props.operateType];
});

const parentOrgName = computed(() => (props.operateType === 'addChild' ? (props.rowData?.orgName ?? '') : ''));

type Model = Pick<Api.Hrm.AddOrgRequest, 'orgName' | 'orgType' | 'leader' | 'orderNum' | 'status' | 'remark'>;

const model: Model = reactive(createDefaultModel());

function createDefaultModel(): Model {
  return {
    orgName: '',
    orgType: '2',
    leader: '',
    orderNum: 0,
    status: '1',
    remark: ''
  };
}

type RuleKey = Extract<keyof Model, 'orgName' | 'orgType'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  orgName: defaultRequiredRule,
  orgType: defaultRequiredRule
};

function handleInitModel() {
  Object.assign(model, createDefaultModel());

  if (props.operateType === 'edit' && props.rowData) {
    Object.assign(model, {
      orgName: props.rowData.orgName,
      orgType: props.rowData.orgType,
      leader: props.rowData.leader,
      orderNum: props.rowData.orderNum,
      status: props.rowData.status,
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

  if (props.operateType === 'add') {
    const { error } = await AddOrg({
      parentUuid: '',
      orgName: model.orgName,
      orgType: model.orgType,
      leader: model.leader,
      orderNum: model.orderNum,
      status: model.status,
      remark: model.remark
    });
    if (!error) {
      window.$message?.success($t('common.addSuccess'));
      closeModel();
      emit('submitted');
    }
  } else if (props.operateType === 'addChild') {
    const { error } = await AddOrgChild({
      parentUuid: props.rowData?.uuid ?? '',
      orgName: model.orgName,
      orgType: model.orgType,
      leader: model.leader,
      orderNum: model.orderNum,
      status: model.status,
      remark: model.remark
    });
    if (!error) {
      window.$message?.success($t('common.addSuccess'));
      closeModel();
      emit('submitted');
    }
  } else {
    const { error } = await EditOrg({
      uuid: props.rowData?.uuid ?? '',
      orgName: model.orgName,
      orgType: model.orgType,
      leader: model.leader,
      orderNum: model.orderNum,
      status: model.status,
      remark: model.remark
    });
    if (!error) {
      window.$message?.success($t('common.updateSuccess'));
      closeModel();
      emit('submitted');
    }
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
  <NModal v-model:show="visible" :title="title" preset="card" class="w-600px">
    <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="100">
      <NFormItem v-if="operateType === 'addChild'" :label="$t('page.hrm.org.parentOrg')" path="parentOrg">
        <NInput :value="parentOrgName" disabled />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.org.orgName')" path="orgName">
        <NInput v-model:value="model.orgName" :placeholder="$t('page.hrm.org.form.orgName')" />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.org.orgTypeTitle')" path="orgType">
        <NRadioGroup v-model:value="model.orgType">
          <NRadio v-for="item in translateOptions(hrmOrgTypeOptions)" :key="item.value" :value="item.value">
            {{ item.label }}
          </NRadio>
        </NRadioGroup>
      </NFormItem>
      <NFormItem :label="$t('page.hrm.org.leader')" path="leader">
        <NInput v-model:value="model.leader" :placeholder="$t('page.hrm.org.form.leader')" />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.org.orderNum')" path="orderNum">
        <NInputNumber v-model:value="model.orderNum" :placeholder="$t('page.hrm.org.form.orderNum')" class="w-full" />
      </NFormItem>
      <NFormItem :label="$t('page.hrm.org.orgStatus')" path="status">
        <NRadioGroup v-model:value="model.status">
          <NRadio v-for="item in translateOptions(hrmOrgStatusOptions)" :key="item.value" :value="item.value">
            {{ item.label }}
          </NRadio>
        </NRadioGroup>
      </NFormItem>
      <NFormItem :label="$t('page.hrm.org.remark')" path="remark">
        <NInput v-model:value="model.remark" type="textarea" :placeholder="$t('page.hrm.org.form.remark')" />
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
