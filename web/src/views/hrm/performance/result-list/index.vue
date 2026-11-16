<script setup lang="tsx">
import { computed, onMounted, ref } from 'vue';
import { NButton, NPopconfirm, NTag } from 'naive-ui';
import { GetOrgTree, GetPerfResultList, SignPerfResult } from '@/service/api';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { useAuth } from '@/hooks/business/auth';
import type { NSwitch } from 'naive-ui';
import { perfResultStatusTag } from '@/constants/hrm-performance';
import { useTable } from '@/hooks/common/table';

defineOptions({
  name: 'PerfResultList'
});

const appStore = useAppStore();
const { hasAuth } = useAuth();

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
  apiFn: GetPerfResultList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    periodUuid: null,
    employeeUuid: null,
    companyUuid: null,
    orgUuid: null,
    signed: null,
    salarySynced: null,
    mine: false
  },
  columns: () => [
    {
      key: 'index',
      title: $t('common.index'),
      align: 'center',
      width: 64
    },
    {
      key: 'periodNo',
      title: $t('page.hrm.performance.result.period'),
      align: 'center',
      width: 150
    },
    {
      key: 'employeeName',
      title: $t('page.hrm.performance.result.employee'),
      align: 'center',
      width: 100
    },
    {
      key: 'companyName',
      title: $t('page.hrm.performance.result.company'),
      align: 'center',
      width: 110
    },
    {
      key: 'orgName',
      title: $t('page.hrm.performance.result.org'),
      align: 'center',
      width: 110
    },
    {
      key: 'score',
      title: $t('page.hrm.performance.result.score'),
      align: 'center',
      width: 80
    },
    {
      key: 'grade',
      title: $t('page.hrm.performance.result.grade'),
      align: 'center',
      width: 70
    },
    {
      key: 'coefficient',
      title: $t('page.hrm.performance.result.coefficient'),
      align: 'center',
      width: 80
    },
    {
      key: 'perfAmount',
      title: $t('page.hrm.performance.result.perfAmount'),
      align: 'center',
      width: 110
    },
    {
      key: 'bonusAmount',
      title: $t('page.hrm.performance.result.bonusAmount'),
      align: 'center',
      width: 110
    },
    {
      key: 'resultStatus',
      title: $t('page.hrm.performance.result.resultStatus'),
      align: 'center',
      width: 90,
      render: row => {
        const tag = perfResultStatusTag(row.resultStatus);
        if (!row.resultStatus || tag === null) {
          return null;
        }
        const label =
          row.resultStatus === '1'
            ? $t('page.hrm.performance.resultStatus.normal')
            : $t('page.hrm.performance.resultStatus.voided');
        return <NTag type={tag}>{label}</NTag>;
      }
    },
    {
      key: 'published',
      title: $t('page.hrm.performance.result.published'),
      align: 'center',
      width: 80,
      render: row => yesNoTag(row.published)
    },
    {
      key: 'signed',
      title: $t('page.hrm.performance.result.signed'),
      align: 'center',
      width: 80,
      render: row => yesNoTag(row.signed)
    },
    {
      key: 'salarySynced',
      title: $t('page.hrm.performance.result.salarySynced'),
      align: 'center',
      width: 100,
      render: row => yesNoTag(row.salarySynced)
    },
    {
      key: 'signTime',
      title: $t('page.hrm.performance.result.signTime'),
      align: 'center',
      width: 160
    },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      width: 100,
      render: row => {
        const canSign = row.published === '1' && row.signed === '0' && hasAuth('hrm:perf_result:sign');
        return (
          <div class="flex-center gap-8px">
            <NPopconfirm onPositiveClick={() => handleSign(row)}>
              {{
                default: () => $t('page.hrm.performance.result.signConfirm'),
                trigger: () => (
                  <NButton type="info" ghost size="small" disabled={!canSign}>
                    {$t('page.hrm.performance.result.sign')}
                  </NButton>
                )
              }}
            </NPopconfirm>
          </div>
        );
      }
    }
  ]
});

// ---- 公司/部门筛选下拉（HRM-46：结果行快照列 company_uuid/org_uuid 直读） ----

const orgTreeData = ref<Api.Hrm.Org[]>([]);

onMounted(async () => {
  const { error, data } = await GetOrgTree();
  if (!error && data) {
    orgTreeData.value = data.records;
  }
});

