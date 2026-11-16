-- fix: 考核任务目录菜单 component 错种为 view.hrm_performance_assessment-task
--   （该目录无 index.vue），动态路由 transform 时 getViewName 抛错 →
--   整棵子树（plan-list / self-review-list / leader-review-list /
--   result-confirm-task-list / workbench）未注册 → 菜单点击 No match for route。
--   目录菜单（menu_type='1'）应与其它 11 个目录一致使用 layout.base。

UPDATE manage_menu SET component = 'layout.base'
WHERE uuid = 'aaa10000-1111-4000-8000-000000003300'
  AND component = 'view.hrm_performance_assessment-task';
