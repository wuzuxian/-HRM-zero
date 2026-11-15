<script setup lang="tsx">
import { computed, reactive, ref, watch } from 'vue';
import { NButton, NPopconfirm, NTag } from 'naive-ui';
import { useLoading } from '@sa/hooks';
import {
  AddEmployeeCertificate,
  DeleteEmployeeCertificate,
  EditEmployeeCertificate,
  GetEmployeeCertificateList,
  GetEmployeeList,
  GetOrgTree
} from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { useTable } from '@/hooks/common/table';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { certExpiryStateOptions } from '@/constants/business';
import { translateOptions } from '@/utils/common';

defineOptions({
  name: 'CertificateExpiring'
});

const appStore = useAppStore();

type DeleteLoadingStatus = Record<string, boolean>;
const deleteLoadingStatus = reactive<DeleteLoadingStatus>({});

function expiryStateTag(row: Api.Hrm.HrmEmployeeCertificate) {
  const tagMap: Record<string, NaiveUI.ThemeColor> = {
    normal: 'default',
    warning: 'warning',
    expired: 'error'
  };
  const labelMap: Record<string, App.I18n.I18nKey> = {
    normal: 'page.hrm.employee.certExpiring.stateNormal',
    warning: 'page.hrm.employee.certExpiring.stateWarning',
    expired: 'page.hrm.employee.certExpiring.stateExpired'
  };
  return <NTag type={tagMap[row.expiryState || 'normal']}>{$t(labelMap[row.expiryState || 'normal'])}</NTag>;
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
  apiFn: GetEmployeeCertificateList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    employeeName: null,
    employeeNo: null,
    certType: null,
    expiryState: null
  },
  columns: () => [
    {
      key: 'index',
      title: $t('common.index'),
      align: 'center',
      width: 64
    },
    {
      key: 'employeeNo',
      title: $t('page.hrm.employee.certExpiring.employeeNo'),
      align: 'center',
      width: 100
    },
    {
      key: 'employeeName',
      title: $t('page.hrm.employee.certExpiring.employeeName'),
      align: 'center',
      width: 90
    },
    {
      key: 'orgUuid',
      title: $t('page.hrm.employee.certExpiring.orgName'),
      align: 'center',
      width: 130,
      render: row => orgNameByUuid(row.orgUuid) || '-'
    },
    {
      key: 'certType',
      title: $t('page.hrm.employee.certExpiring.certType'),
      align: 'center',
      width: 130,
      render: row => row.certType || '-'
    },
    {
      key: 'certNo',
      title: $t('page.hrm.employee.certExpiring.certNo'),
      align: 'center',
      width: 190,
      render: row => row.certNo || '-'
    },
    {
      key: 'expiryDate',
      title: $t('page.hrm.employee.certExpiring.expiryDate'),
      align: 'center',
      width: 110,
      render: row => row.expiryDate || '-'
    },
    {
      key: 'daysToExpiry',
      title: $t('page.hrm.employee.certExpiring.daysToExpiry'),
      align: 'center',
      width: 110,
      render: row => String(row.daysToExpiry)
    },
    {
      key: 'expiryState',
      title: $t('page.hrm.employee.certExpiring.status'),
      align: 'center',
      width: 100,
      render: row => expiryStateTag(row)
    },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      width: 160,
      render: row => renderOperate(row)
    }
  ]
});