const companyOptions = computed<CommonType.Option<string>[]>(() =>
  orgTreeData.value
    .filter(item => item.orgType === '1')
    .map(item => ({ label: item.orgName, value: item.uuid }))
);

const orgOptions = computed<CommonType.Option<string>[]>(() => {
  const out: CommonType.Option<string>[] = [];
  function walk(node: Api.Hrm.Org) {
    if (node.orgType === '2') {
      out.push({ label: node.orgName, value: node.uuid });
    }
    (node.children ?? []).forEach(walk);
  }
  orgTreeData.value.forEach(walk);
  return out;
});

const salarySyncedOptions = computed(() => [
  { label: $t('page.hrm.performance.result.yes'), value: '1' },
  { label: $t('page.hrm.performance.result.no'), value: '0' }
]);

function yesNoTag(v: Api.Hrm.Performance.PerfYesNo) {
  if (v === '1') {
    return (
      <NTag type="success">{$t('page.hrm.performance.result.yes')}</NTag>
    );
  }
  return <NTag type="default">{$t('page.hrm.performance.result.no')}</NTag>;
}

async function handleSign(row: Api.Hrm.Performance.PerfResult) {
  const { error } = await SignPerfResult(row.uuid);
  if (!error) {
    window.$message?.success($t('page.hrm.performance.result.signSuccess'));
    getData();
  }
}
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <NCard :bordered="false" size="small" class="card-wrapper">
      <NForm label-placement="left" :label-width="80">
        <NGrid responsive="screen" item-responsive>
          <NFormItemGi
            span="24 s:12 m:6"
            :label="$t('page.hrm.performance.result.period')"
            path="periodUuid"
            class="pr-24px"
          >
            <NInput
              v-model:value="searchParams.periodUuid"
              :placeholder="$t('page.hrm.performance.result.periodFilter')"
              clearable
            />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:6"
            :label="$t('page.hrm.performance.dashboard.company')"
            path="companyUuid"
            class="pr-24px"
          >
            <NSelect
              v-model:value="searchParams.companyUuid"
              :options="companyOptions"
              :placeholder="$t('page.hrm.performance.dashboard.companyFilter')"
              clearable
              filterable
            />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:6"
            :label="$t('page.hrm.performance.dashboard.org')"
            path="orgUuid"
            class="pr-24px"
          >
            <NSelect
              v-model:value="searchParams.orgUuid"
              :options="orgOptions"
              :placeholder="$t('page.hrm.performance.dashboard.orgFilter')"
              clearable
              filterable
            />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:6"
            :label="$t('page.hrm.performance.result.salarySynced')"
            path="salarySynced"
            class="pr-24px"
          >
            <NSelect
              v-model:value="searchParams.salarySynced"
              :options="salarySyncedOptions"
              :placeholder="$t('page.hrm.performance.result.signedFilter')"
              clearable
            />
          </NFormItemGi>
          <NFormItemGi
            span="24 s:12 m:6"
            :label="$t('page.hrm.performance.result.signed')"
            path="signed"
            class="pr-24px"
          >
            <NSelect
              v-model:value="searchParams.signed"
              :placeholder="$t('page.hrm.performance.result.signedFilter')"
              :options="[
                { label: $t('page.hrm.performance.result.yes'), value: '1' },
                { label: $t('page.hrm.performance.result.no'), value: '0' }
              ]"
              clearable
            />
          </NFormItemGi>
          <NFormItemGi span="24 s:12 m:6" :label="$t('page.hrm.performance.result.mine')" class="pr-24px">
            <NSwitch :value="searchParams.mine === true" @update:value="(v: boolean) => (searchParams.mine = v)" />
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
    </NCard>
    <NCard
      :title="$t('page.hrm.performance.result.title')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <TableHeaderOperation v-model:columns="columnChecks" :loading="loading" @refresh="getData" />
      </template>
      <NDataTable
        :columns="columns"
        :data="data"
        size="small"
        :flex-height="!appStore.isMobile"
        :scroll-x="1600"
        :loading="loading"
        remote
        :row-key="row => row.uuid"
        :pagination="mobilePagination"
        class="sm:h-full"
      />
    </NCard>
  </div>
</template>

<style scoped></style>
