<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { AddPerfPeriod, EditPerfPeriod, GetOrgTree, GetPerfSchemeList } from '@/service/api';
import { $t } from '@/locales';
import {
  perfPeriodTypeOptions
} from '@/constants/hrm-performance';
import { hrmEmployeePersonStatusOptions as personStatusOptions } from '@/constants/business';
import { translateOptions } from '@/utils/common';

defineOptions({
  name: 'PerfPeriodOperateModal'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Hrm.Performance.PerfPeriod | null;
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
    add: $t('page.hrm.performance.period.addPeriod'),
    edit: $t('page.hrm.performance.period.editPeriod')
  };
  return titles[props.operateType];
});

type Model = {
  name: string;
  periodType: Api.Hrm.Performance.PeriodType;
  schemeUuid: string | null;
  companyUuids: string[];
  orgUuids: string[];
  positions: string[];
  posts: string[];
  personStatuses: string[];
  periodStart: number | null;
  periodEnd: number | null;
  resultPublishDate: number | null;
  remark: string;
};

const model: Model = reactive(createDefaultModel());

function createDefaultModel(): Model {
  return {
    name: '',
    periodType: 'month',
    schemeUuid: null,
    companyUuids: [],
    orgUuids: [],
    positions: [],
    posts: [],
    personStatuses: [],
    periodStart: null,
    periodEnd: null,
    resultPublishDate: null,
    remark: ''
  };
}

type RuleKey = Extract<keyof Model, 'name' | 'schemeUuid' | 'periodStart' | 'periodEnd'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  name: defaultRequiredRule,
  schemeUuid: defaultRequiredRule,
  periodStart: defaultRequiredRule,
  periodEnd: defaultRequiredRule
};

/** 绩效方案选项（引用 hrm_perf_scheme，取启用方案全量） */
const schemeOptions = ref<CommonType.Option<string>[]>([]);

async function getSchemeOptions() {
  const { error, data } = await GetPerfSchemeList({ current: 1, size: 1000 });
  if (!error && data) {
    schemeOptions.value = data.records
      .filter(item => item.status === '1')
      .map(item => ({
        label: item.name,
        value: item.uuid
      }));
  }
}

/** 组织树（公司/部门多维筛选，来自 HRM-01 组织管理） */
const orgTreeData = ref<Api.Hrm.Org[]>([]);

async function getOrgTree() {
  const { error, data } = await GetOrgTree();
  if (!error && data) {
    orgTreeData.value = data.records;
  }
}

type OrgTreeOption = {
  key: string;
  label: string;
  children?: OrgTreeOption[];
};

/** 公司（orgType=1）下拉选项 */
const companyOptions = computed<CommonType.Option<string>[]>(() =>
  orgTreeData.value
    .filter(item => item.orgType === '1')
    .map(item => ({ label: item.orgName, value: item.uuid }))
);

/** 部门树（orgType=2，NTreeSelect multiple） */
const departmentTreeOptions = computed<OrgTreeOption[]>(() => {
  function toOption(node: Api.Hrm.Org): OrgTreeOption {
    return {
      key: node.uuid,
      label: node.orgName,
      children: (node.children ?? []).filter(c => c.orgType === '2').map(toOption)
    };
  }
  const roots = orgTreeData.value
    .filter(item => item.orgType === '2')
    .map(toOption);
  // 公司节点下的部门也归入树
  for (const company of orgTreeData.value) {
    if (company.orgType === '1' && company.children?.length) {
      roots.push(...company.children.filter(c => c.orgType === '2').map(toOption));
    }
  }
  return roots;
});

/** 职位/职务 tag 选项（自由输入，暂不聚合历史取值） */

function handleInitModel() {
  Object.assign(model, createDefaultModel());

  if (props.operateType === 'edit' && props.rowData?.uuid) {
    model.name = props.rowData.name;
    model.periodType = props.rowData.periodType;
    model.schemeUuid = props.rowData.schemeUuid;
    model.periodStart = props.rowData.periodStart ? Date.parse(props.rowData.periodStart) : null;
    model.periodEnd = props.rowData.periodEnd ? Date.parse(props.rowData.periodEnd) : null;
    model.resultPublishDate = props.rowData.resultPublishDate
      ? Date.parse(props.rowData.resultPublishDate)
      : null;
    model.remark = props.rowData.remark;
    try {
      const scope = JSON.parse(props.rowData.scope || '{}') as Api.Hrm.Performance.PeriodScope;
      model.companyUuids = scope.companyUuids ?? [];
      model.orgUuids = scope.orgUuids ?? [];
      model.positions = scope.positions ?? [];
      model.posts = scope.posts ?? [];
      model.personStatuses = scope.personStatuses ?? [];
    } catch {
      // keep defaults on malformed scope
    }
  }
}