function renderOperate(row: Api.Hrm.HrmEmployeeCertificate) {
  return (
    <div class="flex-center flex-wrap gap-4px">
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

// ---- org tree (部门名称解析) ----
const orgOptions = ref<Api.Hrm.Org[]>([]);

async function loadOrgTree() {
  const { error, data: tree } = await GetOrgTree();
  if (!error && tree) {
    orgOptions.value = tree.records;
  }
}

const orgFlatMap = computed<Map<string, string>>(() => {
  const map = new Map<string, string>();
  const walk = (nodes: Api.Hrm.Org[]) => {
    for (const node of nodes) {
      map.set(node.uuid, node.orgName);
      if (node.children?.length) walk(node.children);
    }
  };
  walk(orgOptions.value);
  return map;
});

function orgNameByUuid(uuid: string) {
  return orgFlatMap.value.get(uuid) ?? '';
}

loadOrgTree();

// ---- maintain modal (证照维护入口, spec 用户故事 17) ----
const operateVisible = ref(false);
const operateType = ref<NaiveUI.TableOperateType>('add');
const editingData = ref<Api.Hrm.HrmEmployeeCertificate | null>(null);

function handleAdd() {
  operateType.value = 'add';
  editingData.value = null;
  operateVisible.value = true;
}

function handleEdit(row: Api.Hrm.HrmEmployeeCertificate) {
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
  const { error } = await DeleteEmployeeCertificate(uuids);
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
  const { error } = await DeleteEmployeeCertificate(uuids);
  deleteLoadingStatus[uuid] = false;
  if (error) return;
  window.$message?.success($t('common.deleteSuccess'));
  getData();
}

const checkedRowKeys = ref<string[]>([]);
const hasCheckedRows = computed(() => checkedRowKeys.value.length === 0);

// ---- inline maintain modal ----
const { formRef: certFormRef, validate: certValidate, restoreValidation: certRestoreValidation } = useNaiveForm();
const { defaultRequiredRule } = useFormRules();
const { loading: confirmLoading, startLoading: confirmStartLoading, endLoading: confirmEndLoading } = useLoading();

type CertModel = Api.Hrm.AddEmployeeCertificateRequest;

const certModel: CertModel = reactive(createDefaultCertModel());

function createDefaultCertModel(): CertModel {
  return {
    employeeUuid: '',
    certType: '',
    certNo: '',
    issueDate: '',
    expiryDate: '',
    remark: ''
  };
}

const certTitle = computed(() => {
  const titles: Record<NaiveUI.TableOperateType, string> = {
    add: $t('page.hrm.employee.certExpiring.addCert'),
    edit: $t('page.hrm.employee.certExpiring.editCert')
  };
  return titles[operateType.value];
});

interface EmployeeOption {
  label: string;
  value: string;
}

const employeeOptions = ref<EmployeeOption[]>([]);

async function loadEmployeeOptions() {
  if (employeeOptions.value.length > 0) return;
  const { data: empData, error } = await GetEmployeeList({ current: 1, size: 200 });
  if (error || !empData) return;
  employeeOptions.value = (empData.records ?? []).map(rec => ({
    label: `${rec.employeeNo} ${rec.name}`,
    value: rec.uuid
  }));
}

function handleInitCertModel() {
  Object.assign(certModel, createDefaultCertModel());
  if (operateType.value === 'edit' && editingData.value) {
    Object.assign(certModel, {
      employeeUuid: editingData.value.employeeUuid,
      certType: editingData.value.certType,
      certNo: editingData.value.certNo, // 脱敏回显；原样提交视为未修改
      issueDate: editingData.value.issueDate,
      expiryDate: editingData.value.expiryDate,
      remark: editingData.value.remark
    });
  }
  loadEmployeeOptions();
  certRestoreValidation();
}

async function handleCertSubmit() {
  await certValidate();
  confirmStartLoading();
  const request =
    operateType.value === 'add'
      ? AddEmployeeCertificate({ ...certModel })
      : EditEmployeeCertificate({ ...certModel, uuid: editingData.value?.uuid ?? '' });
  const { error } = await request;
  if (!error) {
    window.$message?.success(operateType.value === 'add' ? $t('common.addSuccess') : $t('common.updateSuccess'));
    operateVisible.value = false;
    handleSubmitted();
  }
  confirmEndLoading();
}

const certRules: Record<'employeeUuid' | 'certType' | 'certNo', App.Global.FormRule> = {
  employeeUuid: defaultRequiredRule,
  certType: defaultRequiredRule,
  certNo: defaultRequiredRule
};

watch(operateVisible, () => {
  if (operateVisible.value) {
    handleInitCertModel();
  }
});
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <NCard :bordered="false" size="small" class="card-wrapper">
      <NCollapse>
        <NCollapseItem :title="$t('common.search')" name="cert-expiring-search">
          <NForm :model="searchParams" label-placement="left" :label-width="90">
            <NGrid responsive="screen" item-responsive>
              <NFormItemGi
                span="24 s:12 m:6"
                :label="$t('page.hrm.employee.certExpiring.employeeName')"
                path="employeeName"
                class="pr-24px"
              >
                <NInput v-model:value="searchParams.employeeName" clearable />
              </NFormItemGi>
              <NFormItemGi
                span="24 s:12 m:6"
                :label="$t('page.hrm.employee.certExpiring.employeeNo')"
                path="employeeNo"
                class="pr-24px"
              >
                <NInput v-model:value="searchParams.employeeNo" clearable />
              </NFormItemGi>
              <NFormItemGi
                span="24 s:12 m:6"
                :label="$t('page.hrm.employee.certExpiring.certType')"
                path="certType"
                class="pr-24px"
              >
                <NInput v-model:value="searchParams.certType" clearable />
              </NFormItemGi>
              <NFormItemGi
                span="24 s:12 m:6"
                :label="$t('page.hrm.employee.certExpiring.status')"
                path="expiryState"
                class="pr-24px"
              >
                <NSelect
                  v-model:value="searchParams.expiryState"
                  :options="translateOptions(certExpiryStateOptions)"
                  clearable
                />
              </NFormItemGi>
              <NFormItemGi span="24 m:6" class="pr-24px">
                <NSpace class="w-full" justify="end">
                  <NButton @click="resetSearchParams">
                    <template #icon>
                      <icon-ic-round-refresh class="text-icon" />
                    </template>
                    {{ $t('common.reset') }}
                  </NButton>
                  <NButton type="primary" ghost @click="() => getDataByPage()">
                    <template #icon>
                      <icon-ic-round-search class="text-icon" />
                    </template>
                    {{ $t('common.search') }}
                  </NButton>
                </NSpace>
              </NFormItemGi>
            </NGrid>
          </NForm>
        </NCollapseItem>
      </NCollapse>
    </NCard>
    <NCard
      :title="$t('page.hrm.employee.certExpiring.title')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <TableHeaderOperation
          v-model:columns="columnChecks"
          :disabled-delete="!hasCheckedRows"
          :loading="loading"
          add-auth="hrm:empcert:add"
          delete-auth="hrm:empcert:delete"
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

      <NModal v-model:show="operateVisible" preset="card" class="w-700px" :title="certTitle">
        <NForm ref="certFormRef" :model="certModel" :rules="certRules" label-placement="left" :label-width="90">
          <NGrid responsive="screen" item-responsive>
            <NFormItemGi
              span="24 s:12 m:12"
              :label="$t('page.hrm.employee.certExpiring.selectEmployee')"
              path="employeeUuid"
            >
              <NSelect
                v-model:value="certModel.employeeUuid"
                :options="employeeOptions"
                :placeholder="$t('page.hrm.employee.certExpiring.form.selectEmployee')"
                :disabled="operateType === 'edit'"
                filterable
              />
            </NFormItemGi>
            <NFormItemGi span="24 s:12 m:12" :label="$t('page.hrm.employee.certExpiring.certType')" path="certType">
              <NInput
                v-model:value="certModel.certType"
                :placeholder="$t('page.hrm.employee.certExpiring.form.certType')"
              />
            </NFormItemGi>
            <NFormItemGi span="24 s:12 m:12" :label="$t('page.hrm.employee.certExpiring.certNo')" path="certNo">
              <NInput
                v-model:value="certModel.certNo"
                :placeholder="$t('page.hrm.employee.certExpiring.form.certNo')"
              />
            </NFormItemGi>
            <NFormItemGi span="24 s:12 m:12" :label="$t('page.hrm.employee.certExpiring.issueDate')" path="issueDate">
              <NDatePicker
                v-model:formatted-value="certModel.issueDate"
                type="date"
                value-format="yyyy-MM-dd"
                clearable
                class="w-full"
              />
            </NFormItemGi>
            <NFormItemGi span="24 s:12 m:12" :label="$t('page.hrm.employee.certExpiring.expiryDate')" path="expiryDate">
              <NDatePicker
                v-model:formatted-value="certModel.expiryDate"
                type="date"
                value-format="yyyy-MM-dd"
                clearable
                class="w-full"
              />
            </NFormItemGi>
            <NFormItemGi span="24" :label="$t('page.hrm.employee.certExpiring.remark')" path="remark">
              <NInput
  type="textarea" v-model:value="certModel.remark" :rows="2" />
            </NFormItemGi>
          </NGrid>
        </NForm>
        <template #footer>
          <NSpace justify="end" :size="16">
            <NButton @click="operateVisible = false">{{ $t('common.cancel') }}</NButton>
            <NButton type="primary" :loading="confirmLoading" @click="handleCertSubmit">
              {{ $t('common.save') }}
            </NButton>
          </NSpace>
        </template>
      </NModal>
    </NCard>
  </div>
</template>

<style scoped></style>
