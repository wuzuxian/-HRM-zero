<script setup lang="ts">
import AssessmentTaskTable from '../modules/assessment-task-table.vue';

defineOptions({
  name: 'AssessmentResultConfirmList'
});
</script>

<template>
  <!-- 结果确认 + 申诉入口（demo-inventory 9.2 /hrm/performance/assessment-task/result-confirm-task-list） -->
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <AssessmentTaskTable stage="result" />
  </div>
</template>

<style scoped></style>
