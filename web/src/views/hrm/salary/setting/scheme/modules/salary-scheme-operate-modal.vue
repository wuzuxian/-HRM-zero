<script setup lang="tsx">
import { computed, reactive, ref, watch } from 'vue';
import { NCheckbox, NTag } from 'naive-ui';
import { useLoading } from '@sa/hooks';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import {
  AddSalaryScheme,
  EditSalaryScheme,
  GetOrgTree,
  GetSalaryItemAllList,
  GetSalarySchemeDetail
} from '@/service/api';
import { $t } from '@/locales';
import {
  hrmSalaryItemClassRecord,
  hrmSalaryItemTypeRecord,
  hrmSalarySchemeCurrencyOptions,
  hrmSalarySchemeSettleCycleOptions,
  hrmSalaryStatusOptions
} from '@/constants/business';

defineOptions({
  name: 'SalarySchemeOperateModal'
});

/** 项目勾选列（DataTable 列定义在 setup 顶部, tsx render） */
const itemColumns = [
  {
    key: 'select',
    title: () => $t('page.hrm.salary.scheme.itemChecked'),
    align: 'center' as const,
    width: 70,
    render: (row: Api.Hrm.SalaryItem) => (
      <NCheckbox
        checked={isItemChecked(row.uuid)}
        onUpdateChecked={(checked: boolean) => toggleItem(row.uuid, checked)}
      />
    )
  },
  { key: 'itemName', title: () => $t('page.hrm.salary.item.itemName'), minWidth: 110 },
  { key: 'itemCode', title: () => $t('page.hrm.salary.item.itemCode'), minWidth: 130 },
  {
    key: 'itemType',
    title: () => $t('page.hrm.salary.item.itemType.title'),
    align: 'center' as const,
    width: 90,
    render: (row: Api.Hrm.SalaryItem) => $t(hrmSalaryItemTypeRecord[row.itemType])
  },
  {
    key: 'itemClass',
    title: () => $t('page.hrm.salary.item.itemClass.title'),
    align: 'center' as const,
    width: 80,
    render: (row: Api.Hrm.SalaryItem) => <NTag size="small">{$t(hrmSalaryItemClassRecord[row.itemClass])}</NTag>
  },
  { key: 'calcOrder', title: () => $t('page.hrm.salary.item.calcOrder'), align: 'center' as const, width: 90 }
];


interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Hrm.SalaryScheme | null;
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
    add: $t('page.hrm.salary.scheme.addScheme'),
    edit: $t('page.hrm.salary.scheme.editScheme')
  };
  return titles[props.operateType];
});

/** 人员状态多选: 1=正式 2=试用期 3=实习生 4=临时工 5=离职 6=退休 */
const personStatusOptions: CommonType.Option<string>[] = [
  { label: $t('page.hrm.employee.personStatus.regular'), value: '1' },
  { label: $t('page.hrm.employee.personStatus.probation'), value: '2' },
  { label: $t('page.hrm.employee.personStatus.intern'), value: '3' },
  { label: $t('page.hrm.employee.personStatus.temp'), value: '4' },
  { label: $t('page.hrm.employee.personStatus.resigned'), value: '5' },
  { label: $t('page.hrm.employee.personStatus.retired'), value: '6' }
];

type Model = {
  schemeName: string;
  subjectUuid: string;
  personStatus: string[];
  settleCycle: string;
  payDay: number | null;
  currency: string;
  status: Api.Hrm.SalarySubjectStatus;
  remark: string;
};

const model: Model = reactive(createDefaultModel());

function createDefaultModel(): Model {
  return {
    schemeName: '',
    subjectUuid: '',
    personStatus: [],
    settleCycle: 'monthly',
    payDay: 1,
    currency: 'CNY',
    status: '1',
    remark: ''
  };
}

