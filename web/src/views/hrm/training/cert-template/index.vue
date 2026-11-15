<script setup lang="tsx">
import { computed, reactive, ref } from 'vue';
import { NButton, NPopconfirm, NTag } from 'naive-ui';
import {
  DeleteCertTemplate,
  GetCertTemplateList,
  ToggleCertTemplateStatus
} from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import {
  hrmCertTemplateSourceRecord,
  hrmCertTemplateStatusRecord
} from '@/constants/business';
import { useTable } from '@/hooks/common/table';
import CertTemplateOperateModal from './modules/cert-template-operate-modal.vue';
import CertTemplateSearch from './modules/cert-template-search.vue';

const appStore = useAppStore();

type LoadingStatus = Record<string, boolean>;
const deleteLoadingStatus = reactive<LoadingStatus>({});
const toggleLoadingStatus = reactive<LoadingStatus>({});

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
  apiFn: GetCertTemplateList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    templateName: null,
    certType: null,
    source: null,
    status: null
  },
  columns: () => [
    {
      key: 'index',
      title: $t('common.index'),
      align: 'center',
      width: 64
    },
    {
      key: 'templateName',
      title: $t('page.hrm.training.certTemplate.templateName'),
      align: 'center',
      minWidth: 140
    },
    {
      key: 'certType',
      title: $t('page.hrm.training.certTemplate.certType'),
      align: 'center',
      minWidth: 110
    },
    {
      key: 'source',
      title: $t('page.hrm.training.certTemplate.sourceLabel'),
      align: 'center',
      width: 90,
      render: row => {
        if (row.source === null) {
          return null;
        }
        return $t(hrmCertTemplateSourceRecord[row.source]);
      }
    },
    {
      key: 'noPrefix',
      title: $t('page.hrm.training.certTemplate.noPrefix'),
      align: 'center',
      width: 100
    },
    {
      key: 'status',
      title: $t('page.hrm.training.certTemplate.statusLabel'),
      align: 'center',
      width: 90,
      render: row => {
        if (row.status === null) {
          return null;
        }
        const tagMap: Record<Api.Hrm.CertTemplateStatus, NaiveUI.ThemeColor> = {
          '1': 'success',
          '2': 'default'
        };
        return <NTag type={tagMap[row.status]}>{$t(hrmCertTemplateStatusRecord[row.status])}</NTag>;
      }
    },
    {
      key: 'remark',
      title: $t('page.hrm.training.certTemplate.remark'),
      align: 'center',
      minWidth: 120,
      render: row => row.remark || '-'
    },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      width: 200,
      fixed: 'right',
      render: row => renderOperate(row)
    }
  ]
});

function renderOperate(row: Api.Hrm.CertTemplate) {
  const enabled = row.status === '1';
  return (
    <div class="flex-center flex-wrap gap-4px">
      {enabled ? (
        <NPopconfirm onPositiveClick={() => handleToggle(row, '2')}>
          {{
            default: () => $t('page.hrm.training.certTemplate.confirmDisable'),
            trigger: () => (
              <NButton loading={toggleLoadingStatus[row.uuid]} type="warning" ghost size="small">
                {$t('page.hrm.training.certTemplate.disable')}
              </NButton>
            )
          }}
        </NPopconfirm>
      ) : (
        <NPopconfirm onPositiveClick={() => handleToggle(row, '1')}>
          {{
            default: () => $t('page.hrm.training.certTemplate.confirmEnable'),
            trigger: () => (
              <NButton loading={toggleLoadingStatus[row.uuid]} type="primary" ghost size="small">
                {$t('page.hrm.training.certTemplate.enable')}
              </NButton>
            )
          }}
        </NPopconfirm>
      )}
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
  );
}

// modal state (manual, same pattern as 培训班次 page)
const operateVisible = ref(false);
const operateType = ref<NaiveUI.TableOperateType>('add');
const editingData = ref<Api.Hrm.CertTemplate | null>(null);


function handleAdd() {
  operateType.value = 'add';
  editingData.value = null;
  operateVisible.value = true;
}

function handleEdit(row: Api.Hrm.CertTemplate) {
  operateType.value = 'edit';
  editingData.value = { ...row };
  operateVisible.value = true;
}

function handleSubmitted() {
  getDataByPage();
}

async function handleToggle(row: Api.Hrm.CertTemplate, status: Api.Hrm.CertTemplateStatus) {
  toggleLoadingStatus[row.uuid] = true;
  const { error } = await ToggleCertTemplateStatus({ uuid: row.uuid, status });
  toggleLoadingStatus[row.uuid] = false;
  if (error) {
    return;
  }
  window.$message?.success($t('common.updateSuccess'));
  getData();
}

const checkedRowKeys = ref<string[]>([]);

async function handleBatchDelete() {
  const uuids: string[] = [...checkedRowKeys.value];
  uuids.forEach(uuid => {
    deleteLoadingStatus[uuid] = true;
  });
  const { error } = await DeleteCertTemplate(uuids);
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
  const { error } = await DeleteCertTemplate(uuids);
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
    <CertTemplateSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <NCard
      :title="$t('page.hrm.training.certTemplate.title')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <TableHeaderOperation
          v-model:columns="columnChecks"
          :disabled-delete="!checkedRowKeys.length"
          :loading="loading"
          add-auth="hrm:cert_template:add"
          delete-auth="hrm:cert_template:delete"
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
        :scroll-x="1200"
        :loading="loading"
        remote
        :row-key="row => row.uuid"
        :pagination="mobilePagination"
        class="sm:h-full"
      />
      <CertTemplateOperateModal
        v-model:visible="operateVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="handleSubmitted"
      />
    </NCard>
  </div>
</template>

<style scoped></style>
