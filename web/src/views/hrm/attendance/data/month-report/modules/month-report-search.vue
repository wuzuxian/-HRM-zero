<script setup lang="ts">
import { computed, ref } from 'vue';
import { GetAttendanceGroupList, GetOrgTree } from '@/service/api';
import { $t } from '@/locales';
import { useNaiveForm } from '@/hooks/common/form';

defineOptions({
  name: 'MonthReportSearch'
});

interface Emits {
  (e: 'reset'): void;
  (e: 'search'): void;
}

const emit = defineEmits<Emits>();

const { formRef, validate, restoreValidation } = useNaiveForm();

const model = defineModel<Api.Hrm.MonthReportSearchParams>('model', { required: true });

/** 月份选项（近 12 个月，倒序） */
const monthOptions = computed(() => {
  const options: { label: string; value: string }[] = [];
  const now = new Date();
  for (let i = 0; i < 12; i += 1) {
    const d = new Date(now.getFullYear(), now.getMonth() - i, 1);
    const value = `${d.getFullYear()}-${`${d.getMonth() + 1}`.padStart(2, '0')}`;
    options.push({ label: value, value });
  }
  return options;
});

// ---- 部门树（含下级，NTreeSelect） ----
interface OrgTreeOption {
  key: string;
  label: string;
  children?: OrgTreeOption[];
}

const orgTreeOptions = ref<OrgTreeOption[]>([]);

async function loadOrgTree() {
  const { error, data: tree } = await GetOrgTree();
  if (error || !tree) return;
  function toOption(node: Api.Hrm.Org): OrgTreeOption {
    return {
      key: node.uuid,
      label: node.orgName,
      children: (node.children ?? []).map(toOption)
    };
  }
  orgTreeOptions.value = (tree.records ?? []).map(toOption);
}
loadOrgTree();

// ---- 考勤组选项 ----
interface GroupOption {
  label: string;
  value: string;
}

const groupOptions = ref<GroupOption[]>([]);

async function loadGroups() {
  const { error, data } = await GetAttendanceGroupList({ current: 1, size: 100, status: '1' });
  if (error || !data) return;
  groupOptions.value = (data.records ?? []).map(item => ({
    label: item.groupName,
    value: item.groupName
  }));
}
loadGroups();

const onlyAbnormalOptions = [{ label: $t('page.hrm.attendance.monthReport.onlyAbnormalYes'), value: '1' }];
const onlyChangedOptions = [{ label: $t('page.hrm.attendance.monthReport.onlyChangedYes'), value: '1' }];

async function reset() {
  model.value.yearMonth = undefined;
  model.value.orgUuid = undefined;
  model.value.attGroupName = undefined;
  model.value.name = undefined;
  model.value.onlyAbnormal = undefined;
  model.value.onlyChanged = undefined;
  await restoreValidation();
  emit('reset');
}

async function search() {
  await validate();
  emit('search');
}
</script>

<template>
  <NCard :bordered="false" size="small" class="card-wrapper">
    <NCollapse>
      <NCollapseItem :title="$t('common.search')" name="month-report-search">
        <NForm ref="formRef" :model="model" label-placement="left" :label-width="90">
          <NGrid responsive="screen" item-responsive>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.attendance.monthReport.yearMonth')"
              path="yearMonth"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.yearMonth"
                :options="monthOptions"
                :placeholder="$t('page.hrm.attendance.monthReport.form.yearMonth')"
                clearable
              />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.attendance.monthReport.org')"
              path="orgUuid"
              class="pr-24px"
            >
              <NTreeSelect
                v-model:value="model.orgUuid"
                :options="orgTreeOptions"
                key-field="key"
                label-field="label"
                children-field="children"
                :placeholder="$t('page.hrm.attendance.monthReport.form.org')"
                clearable
              />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.attendance.monthReport.attGroup')"
              path="attGroupName"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.attGroupName"
                :options="groupOptions"
                :placeholder="$t('page.hrm.attendance.monthReport.form.attGroup')"
                clearable
                filterable
              />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.attendance.monthReport.name')"
              path="name"
              class="pr-24px"
            >
              <NInput v-model:value="model.name" :placeholder="$t('page.hrm.attendance.monthReport.form.name')" />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.attendance.monthReport.onlyAbnormal')"
              path="onlyAbnormal"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.onlyAbnormal"
                :options="onlyAbnormalOptions"
                :placeholder="$t('page.hrm.attendance.monthReport.form.onlyAbnormal')"
                clearable
              />
            </NFormItemGi>
            <NFormItemGi
              span="24 s:12 m:6"
              :label="$t('page.hrm.attendance.monthReport.onlyChanged')"
              path="onlyChanged"
              class="pr-24px"
            >
              <NSelect
                v-model:value="model.onlyChanged"
                :options="onlyChangedOptions"
                :placeholder="$t('page.hrm.attendance.monthReport.form.onlyChanged')"
                clearable
              />
            </NFormItemGi>
            <NFormItemGi span="24 m:12" class="pr-24px">
              <NSpace class="w-full" justify="end">
                <NButton @click="reset">{{ $t('common.reset') }}</NButton>
                <NButton type="primary" ghost @click="search">{{ $t('common.search') }}</NButton>
              </NSpace>
            </NFormItemGi>
          </NGrid>
        </NForm>
      </NCollapseItem>
    </NCollapse>
  </NCard>
</template>
