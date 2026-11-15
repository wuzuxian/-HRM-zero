<script setup lang="ts">
import { computed, provide, ref } from 'vue';
import { NButton, NEmpty, NSpin } from 'naive-ui';
import { GetOrgChartTree } from '@/service/api';
import { $t } from '@/locales';
import OrgChartNode from './modules/org-chart-node.vue';

defineOptions({
  name: 'HrmOrgOrgChart'
});

const loading = ref(false);
const records = ref<Api.Hrm.OrgChartNode[]>([]);

// ---- expand/collapse: the parent owns the collapsed-uuid set; recursive
// nodes toggle through the injected handlers ----
const collapsedUuids = ref(new Set<string>());

function toggle(uuid: string) {
  const next = new Set(collapsedUuids.value);
  if (next.has(uuid)) {
    next.delete(uuid);
  } else {
    next.add(uuid);
  }
  collapsedUuids.value = next;
}

function collectUuidsWithChildren(nodes: Api.Hrm.OrgChartNode[], acc: string[] = []) {
  nodes.forEach(node => {
    if (node.children?.length) {
      acc.push(node.uuid);
      collectUuidsWithChildren(node.children, acc);
    }
  });
  return acc;
}

const allCollapsed = computed(() => {
  const uuids = collectUuidsWithChildren(records.value);
  return uuids.length > 0 && uuids.every(uuid => collapsedUuids.value.has(uuid));
});

function handleExpandAll() {
  collapsedUuids.value = new Set<string>();
}

function handleCollapseAll() {
  collapsedUuids.value = new Set<string>(collectUuidsWithChildren(records.value));
}

provide('orgChartToggle', toggle);
provide('orgChartIsCollapsed', (uuid: string) => collapsedUuids.value.has(uuid));

// ---- pan & zoom (拖拽移动 / 滚轮缩放, demo-inventory 1.2) ----
const viewportRef = ref<HTMLElement | null>(null);
const scale = ref(1);
const translate = ref({ x: 0, y: 0 });
const panning = ref(false);

const MIN_SCALE = 0.4;
const MAX_SCALE = 2;

function handleWheel(event: WheelEvent) {
  event.preventDefault();
  const next = scale.value * (event.deltaY > 0 ? 0.9 : 1.1);
  scale.value = Math.min(MAX_SCALE, Math.max(MIN_SCALE, next));
}

function handlePointerDown(event: PointerEvent) {
  if (event.button !== 0) return;
  panning.value = true;
  (event.currentTarget as HTMLElement).setPointerCapture(event.pointerId);
  (event.currentTarget as HTMLElement).dataset.startX = String(event.clientX - translate.value.x);
  (event.currentTarget as HTMLElement).dataset.startY = String(event.clientY - translate.value.y);
}

function handlePointerMove(event: PointerEvent) {
  if (!panning.value) return;
  const target = event.currentTarget as HTMLElement;
  translate.value = {
    x: event.clientX - Number(target.dataset.startX),
    y: event.clientY - Number(target.dataset.startY)
  };
}

function handlePointerUp(event: PointerEvent) {
  panning.value = false;
  (event.currentTarget as HTMLElement).releasePointerCapture(event.pointerId);
}

async function getData() {
  loading.value = true;
  const { error, data } = await GetOrgChartTree();
  loading.value = false;
  if (!error) {
    records.value = data?.records ?? [];
    handleExpandAll();
    // 展示全图：居中起始视图
    scale.value = 1;
    translate.value = { x: 24, y: 24 };
  }
}

getData();
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <NCard
      :title="$t('page.hrm.org.orgChart.title')"
      :bordered="false"
      size="small"
      class="sm:flex-1-hidden card-wrapper"
    >
      <template #header-extra>
        <NSpace align="center" wrap justify="end">
          <NButton size="small" @click="allCollapsed ? handleExpandAll() : handleCollapseAll()">
            {{ allCollapsed ? $t('page.hrm.org.orgChart.expandAll') : $t('page.hrm.org.orgChart.collapseAll') }}
          </NButton>
          <NButton size="small" ghost type="primary" :loading="loading" @click="getData">
            {{ $t('page.hrm.org.orgChart.refresh') }}
          </NButton>
        </NSpace>
      </template>
      <div
        ref="viewportRef"
        class="relative h-600px overflow-auto lt-sm:h-500px sm:h-full"
        :class="panning ? 'cursor-grabbing' : 'cursor-grab'"
        @wheel="handleWheel"
        @pointerdown="handlePointerDown"
        @pointermove="handlePointerMove"
        @pointerup="handlePointerUp"
        @pointercancel="handlePointerUp"
      >
        <NSpin :show="loading" class="min-h-full w-full">
          <div
            v-if="records.length"
            class="org-chart-canvas inline-block min-w-full px-24px py-24px"
            :style="{ transform: `translate(${translate.x}px, ${translate.y}px) scale(${scale})` }"
          >
            <ul class="org-chart-root">
              <OrgChartNode v-for="node in records" :key="node.uuid" :node="node" />
            </ul>
          </div>
          <NEmpty v-else-if="!loading" :description="$t('page.hrm.org.orgChart.empty')" class="mt-120px" />
        </NSpin>
      </div>
    </NCard>
  </div>
</template>

<style scoped>
.org-chart-canvas {
  transform-origin: 0 0;
}

.org-chart-root,
.org-chart-root ul {
  display: flex;
  justify-content: center;
  list-style: none;
  margin: 0;
  padding: 0;
  position: relative;
}
</style>