type RuleKey = Extract<keyof Model, 'schemeName' | 'subjectUuid' | 'personStatus' | 'payDay'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  schemeName: defaultRequiredRule,
  subjectUuid: defaultRequiredRule,
  personStatus: {
    required: true,
    type: 'array',
    message: $t('page.hrm.salary.scheme.form.personStatusRequired'),
    trigger: 'change'
  },
  payDay: {
    required: true,
    type: 'number',
    validator: (_rule: unknown, value: number | null) => {
      if (value === null || value < 1 || value > 31) {
        return new Error($t('page.hrm.salary.scheme.form.payDayInvalid'));
      }
      return true;
    },
    trigger: 'blur'
  }
};

// ---- org tree options (薪酬主体 select, GetOrgTree 全树, 展平公司节点 orgType=1) ----
const subjectOptions = ref<CommonType.Option<string>[]>([]);

async function loadSubjectOptions() {
  const { data, error } = await GetOrgTree();
  if (!error && data) {
    const flat: CommonType.Option<string>[] = [];
    const walk = (nodes: Api.Hrm.Org[]) => {
      nodes.forEach(node => {
        if (node.orgType === '1') {
          flat.push({ label: node.orgName, value: node.uuid });
        }
        if (node.children && node.children.length > 0) {
          walk(node.children);
        }
      });
    };
    walk(data.records ?? []);
    subjectOptions.value = flat;
  }
}

// ---- 薪资项目全量清单（配置完整度: 项目✓ = 勾选 ≥1 项）----
const itemOptions = ref<Api.Hrm.SalaryItem[]>([]);

async function loadItemOptions() {
  const { error, data } = await GetSalaryItemAllList();
  if (!error) {
    itemOptions.value = data.records;
  }
}

/** 项目配置草稿: itemUuid → 覆盖值 */
type ItemDraft = {
  itemUuid: string | null;
  overrideAmount: number | null;
  overrideFormula: string;
};

const itemDrafts = ref<ItemDraft[]>([]);

function toggleItem(itemUuid: string, checked: boolean) {
  if (checked) {
    itemDrafts.value.push({ itemUuid, overrideAmount: null, overrideFormula: '' });
  } else {
    itemDrafts.value = itemDrafts.value.filter(d => d.itemUuid !== itemUuid);
  }
}

function isItemChecked(itemUuid: string) {
  return itemDrafts.value.some(d => d.itemUuid === itemUuid);
}

function draftOf(itemUuid: string) {
  return itemDrafts.value.find(d => d.itemUuid === itemUuid);
}

/** 勾选的项目行（含项目快照，用于展示类型/分类） */
const checkedItems = computed(() =>
  itemDrafts.value
    .map(draft => {
      const item = itemOptions.value.find(i => i.uuid === draft.itemUuid);
      return item ? { draft, item } : null;
    })
    .filter((row): row is { draft: ItemDraft; item: Api.Hrm.SalaryItem } => row !== null)
);

function handleInitModel() {
  Object.assign(model, createDefaultModel());
  itemDrafts.value = [];

  if (props.operateType === 'edit' && props.rowData?.uuid) {
    GetSalarySchemeDetail(props.rowData.uuid).then(({ error, data }) => {
      if (!error && data) {
        model.schemeName = data.scheme.schemeName;
        model.subjectUuid = data.scheme.subjectUuid;
        model.personStatus = data.scheme.personStatus ? data.scheme.personStatus.split(',') : [];
        model.settleCycle = data.scheme.settleCycle || 'monthly';
        model.payDay = data.scheme.payDay || 1;
        model.currency = data.scheme.currency || 'CNY';
        model.status = (data.scheme.status ?? '1') as Api.Hrm.SalarySubjectStatus;
        model.remark = data.scheme.remark;
        itemDrafts.value = data.items.map(item => ({
          itemUuid: item.itemUuid,
          overrideAmount: item.overrideAmount || null,
          overrideFormula: item.overrideFormula || ''
        }));
      }
    });
  }
}

