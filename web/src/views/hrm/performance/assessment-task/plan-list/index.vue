<script setup lang="ts">
import AssessmentTaskTable from '../modules/assessment-task-table.vue';

defineOptions({
  name: 'AssessmentPlanList'
});
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <!--
 目标确认（demo-inventory 9.2 /hrm/performance/assessment-task/plan-list）
      注意：注释不可放在 template 根部（组件根会变成含注释的 fragment），
      Transition out-in 离场 fragment 时新页面不会进场（页面空白），见 e2e/tests/bugs 
-->
    <AssessmentTaskTable stage="plan" />
  </div>
</template>

<style scoped></style>
