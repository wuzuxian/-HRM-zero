<script setup lang="tsx">
import { computed, reactive, ref } from 'vue';
import { NButton, NPopconfirm, NTag } from 'naive-ui';
import { DeleteSalarySubject, GetSalarySubjectList } from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { hrmSalaryStatusRecord } from '@/constants/business';
import { useTable } from '@/hooks/common/table';
import SalarySubjectOperateModal from './modules/salary-subject-operate-modal.vue';
import SalarySubjectSearch from './modules/salary-subject-search.vue';

const appStore = useAppStore();

type LoadingStatus = Record<string, boolean>;
const deleteLoadingStatus = reactive<LoadingStatus>({});

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
  apiFn: GetSalarySubjectList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    subjectName: null,
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
      key: 'subjectName',
      title: $t('page.hrm.salary.subject.subjectName'),
      align: 'center',
      minWidth: 140
    },
    {
      key: 'orgName',
      title: $t('page.hrm.salary.subject.orgName'),
      align: 'center',
      minWidth: 120
    },
    {
      key: 'payAccount',
      title: $t('page.hrm.salary.subject.payAccount'),
      align: 'center',
      minWidth: 130
    },
    {
      key: 'payBank',
      title: $t('page.hrm.salary.subject.payBank'),
      align: 'center',
      minWidth: 110
    },
    {
      key: 'status',
      title: $t('page.hrm.salary.subject.status'),
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
      width: 130,
      render: row => (
        <div class="flex-center gap-8px">
          <NButton type="primary" ghost size="small" onClick={() => handleEdit(row)}>
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

// modal state (manual, same pattern as 假期类型 page)
const operateVisible = ref(false);
const operateType = ref<NaiveUI.TableOperateType>('add');
const editingData = ref<Api.Hrm.SalarySubject | null>(null);

function handleAdd() {
  operateType.value = 'add';
  editingData.value = null;
  operateVisible.value = true;
}

function handleEdit(row: Api.Hrm.SalarySubject) {
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
  const { error } = await DeleteSalarySubject(uuids);

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
  const { error } = await DeleteSalarySubject(uuids);
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
    <SalarySubjectSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <NCard
      :title="$t('page.hrm.salary.subject.title')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <TableHeaderOperation
          v-model:columns="columnChecks"
          :disabled-delete="!hasCheckedRows"
          :loading="loading"
          add-auth="hrm:salary_subject:add"
          delete-auth="hrm:salary_subject:delete"
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
        :scroll-x="1100"
        :loading="loading"
        remote
        :row-key="row => row.uuid"
        :pagination="mobilePagination"
        class="sm:h-full"
      />
      <SalarySubjectOperateModal
        v-model:visible="operateVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="handleSubmitted"
      />
    </NCard>
  </div>
</template>

<style scoped></style>
