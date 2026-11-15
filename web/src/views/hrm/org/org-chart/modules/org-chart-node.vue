<script setup lang="ts">
import { computed, inject } from 'vue';
import { NTag } from 'naive-ui';
import { $t } from '@/locales';
import { hrmOrgStatusRecord, hrmOrgTypeRecord } from '@/constants/business';

defineOptions({
  name: 'OrgChartNode'
});

const props = defineProps<{
  node: Api.Hrm.OrgChartNode;
}>();

const toggle = inject<(uuid: string) => void>('orgChartToggle', () => undefined);
const isCollapsed = inject<(uuid: string) => boolean>('orgChartIsCollapsed', () => false);

const hasChildren = computed(() => Boolean(props.node.children?.length));
const expanded = computed(() => hasChildren.value && !isCollapsed(props.node.uuid));
const childCount = computed(() => props.node.children?.length ?? 0);

const orgTypeName = computed(() => $t(hrmOrgTypeRecord[props.node.orgType] || 'page.hrm.org.orgType.department'));
</script>

<template>
  <li class="org-chart-node">
    <div class="org-chart-card">
      <div class="flex items-center gap-8px">
        <span class="text-15px font-600">{{ node.orgName }}</span>
        <NTag size="tiny" :bordered="false" round>
          {{ orgTypeName }}
        </NTag>
        <NTag v-if="node.status === '0'" size="tiny" type="warning" :bordered="false" round>
          {{ $t(hrmOrgStatusRecord['0']) }}
        </NTag>
      </div>
      <div class="mt-6px text-12px text-gray-500 dark:text-gray-400">
        {{ $t('page.hrm.org.leader') }}: {{ node.leader || $t('page.hrm.org.orgChart.noLeader') }}
      </div>
      <div class="mt-8px flex gap-8px">
        <span
          class="rounded-4px bg-gray-100 px-8px py-2px text-12px dark:bg-gray-700"
          :title="$t('page.hrm.org.orgChart.memberCountTip')"
        >
          {{ $t('page.hrm.org.orgChart.memberCount') }} {{ node.memberCount }}
        </span>
        <span
          class="rounded-4px bg-gray-100 px-8px py-2px text-12px dark:bg-gray-700"
          :title="$t('page.hrm.org.orgChart.totalMemberCountTip')"
        >
          {{ $t('page.hrm.org.orgChart.totalMemberCount') }} {{ node.totalMemberCount }}
        </span>
      </div>
    </div>
    <div v-if="hasChildren" class="flex justify-center">
      <button class="org-chart-toggle" type="button" @click.stop="toggle(node.uuid)">
        <template v-if="expanded">−</template>
        <template v-else>+{{ childCount }}</template>
      </button>
    </div>
    <ul v-if="expanded" class="org-chart-children">
      <OrgChartNode v-for="child in node.children" :key="child.uuid" :node="child" />
    </ul>
  </li>
</template>

<style scoped>
.org-chart-node {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 24px 12px 0;
}

/* horizontal connector: left half (to previous sibling) */
.org-chart-node::before {
  content: '';
  position: absolute;
  top: 0;
  right: 50%;
  width: 50%;
  height: 24px;
  border-top: 1px solid var(--org-chart-line);
}

/* horizontal connector: right half (to next sibling) + vertical drop to this card */
.org-chart-node::after {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  width: 50%;
  height: 24px;
  border-top: 1px solid var(--org-chart-line);
  border-left: 1px solid var(--org-chart-line);
}

/* trim the dangling halves at both ends */
.org-chart-children > .org-chart-node:first-child::before {
  border-top: 0 none;
}

/* last child: keep only the vertical drop (::after border-left) */
.org-chart-children > .org-chart-node:last-child::after {
  border-top: 0 none;
}

/* single child: straight vertical drop only */
.org-chart-children > .org-chart-node:only-child::before {
  border: 0 none;
}

.org-chart-children > .org-chart-node:only-child::after {
  border-top: 0 none;
}

/* stub from the parent card down to the children row */
.org-chart-children::before {
  content: '';
  position: absolute;
  top: -24px;
  left: 50%;
  width: 0;
  height: 24px;
  border-left: 1px solid var(--org-chart-line);
}

.org-chart-card {
  min-width: 180px;
  max-width: 240px;
  padding: 10px 14px;
  border: 1px solid rgb(16 185 129 / 35%);
  border-radius: 8px;
  background-color: rgb(255 255 255);
  box-shadow: 0 1px 4px rgb(0 0 0 / 8%);
  z-index: 1;
}

:global(.dark) .org-chart-card {
  background-color: rgb(36 36 50);
}

.org-chart-toggle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 20px;
  height: 18px;
  margin-top: 2px;
  padding: 0 6px;
  border: 1px solid var(--org-chart-line);
  border-radius: 9px;
  background-color: rgb(255 255 255);
  color: inherit;
  cursor: pointer;
  font-size: 11px;
  line-height: 1;
  z-index: 1;
}

:global(.dark) .org-chart-toggle {
  background-color: rgb(36 36 50);
}

:global(:root) {
  --org-chart-line: #c2c7cf;
}

:global(.dark) {
  --org-chart-line: #4c515b;
}
</style>
