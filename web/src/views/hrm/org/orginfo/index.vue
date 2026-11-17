<script setup lang="tsx">
import { computed, reactive, ref } from 'vue';
import { NButton, NPopconfirm, NTag } from 'naive-ui';
import { DeleteOrg, ExportOrg, GetOrgTree } from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { hrmOrgStatusRecord, hrmOrgTypeRecord } from '@/constants/business';
import { useAuth } from '@/hooks/business/auth';
import OrgOperateModal, { type OperateType } from './modules/org-operate-modal.vue';
import OrgImportModal from './modules/org-import-modal.vue';
import OrgSearch from './modules/org-search.vue';

defineOptions({
  name: 'HrmOrgOrginfo'
});

const appStore = useAppStore();
const { hasAuth } = useAuth();

const loading = ref(false);
const data = ref<Api.Hrm.Org[]>([]);
const searchParams = reactive<Api.Hrm.OrgSearchParams>({
  orgName: null,
  orgType: null,
  status: null
});
const checkedRowKeys = ref<string[]>([]);
const expandedKeys = ref<string[]>([]);
const collapsed = ref(false);

function collectExpandableUuids(orgs: Api.Hrm.Org[], acc: string[] = []) {
  orgs.forEach(org => {
    if (org.children?.length) {
      acc.push(org.uuid);
      collectExpandableUuids(org.children, acc);
    }
  });
  return acc;
}

function syncExpansion() {
  expandedKeys.value = collapsed.value ? [] : collectExpandableUuids(data.value);
}

async function getData() {
  loading.value = true;
  const { error, data: records } = await GetOrgTree(searchParams);
  loading.value = false;
  if (!error) {
    data.value = records?.records ?? [];
    syncExpansion();
  }
}

function handleResetSearch() {
  Object.assign(searchParams, { orgName: null, orgType: null, status: null });
  getData();
}

const hasCheckedRows = computed(() => checkedRowKeys.value.length === 0);

const deleteLoadingStatus = reactive<Record<string, boolean>>({});

async function handleBatchDelete() {
  const uuids = [...checkedRowKeys.value];
  uuids.forEach(uuid => {
    deleteLoadingStatus[uuid] = true;
  });
  const { error } = await DeleteOrg(uuids);
  uuids.forEach(uuid => {
    deleteLoadingStatus[uuid] = false;
  });
  if (error) return;
  checkedRowKeys.value = [];
  window.$message?.success($t('common.deleteSuccess'));
  getData();
}

async function handleDelete(uuid: string) {
  deleteLoadingStatus[uuid] = true;
  const { error } = await DeleteOrg([uuid]);
  deleteLoadingStatus[uuid] = false;
  if (error) return;
  window.$message?.success($t('common.deleteSuccess'));
  getData();
}

// ---- operate modal (新增组织/新增下级/编辑) ----
const operateVisible = ref(false);
const operateType = ref<OperateType>('add');
/** the edited org, or the parent org when adding a child */
const editingData = ref<Api.Hrm.Org | null>(null);

function handleAdd() {
  operateType.value = 'add';
  editingData.value = null;
  operateVisible.value = true;
}

function handleAddChild(org: Api.Hrm.Org) {
  operateType.value = 'addChild';
  editingData.value = { ...org };
  operateVisible.value = true;
}

function handleEdit(org: Api.Hrm.Org) {
  operateType.value = 'edit';
  editingData.value = { ...org };
  operateVisible.value = true;
}

// ---- import modal ----
const importVisible = ref(false);

// ---- export ----
async function handleExport() {
  const { error, data: exported } = await ExportOrg();
  if (error || !exported) return;
  const bin = atob(exported.content);
  const bytes = new Uint8Array(bin.length);
  for (let i = 0; i < bin.length; i += 1) {
    bytes[i] = bin.charCodeAt(i);
  }
  const blob = new Blob([bytes], {
    type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
  });
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement('a');
  anchor.href = url;
  anchor.download = exported.fileName;
  anchor.click();
  URL.revokeObjectURL(url);
}

// ---- columns: a check list for the column setting + the real columns ----
const columnChecks = ref<NaiveUI.TableColumnCheck[]>([
  { key: 'orgName', title: $t('page.hrm.org.orgName'), checked: true },
  { key: 'orgType', title: $t('page.hrm.org.orgTypeTitle'), checked: true },
  { key: 'leader', title: $t('page.hrm.org.leader'), checked: true },
  { key: 'orderNum', title: $t('page.hrm.org.orderNum'), checked: true },
  { key: 'status', title: $t('page.hrm.org.orgStatus'), checked: true },
  { key: 'createTime', title: $t('page.hrm.org.createTime'), checked: true }
]);

