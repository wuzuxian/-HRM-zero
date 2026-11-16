<script setup lang="ts">
import AssessmentTaskTable from '../modules/assessment-task-table.vue';

defineOptions({
  name: 'AssessmentPlanList'
});
</script>

<template>
  <!-- 目标确认（demo-inventory 9.2 /hrm/performance/assessment-task/plan-list） -->
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <AssessmentTaskTable stage="plan" />
  </div>
</template>

<style scoped></style>
