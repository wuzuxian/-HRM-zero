<script setup lang="ts">
import AssessmentTaskTable from '../modules/assessment-task-table.vue';

defineOptions({
  name: 'AssessmentSelfReviewList'
});
</script>

<template>
  <!-- 自评填报（demo-inventory 9.2 /hrm/performance/assessment-task/self-review-list） -->
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <AssessmentTaskTable stage="self-review" />
  </div>
</template>

<style scoped></style>