function buildColumns() {
  return [
    {
      type: 'selection' as const,
      align: 'center' as const,
      width: 48
    },
    {
      key: 'orgName',
      title: $t('page.hrm.org.orgName'),
      align: 'left' as const,
      minWidth: 200,
      tree: true
    },
    {
      key: 'orgType',
      title: $t('page.hrm.org.orgTypeTitle'),
      align: 'center' as const,
      width: 100,
      render: (row: Api.Hrm.Org) => {
        const tagMap: Record<Api.Hrm.OrgType, NaiveUI.ThemeColor> = {
          1: 'primary',
          2: 'default'
        };
        return <NTag type={tagMap[row.orgType]}>{$t(hrmOrgTypeRecord[row.orgType])}</NTag>;
      }
    },
    {
      key: 'leader',
      title: $t('page.hrm.org.leader'),
      align: 'center' as const,
      width: 120
    },
    {
      key: 'orderNum',
      title: $t('page.hrm.org.orderNum'),
      align: 'center' as const,
      width: 100
    },
    {
      key: 'status',
      title: $t('page.hrm.org.orgStatus'),
      align: 'center' as const,
      width: 100,
      render: (row: Api.Hrm.Org) => {
        if (row.status === null) {
          return null;
        }

        const tagMap: Record<Api.Hrm.OrgStatus, NaiveUI.ThemeColor> = {
          1: 'success',
          0: 'warning'
        };
        return <NTag type={tagMap[row.status]}>{$t(hrmOrgStatusRecord[row.status])}</NTag>;
      }
    },
    {
      key: 'createTime',
      title: $t('page.hrm.org.createTime'),
      align: 'center' as const,
      width: 170
    },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center' as const,
      width: 240,
      render: (row: Api.Hrm.Org) => (
        <div class="flex-center justify-end gap-8px">
          {hasAuth('hrm:org:addChild') && (
            <NButton type="primary" ghost size="small" onClick={() => handleAddChild(row)}>
              {$t('page.hrm.org.addChildOrg')}
            </NButton>
          )}
          {hasAuth('hrm:org:edit') && (
            <NButton type="primary" ghost size="small" onClick={() => handleEdit(row)}>
              {$t('common.edit')}
            </NButton>
          )}
          {hasAuth('hrm:org:delete') && (
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
          )}
        </div>
      )
    }
  ];
}

const columns = computed(() => {
  const visible = new Set(columnChecks.value.filter(item => item.checked).map(item => item.key));
  // 操作列（operate）不参与列设置过滤，只管数据列——否则操作按钮整列消失
  return buildColumns().filter(column => 'type' in column || column.key === 'operate' || visible.has(column.key!));
});

getData();
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <OrgSearch v-model:model="searchParams" @reset="handleResetSearch" @search="getData" />
    <NCard :title="$t('page.hrm.org.title')" :bordered="false" size="small" class="sm:flex-1-hidden card-wrapper">
      <template #header-extra>
        <NSpace align="center" wrap justify="end">
          <NButton v-if="hasAuth('hrm:org:import')" size="small" ghost type="primary" @click="importVisible = true">
            {{ $t('page.hrm.org.importOrg') }}
          </NButton>
          <NButton v-if="hasAuth('hrm:org:export')" size="small" ghost type="primary" @click="handleExport">
            {{ $t('page.hrm.org.exportOrg') }}
          </NButton>
          <TableHeaderOperation
            v-model:columns="columnChecks"
            :disabled-delete="hasCheckedRows"
            :loading="loading"
            add-auth="hrm:org:add"
            delete-auth="hrm:org:delete"
            @add="handleAdd"
            @delete="handleBatchDelete"
            @refresh="getData"
          >
            <template #prefix>
              <NButton size="small" @click="collapsed = !collapsed">
                {{ collapsed ? $t('page.hrm.org.expand') : $t('page.hrm.org.collapse') }}
              </NButton>
            </template>
          </TableHeaderOperation>
        </NSpace>
      </template>
      <NDataTable
        v-model:checked-row-keys="checkedRowKeys"
        v-model:expanded-row-keys="expandedKeys"
        :columns="columns"
        :data="data"
        size="small"
        :flex-height="!appStore.isMobile"
        :scroll-x="1100"
        :loading="loading"
        remote
        :row-key="row => row.uuid"
        class="sm:h-full"
      />
      <OrgOperateModal
        v-model:visible="operateVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="getData"
      />
      <OrgImportModal v-model:visible="importVisible" @submitted="getData" />
    </NCard>
  </div>
</template>

<style scoped></style>
