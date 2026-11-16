<script setup lang="ts">
import AssessmentTaskTable from '../modules/assessment-task-table.vue';

defineOptions({
  name: 'AssessmentLeaderReviewList'
});
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <!-- 评审填报（demo-inventory 9.2 /hrm/performance/assessment-task/leader-review-list） -->
    <AssessmentTaskTable stage="leader-review" />
  </div>
</template>

<style scoped></style>