function buildScopeJson(): string {
  const scope: Api.Hrm.Performance.PeriodScope = {
    companyUuids: model.companyUuids,
    orgUuids: model.orgUuids,
    positions: model.positions,
    posts: model.posts,
    personStatuses: model.personStatuses
  };
  return JSON.stringify(scope);
}

function formatDate(ts: number | null): string {
  if (!ts) return '';
  const d = new Date(ts);
  const y = d.getFullYear();
  const m = `${d.getMonth() + 1}`.padStart(2, '0');
  const day = `${d.getDate()}`.padStart(2, '0');
  return `${y}-${m}-${day}`;
}

function closeModel() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();

  if (!model.schemeUuid) {
    window.$message?.error($t('page.hrm.performance.period.form.schemeUuid'));
    return;
  }

  confirmStartLoading();
  const { error } =
    props.operateType === 'add'
      ? await AddPerfPeriod({
          name: model.name,
          periodType: model.periodType,
          schemeUuid: model.schemeUuid,
          scope: buildScopeJson(),
          periodStart: formatDate(model.periodStart),
          periodEnd: formatDate(model.periodEnd),
          resultPublishDate: formatDate(model.resultPublishDate),
          remark: model.remark
        })
      : await EditPerfPeriod({
          uuid: props.rowData?.uuid,
          name: model.name,
          periodType: model.periodType,
          schemeUuid: model.schemeUuid,
          scope: buildScopeJson(),
          periodStart: formatDate(model.periodStart),
          periodEnd: formatDate(model.periodEnd),
          resultPublishDate: formatDate(model.resultPublishDate),
          remark: model.remark
        });
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
    getSchemeOptions();
    getOrgTree();
  }
});
</script>

<template>
  <NModal v-model:show="visible" :title="title" preset="card" class="w-680px">
    <NScrollbar class="h-520px pr-20px">
      <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="110">
        <NFormItem :label="$t('page.hrm.performance.period.name')" path="name">
          <NInput v-model:value="model.name" :placeholder="$t('page.hrm.performance.period.form.name')" />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.performance.period.periodType')" path="periodType">
          <NRadioGroup v-model:value="model.periodType">
            <NRadio
              v-for="item in translateOptions(perfPeriodTypeOptions)"
              :key="item.value"
              :value="item.value"
              :label="item.label"
            />
          </NRadioGroup>
        </NFormItem>
        <NFormItem :label="$t('page.hrm.performance.period.schemeName')" path="schemeUuid">
          <NSelect
            v-model:value="model.schemeUuid"
            :placeholder="$t('page.hrm.performance.period.form.schemeUuid')"
            :options="schemeOptions"
            filterable
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.performance.period.periodStart')" path="periodStart">
          <NDatePicker v-model:value="model.periodStart" type="date" class="w-full" />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.performance.period.periodEnd')" path="periodEnd">
          <NDatePicker v-model:value="model.periodEnd" type="date" class="w-full" />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.performance.period.resultPublishDate')" path="resultPublishDate">
          <NDatePicker v-model:value="model.resultPublishDate" type="date" class="w-full" clearable />
        </NFormItem>
        <NDivider>{{ $t('page.hrm.performance.period.scopeDivider') }}</NDivider>
        <NFormItem :label="$t('page.hrm.performance.period.companyUuids')" path="companyUuids">
          <NSelect
            v-model:value="model.companyUuids"
            :placeholder="$t('page.hrm.performance.period.form.companyUuids')"
            :options="companyOptions"
            multiple
            clearable
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.performance.period.orgUuids')" path="orgUuids">
          <NTreeSelect
            v-model:value="model.orgUuids"
            :placeholder="$t('page.hrm.performance.period.form.orgUuids')"
            :options="departmentTreeOptions"
            key-field="key"
            label-field="label"
            children-field="children"
            multiple
            clearable
            :max-tag-count="3"
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.performance.period.positions')" path="positions">
          <NSelect
            v-model:value="model.positions"
            :placeholder="$t('page.hrm.performance.period.form.positions')"
            :options="[]"
            multiple
            tag
            filterable
            clearable
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.performance.period.posts')" path="posts">
          <NSelect
            v-model:value="model.posts"
            :placeholder="$t('page.hrm.performance.period.form.posts')"
            :options="[]"
            multiple
            tag
            filterable
            clearable
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.performance.period.personStatuses')" path="personStatuses">
          <NSelect
            v-model:value="model.personStatuses"
            :placeholder="$t('page.hrm.performance.period.form.personStatuses')"
            :options="translateOptions(personStatusOptions)"
            multiple
            clearable
            :max-tag-count="6"
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.performance.period.remark')" path="remark">
          <NInput v-model:value="model.remark" :placeholder="$t('page.hrm.performance.period.form.remark')" />
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
