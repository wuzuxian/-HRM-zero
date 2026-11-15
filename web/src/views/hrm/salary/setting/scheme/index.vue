<script setup lang="tsx">
import { computed, h, reactive, ref } from 'vue';
import { NButton, NPopconfirm, NSpace, NTag, NTooltip } from 'naive-ui';
import { DeleteSalaryScheme, GetSalarySchemeList } from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { hrmSalaryStatusRecord } from '@/constants/business';
import { useTable } from '@/hooks/common/table';
import SalarySchemeOperateModal from './modules/salary-scheme-operate-modal.vue';
import SalarySchemeSearch from './modules/salary-scheme-search.vue';

const appStore = useAppStore();

type LoadingStatus = Record<string, boolean>;
const deleteLoadingStatus = reactive<LoadingStatus>({});

/** 配置完整度 6 项 (项目/计薪/岗位/社保/个税/考勤); 岗位/社保/考勤 未接入, 值恒 0 → 灰态「未接入」 */
type CompletenessKey = keyof Api.Hrm.SalarySchemeCompleteness;

const completenessKeys: { key: CompletenessKey; labelKey: App.I18n.I18nKey; integrated: boolean }[] = [
  { key: 'schemeItems', labelKey: 'page.hrm.salary.scheme.completeness.items', integrated: true },
  { key: 'schemeCalc', labelKey: 'page.hrm.salary.scheme.completeness.calc', integrated: true },
  { key: 'schemePosition', labelKey: 'page.hrm.salary.scheme.completeness.position', integrated: false },
  { key: 'schemeSi', labelKey: 'page.hrm.salary.scheme.completeness.si', integrated: false },
  { key: 'schemeTax', labelKey: 'page.hrm.salary.scheme.completeness.tax', integrated: true },
  { key: 'schemeAttendance', labelKey: 'page.hrm.salary.scheme.completeness.attendance', integrated: false }
];

function completenessCell(completeness: Api.Hrm.SalarySchemeCompleteness) {
  return (
    <NSpace size={4} wrap={false} justify="center">
      {completenessKeys.map(({ key, labelKey, integrated }) => {
        const done = completeness[key] === 1;
        if (!integrated) {
          return (
            <NTooltip key={key}>
              {{
                trigger: () => (
                  <span class="text-12px text-gray-400">
                    {$t(labelKey)}·{$t('page.hrm.salary.scheme.completeness.notIntegrated')}
                  </span>
                ),
                default: () => $t('page.hrm.salary.scheme.completeness.notIntegratedTip')
              }}
            </NTooltip>
          );
        }
        return (
          <NTag key={key} size="small" type={done ? 'success' : 'default'}>
            {$t(labelKey)}
            {done ? '✓' : '✗'}
          </NTag>
        );
      })}
    </NSpace>
  );
}