function closeModel() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();

  const items: Api.Hrm.SalarySchemeItemPayload[] = itemDrafts.value.map(draft => ({
    itemUuid: draft.itemUuid as string,
    overrideAmount: draft.overrideAmount ?? 0,
    overrideFormula: draft.overrideFormula ?? ''
  }));

  confirmStartLoading();
  const request =
    props.operateType === 'add'
      ? AddSalaryScheme({
          schemeName: model.schemeName,
          subjectUuid: model.subjectUuid,
          personStatus: model.personStatus,
          settleCycle: model.settleCycle,
          payDay: model.payDay ?? 1,
          currency: model.currency,
          status: model.status,
          remark: model.remark,
          items
        })
      : EditSalaryScheme({
          uuid: props.rowData?.uuid ?? '',
          schemeName: model.schemeName,
          subjectUuid: model.subjectUuid,
          personStatus: model.personStatus,
          settleCycle: model.settleCycle,
          payDay: model.payDay ?? 1,
          currency: model.currency,
          status: model.status,
          remark: model.remark,
          items
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
    loadSubjectOptions();
    loadItemOptions();
  }
});
</script>

<template>
  <NModal v-model:show="visible" :title="title" preset="card" class="w-780px">
    <NScrollbar class="h-540px pr-20px">
      <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="110">
        <NFormItem :label="$t('page.hrm.salary.scheme.schemeName')" path="schemeName">
          <NInput v-model:value="model.schemeName" :placeholder="$t('page.hrm.salary.scheme.form.schemeName')" />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.scheme.subject')" path="subjectUuid">
          <NSelect
            v-model:value="model.subjectUuid"
            :options="subjectOptions"
            :placeholder="$t('page.hrm.salary.scheme.form.subject')"
            filterable
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.scheme.personStatus')" path="personStatus">
          <NSelect
            v-model:value="model.personStatus"
            multiple
            :options="personStatusOptions"
            :placeholder="$t('page.hrm.salary.scheme.form.personStatus')"
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.scheme.settleCycle')" path="settleCycle">
          <NSelect
            v-model:value="model.settleCycle"
            :options="hrmSalarySchemeSettleCycleOptions.map(o => ({ label: $t(o.label), value: o.value }))"
          />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.scheme.payDay')" path="payDay">
          <NInputNumber v-model:value="model.payDay" :min="1" :max="31" class="w-160px">
            <template #suffix>{{ $t('page.hrm.salary.scheme.payDayUnit') }}</template>
          </NInputNumber>
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.scheme.currency')" path="currency">
          <NSelect v-model:value="model.currency" :options="hrmSalarySchemeCurrencyOptions" />
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.scheme.status')" path="status">
          <NRadioGroup v-model:value="model.status">
            <NRadio
              v-for="item in hrmSalaryStatusOptions"
              :key="item.value"
              :value="item.value"
              :label="$t(item.label)"
            />
          </NRadioGroup>
        </NFormItem>
        <NFormItem :label="$t('page.hrm.salary.scheme.remark')" path="remark">
          <NInput v-model:value="model.remark" type="textarea" :placeholder="$t('page.hrm.salary.scheme.form.remark')" />
        </NFormItem>

        <NDivider :title="$t('page.hrm.salary.scheme.configItems')" title-placement="left" />
        <NAlert :type="itemDrafts.length > 0 ? 'success' : 'warning'" class="mb-12px">
          {{ $t('page.hrm.salary.scheme.itemCountTip') }}: {{ itemDrafts.length }}
        </NAlert>
        <NDataTable
          :columns="itemColumns"
          :data="itemOptions"
          size="small"
          max-height="260"
          :row-key="(row: Api.Hrm.SalaryItem) => row.uuid"
        />
        <template v-if="checkedItems.length > 0">
          <NDivider :title="$t('page.hrm.salary.scheme.overrideTitle')" title-placement="left" />
          <div v-for="row in checkedItems" :key="row.item.uuid" class="mb-12px flex items-center gap-12px">
            <span class="w-120px shrink-0 truncate">{{ row.item.itemName }}</span>
            <NInputNumber
              v-if="row.item.valueMode !== 'formula'"
              v-model:value="row.draft.overrideAmount"
              :min="0"
              :precision="2"
              class="w-200px"
              :placeholder="$t('page.hrm.salary.scheme.overrideAmountPlaceholder')"
            />
            <NInput
              v-else
              v-model:value="row.draft.overrideFormula"
              class="flex-1"
              :placeholder="$t('page.hrm.salary.scheme.overrideFormulaPlaceholder')"
            />
          </div>
        </template>
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
