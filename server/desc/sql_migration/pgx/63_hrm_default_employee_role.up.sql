-- HRM-49 开号接缝与默认员工角色 (spec #50 / ADR 0007): pgx 方言, 与 sqlite/mysql/63 同号同集合。
--   1. 种子「普通员工」角色 R_EMPLOYEE (状态启用)。
--   2. manage_role_menu 只授自服务页面: 首页、个人中心、考勤打卡、个人看板、
--      我的培训、目标确认、自评填报、结果确认 (+ 各页自服务按钮与目录链);
--      「评审填报」「考核任务工作台」属管理者视角, 不授。
--      动态路由只沿 role_menu 行建树, 目录 (menu_type='1') 不发则子页面挂不上,
--      因此人力资源/假勤管理/培训管理/绩效管理/考核任务五个目录随页面一并授予
--      (目录本身 permissions 为空, 不产生任何 casbin 策略)。
--   3. 为新角色补 casbin_rule 存量行 (应用首启 InitCasbinRule 也会重建, 双保险)。
--   4. 存量补挂: 零角色且已关联档案 (hrm_employee.user_uuid 指向它) 的账号
--      补挂 R_EMPLOYEE; 未关联档案的员工不自动建号 (避免迁移静默产出 123456 账号,
--      走批量「生成用户」, ADR 0007 Consequences)。
--   uuid 段: 角色 1a2b3c4d-...5c6e (R_SUPER ...5c6d +1);
--   role_menu 行 aaa20000-...42xx (空闲段, grep 全部迁移确认)。

CREATE TABLE IF NOT EXISTS casbin_rule (
    id     BIGSERIAL PRIMARY KEY,
    p_type VARCHAR(32)  DEFAULT '' NOT NULL,
    v0     VARCHAR(255) DEFAULT '' NOT NULL,
    v1     VARCHAR(255) DEFAULT '' NOT NULL,
    v2     VARCHAR(255) DEFAULT '' NOT NULL,
    v3     VARCHAR(255) DEFAULT '' NOT NULL,
    v4     VARCHAR(255) DEFAULT '' NOT NULL,
    v5     VARCHAR(255) DEFAULT '' NOT NULL
);

-- 默认角色 (幂等: 已存在则跳过)
INSERT INTO "manage_role" (uuid, create_time, update_time, name, status, code, "desc")
SELECT '1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e','2026-11-17 00:00:00','2026-11-17 00:00:00','普通员工','1','R_EMPLOYEE','员工自助页面 (开号默认角色)'
WHERE NOT EXISTS (SELECT 1 FROM "manage_role" WHERE code = 'R_EMPLOYEE');

-- 自服务页面授予 (42xx 段; is_home 首页行 4201)
INSERT INTO "manage_role_menu" (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
SELECT 'aaa20000-1111-4000-8000-000000004201','2026-11-17 00:00:00','2026-11-17 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e','f7e8d9c6-b5a4-4382-8271-605f4e3d2c1b',1
WHERE NOT EXISTS (SELECT 1 FROM "manage_role_menu" WHERE role_uuid = '1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e' AND menu_uuid = 'f7e8d9c6-b5a4-4382-8271-605f4e3d2c1b');

INSERT INTO "manage_role_menu" (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
SELECT 'aaa20000-1111-4000-8000-000000004202','2026-11-17 00:00:00','2026-11-17 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e','f5a6b7c8-d9e0-a1b2-c3f6-a7b8c9d0e1f2',0
WHERE NOT EXISTS (SELECT 1 FROM "manage_role_menu" WHERE role_uuid = '1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e' AND menu_uuid = 'f5a6b7c8-d9e0-a1b2-c3f6-a7b8c9d0e1f2');

-- 假勤管理目录 + 考勤打卡页 + 页内自服务按钮
INSERT INTO "manage_role_menu" (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
SELECT 'aaa20000-1111-4000-8000-000000004203','2026-11-17 00:00:00','2026-11-17 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e','aaa10000-1111-4000-8000-000000000101',0
WHERE NOT EXISTS (SELECT 1 FROM "manage_role_menu" WHERE role_uuid = '1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e' AND menu_uuid = 'aaa10000-1111-4000-8000-000000000101');

INSERT INTO "manage_role_menu" (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
SELECT 'aaa20000-1111-4000-8000-000000004204','2026-11-17 00:00:00','2026-11-17 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e','aaa10000-1111-4000-8000-000000002001',0
WHERE NOT EXISTS (SELECT 1 FROM "manage_role_menu" WHERE role_uuid = '1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e' AND menu_uuid = 'aaa10000-1111-4000-8000-000000002001');

INSERT INTO "manage_role_menu" (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
SELECT 'aaa20000-1111-4000-8000-000000004205','2026-11-17 00:00:00','2026-11-17 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e','aaa10000-1111-4000-8000-000000002011',0
WHERE NOT EXISTS (SELECT 1 FROM "manage_role_menu" WHERE role_uuid = '1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e' AND menu_uuid = 'aaa10000-1111-4000-8000-000000002011');

INSERT INTO "manage_role_menu" (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
SELECT 'aaa20000-1111-4000-8000-000000004206','2026-11-17 00:00:00','2026-11-17 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e','aaa10000-1111-4000-8000-000000002012',0
WHERE NOT EXISTS (SELECT 1 FROM "manage_role_menu" WHERE role_uuid = '1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e' AND menu_uuid = 'aaa10000-1111-4000-8000-000000002012');

-- 个人看板页 + 刷新按钮
INSERT INTO "manage_role_menu" (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
SELECT 'aaa20000-1111-4000-8000-000000004207','2026-11-17 00:00:00','2026-11-17 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e','aaa10000-1111-4000-8000-000000002801',0
WHERE NOT EXISTS (SELECT 1 FROM "manage_role_menu" WHERE role_uuid = '1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e' AND menu_uuid = 'aaa10000-1111-4000-8000-000000002801');

INSERT INTO "manage_role_menu" (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
SELECT 'aaa20000-1111-4000-8000-000000004208','2026-11-17 00:00:00','2026-11-17 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e','aaa10000-1111-4000-8000-000000002811',0
WHERE NOT EXISTS (SELECT 1 FROM "manage_role_menu" WHERE role_uuid = '1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e' AND menu_uuid = 'aaa10000-1111-4000-8000-000000002811');

-- 培训管理目录 + 我的培训页 + 下载证书按钮
INSERT INTO "manage_role_menu" (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
SELECT 'aaa20000-1111-4000-8000-000000004209','2026-11-17 00:00:00','2026-11-17 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e','aaa10000-1111-4000-8000-000000000301',0
WHERE NOT EXISTS (SELECT 1 FROM "manage_role_menu" WHERE role_uuid = '1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e' AND menu_uuid = 'aaa10000-1111-4000-8000-000000000301');

INSERT INTO "manage_role_menu" (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
SELECT 'aaa20000-1111-4000-8000-000000004210','2026-11-17 00:00:00','2026-11-17 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e','aaa10000-1111-4000-8000-000000000302',0
WHERE NOT EXISTS (SELECT 1 FROM "manage_role_menu" WHERE role_uuid = '1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e' AND menu_uuid = 'aaa10000-1111-4000-8000-000000000302');

INSERT INTO "manage_role_menu" (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
SELECT 'aaa20000-1111-4000-8000-000000004211','2026-11-17 00:00:00','2026-11-17 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e','aaa10000-1111-4000-8000-000000000311',0
WHERE NOT EXISTS (SELECT 1 FROM "manage_role_menu" WHERE role_uuid = '1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e' AND menu_uuid = 'aaa10000-1111-4000-8000-000000000311');

-- 人力资源根目录 + 绩效管理目录 + 考核任务目录 (动态路由目录链)
INSERT INTO "manage_role_menu" (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
SELECT 'aaa20000-1111-4000-8000-000000004212','2026-11-17 00:00:00','2026-11-17 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e','aaa10000-1111-4000-8000-000000000001',0
WHERE NOT EXISTS (SELECT 1 FROM "manage_role_menu" WHERE role_uuid = '1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e' AND menu_uuid = 'aaa10000-1111-4000-8000-000000000001');

INSERT INTO "manage_role_menu" (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
SELECT 'aaa20000-1111-4000-8000-000000004213','2026-11-17 00:00:00','2026-11-17 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e','f0e1d2c3-a4b5-4c6d-8e7f-0a1b2c3d4e5f',0
WHERE NOT EXISTS (SELECT 1 FROM "manage_role_menu" WHERE role_uuid = '1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e' AND menu_uuid = 'f0e1d2c3-a4b5-4c6d-8e7f-0a1b2c3d4e5f');

INSERT INTO "manage_role_menu" (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
SELECT 'aaa20000-1111-4000-8000-000000004214','2026-11-17 00:00:00','2026-11-17 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e','aaa10000-1111-4000-8000-000000003300',0
WHERE NOT EXISTS (SELECT 1 FROM "manage_role_menu" WHERE role_uuid = '1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e' AND menu_uuid = 'aaa10000-1111-4000-8000-000000003300');

-- 目标确认页 (3301: v1:hrm:perf_assessment:list/getDetail 页面码) + 确认目标按钮 (3312)
INSERT INTO "manage_role_menu" (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
SELECT 'aaa20000-1111-4000-8000-000000004215','2026-11-17 00:00:00','2026-11-17 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e','aaa10000-1111-4000-8000-000000003301',0
WHERE NOT EXISTS (SELECT 1 FROM "manage_role_menu" WHERE role_uuid = '1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e' AND menu_uuid = 'aaa10000-1111-4000-8000-000000003301');

INSERT INTO "manage_role_menu" (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
SELECT 'aaa20000-1111-4000-8000-000000004216','2026-11-17 00:00:00','2026-11-17 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e','aaa10000-1111-4000-8000-000000003312',0
WHERE NOT EXISTS (SELECT 1 FROM "manage_role_menu" WHERE role_uuid = '1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e' AND menu_uuid = 'aaa10000-1111-4000-8000-000000003312');

-- 自评填报页 (3302) + 自评提交按钮 (3313)
INSERT INTO "manage_role_menu" (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
SELECT 'aaa20000-1111-4000-8000-000000004217','2026-11-17 00:00:00','2026-11-17 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e','aaa10000-1111-4000-8000-000000003302',0
WHERE NOT EXISTS (SELECT 1 FROM "manage_role_menu" WHERE role_uuid = '1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e' AND menu_uuid = 'aaa10000-1111-4000-8000-000000003302');

INSERT INTO "manage_role_menu" (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
SELECT 'aaa20000-1111-4000-8000-000000004218','2026-11-17 00:00:00','2026-11-17 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e','aaa10000-1111-4000-8000-000000003313',0
WHERE NOT EXISTS (SELECT 1 FROM "manage_role_menu" WHERE role_uuid = '1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e' AND menu_uuid = 'aaa10000-1111-4000-8000-000000003313');

-- 结果确认页 (3304) + 确认结果按钮 (3316) + 发起申诉按钮 (3317)
INSERT INTO "manage_role_menu" (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
SELECT 'aaa20000-1111-4000-8000-000000004219','2026-11-17 00:00:00','2026-11-17 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e','aaa10000-1111-4000-8000-000000003304',0
WHERE NOT EXISTS (SELECT 1 FROM "manage_role_menu" WHERE role_uuid = '1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e' AND menu_uuid = 'aaa10000-1111-4000-8000-000000003304');

INSERT INTO "manage_role_menu" (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
SELECT 'aaa20000-1111-4000-8000-000000004220','2026-11-17 00:00:00','2026-11-17 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e','aaa10000-1111-4000-8000-000000003316',0
WHERE NOT EXISTS (SELECT 1 FROM "manage_role_menu" WHERE role_uuid = '1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e' AND menu_uuid = 'aaa10000-1111-4000-8000-000000003316');

INSERT INTO "manage_role_menu" (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
SELECT 'aaa20000-1111-4000-8000-000000004221','2026-11-17 00:00:00','2026-11-17 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e','aaa10000-1111-4000-8000-000000003317',0
WHERE NOT EXISTS (SELECT 1 FROM "manage_role_menu" WHERE role_uuid = '1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e' AND menu_uuid = 'aaa10000-1111-4000-8000-000000003317');

-- casbin 存量行: R_EMPLOYEE × 其被授页面/按钮的 permissions JSON 里的全部 API 码
-- (与 role set_menus / InitCasbinRule 的重建语义一致; 幂等)
INSERT INTO casbin_rule (p_type, v0, v1)
SELECT DISTINCT 'p', rm.role_uuid, pe.code
FROM manage_role_menu rm
JOIN manage_menu mm ON mm.uuid = rm.menu_uuid
CROSS JOIN LATERAL jsonb_to_recordset(mm.permissions::jsonb) AS pe(code VARCHAR(255))
WHERE rm.role_uuid = '1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e'
  AND pe.code <> ''
  AND NOT EXISTS (
    SELECT 1 FROM casbin_rule c
    WHERE c.p_type = 'p' AND c.v0 = rm.role_uuid AND c.v1 = pe.code
  );

-- 存量补挂: 零角色且已关联档案的账号 → R_EMPLOYEE (未关联档案的员工不自动建号)
INSERT INTO "manage_user_role" (uuid, create_time, update_time, user_uuid, role_uuid)
SELECT 'aaa30000-1111-4000-8000-000000004201','2026-11-17 00:00:00','2026-11-17 00:00:00', mu.uuid, '1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e'
FROM manage_user mu
WHERE EXISTS (
        SELECT 1 FROM hrm_employee he
        WHERE he.user_uuid = mu.uuid AND he.user_uuid <> '' AND he.person_status <> '5'
      )
  AND NOT EXISTS (SELECT 1 FROM manage_user_role mur WHERE mur.user_uuid = mu.uuid)
  AND NOT EXISTS (
    SELECT 1 FROM manage_user_role mur
    WHERE mur.user_uuid = mu.uuid
      AND mur.role_uuid = '1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6e'
  );