const {
  columns,
  columnChecks,
  data,
  getData,
  getDataByPage,
  loading,
  mobilePagination,
  searchParams,
  resetSearchParams
} = useTable({
  apiFn: GetSalarySchemeList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    schemeName: null,
    subjectUuid: null,
    status: null
  },
  columns: () => [
    {
      type: 'selection',
      align: 'center',
      width: 48
    },
    {
      key: 'index',
      title: $t('common.index'),
      align: 'center',
      width: 64
    },
    {
      key: 'schemeName',
      title: $t('page.hrm.salary.scheme.schemeName'),
      align: 'center',
      minWidth: 140
    },
    {
      key: 'subjectName',
      title: $t('page.hrm.salary.scheme.subject'),
      align: 'center',
      minWidth: 110
    },
    {
      key: 'personStatus',
      title: $t('page.hrm.salary.scheme.personStatus'),
      align: 'center',
      minWidth: 130,
      render: row => {
        if (!row.personStatus) return null;
        const statusKeys: Record<string, string> = {
          '1': 'page.hrm.employee.personStatus.regular',
          '2': 'page.hrm.employee.personStatus.probation',
          '3': 'page.hrm.employee.personStatus.intern',
          '4': 'page.hrm.employee.personStatus.temp',
          '5': 'page.hrm.employee.personStatus.resigned',
          '6': 'page.hrm.employee.personStatus.retired'
        };
        return (
          <NSpace size={4} wrap justify="center">
            {row.personStatus.split(',').map(code => (
              <NTag key={code} size="small" bordered={false}>
                {statusKeys[code] ? $t(statusKeys[code] as App.I18n.I18nKey) : code}
              </NTag>
            ))}
          </NSpace>
        );
      }
    },
    {
      key: 'settleCycle',
      title: $t('page.hrm.salary.scheme.settleCycle'),
      align: 'center',
      width: 90,
      render: row => (row.settleCycle === 'monthly' ? $t('page.hrm.salary.scheme.settleCycleEnum.monthly') : row.settleCycle)
    },
    {
      key: 'payDay',
      title: $t('page.hrm.salary.scheme.payDay'),
      align: 'center',
      width: 80,
      render: row => h('span', `${row.payDay}`)
    },
    {
      key: 'currency',
      title: $t('page.hrm.salary.scheme.currency'),
      align: 'center',
      width: 80
    },
    {
      key: 'completeness',
      title: $t('page.hrm.salary.scheme.completeness.title'),
      align: 'center',
      minWidth: 260,
      render: row => completenessCell(row.completeness)
    },
    {
      key: 'itemCount',
      title: $t('page.hrm.salary.scheme.itemCount'),
      align: 'center',
      width: 80
    },
    {
      key: 'status',
      title: $t('page.hrm.salary.scheme.status'),
      align: 'center',
      width: 90,
      render: row => {
        if (row.status === null) {
          return null;
        }
        const tagMap: Record<Api.Hrm.SalarySubjectStatus, NaiveUI.ThemeColor> = {
          '1': 'success',
          '0': 'warning'
        };
        return <NTag type={tagMap[row.status]}>{$t(hrmSalaryStatusRecord[row.status])}</NTag>;
      }
    },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      width: 190,
      render: row => (
        <div class="flex-center gap-8px">
          <NButton type="primary" ghost size="small" onClick={() => handleEdit(row)}>
            {$t('page.hrm.salary.scheme.config')}
          </NButton>
          <NButton quaternary type="primary" size="small" onClick={() => handleEdit(row)}>
            {$t('common.edit')}
          </NButton>
          <NPopconfirm onPositiveClick={() => handleDelete(row.uuid)}>
            {{
              default: () => $t('common.confirmDelete'),
              trigger: () => (
                <NButton loading={deleteLoadingStatus[row.uuid]} type="error" ghost size="small">
                  {$t('common.delete')}
                </NButton>
              )
            }}
          </NPopconfirm>
        </div>
      )
    }
  ]
});

const checkedRowKeys = ref<string[]>([]);
const hasCheckedRows = computed(() => checkedRowKeys.value.length === 0);

// modal state (manual, same pattern as 薪酬主体 page)
const operateVisible = ref(false);
const operateType = ref<NaiveUI.TableOperateType>('add');
const editingData = ref<Api.Hrm.SalaryScheme | null>(null);

function handleAdd() {
  operateType.value = 'add';
  editingData.value = null;
  operateVisible.value = true;
}

function handleEdit(row: Api.Hrm.SalaryScheme) {
  operateType.value = 'edit';
  editingData.value = { ...row };
  operateVisible.value = true;
}

function handleSubmitted() {
  getDataByPage();
}

async function handleBatchDelete() {
  const uuids: string[] = [...checkedRowKeys.value];

  uuids.forEach(uuid => {
    deleteLoadingStatus[uuid] = true;
  });
  const { error } = await DeleteSalaryScheme(uuids);

  uuids.forEach(uuid => {
    deleteLoadingStatus[uuid] = false;
  });

  if (error) return;

  checkedRowKeys.value = [];
  window.$message?.success($t('common.deleteSuccess'));
  getData();
}

async function handleDelete(uuid: string) {
  const uuids: string[] = [uuid];
  deleteLoadingStatus[uuid] = true;
  const { error } = await DeleteSalaryScheme(uuids);
  deleteLoadingStatus[uuid] = false;
  if (error) {
    return;
  }
  window.$message?.success($t('common.deleteSuccess'));
  getData();
}
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <SalarySchemeSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <NCard
      :title="$t('page.hrm.salary.scheme.title')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <TableHeaderOperation
          v-model:columns="columnChecks"
          :disabled-delete="!hasCheckedRows"
          :loading="loading"
          add-auth="hrm:salary_scheme:add"
          delete-auth="hrm:salary_scheme:delete"
          @add="handleAdd"
          @delete="handleBatchDelete"
          @refresh="getData"
        />
      </template>
      <NDataTable
        v-model:checked-row-keys="checkedRowKeys"
        :columns="columns"
        :data="data"
        size="small"
        :flex-height="!appStore.isMobile"
        :scroll-x="1300"
        :loading="loading"
        remote
        :row-key="row => row.uuid"
        :pagination="mobilePagination"
        class="sm:h-full"
      />
      <SalarySchemeOperateModal
        v-model:visible="operateVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="handleSubmitted"
      />
    </NCard>
  </div>
</template>

<style scoped></style>
