-- HRM-04 员工↔系统账号双向同步:
-- HRM-03 落地时 生成用户 button menu (uuid aaa10000-...000416, permission code
-- v1:hrm:employee:generateUser) 已预种子, 但刻意未发 role_menu 行 (端点当时不存在).
-- 本迁移补发超级管理员的 role_menu 行, 使 /api/v1/hrm/generateUser 通过 casbin 校验.
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000000408','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000416',0);
