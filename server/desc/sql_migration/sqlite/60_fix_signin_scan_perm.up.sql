-- fix: "扫码签到"菜单（45_training_signin 种子）permissions 为空数组。
--   route2code.go（生成产物）把 POST /api/v1/hrm/doSignin 映射为
--   v1:hrm:signin:doSignin，但该码不在任何菜单 permissions JSON 中
--   → casbin 全表无此策略，任何用户（含学员）调 doSignin 都被 Authx 裸 403，
--   学员扫码签到链路在 UI 上不可用（落地页永远落在错误态）。
--   doSignin 的真实授权是"jwt + 动态码 token + 学员名册"三重校验（do_signin.go），
--   casbin 只是网关层粗粒度门禁，故把 API 码补进"扫码签到"菜单 permissions，
--   并为绑定了该菜单的角色补 casbin_rule 存量行（与 role set_menus 的重建语义一致）。

-- casbin_rule 表由 casbin adapter 在应用首启时创建；全新环境上迁移先于首启执行，
-- 因此触及 casbin_rule 的迁移必须自带幂等建表（否则 fresh deploy 在此迁移即失败）。
CREATE TABLE IF NOT EXISTS casbin_rule (
    id     INTEGER PRIMARY KEY AUTOINCREMENT,
    p_type VARCHAR(32)  DEFAULT '' NOT NULL,
    v0     VARCHAR(255) DEFAULT '' NOT NULL,
    v1     VARCHAR(255) DEFAULT '' NOT NULL,
    v2     VARCHAR(255) DEFAULT '' NOT NULL,
    v3     VARCHAR(255) DEFAULT '' NOT NULL,
    v4     VARCHAR(255) DEFAULT '' NOT NULL,
    v5     VARCHAR(255) DEFAULT '' NOT NULL
);


UPDATE manage_menu SET permissions = '[{"code":"v1:hrm:signin:doSignin","desc":"学员扫码签到"}]'
WHERE route_path = '/hrm/training/signin/scan' AND permissions NOT LIKE '%doSignin%';

INSERT INTO casbin_rule (p_type, v0, v1)
SELECT DISTINCT 'p', rm.role_uuid, 'v1:hrm:signin:doSignin'
FROM manage_role_menu rm
WHERE rm.menu_uuid = (SELECT uuid FROM manage_menu WHERE route_path = '/hrm/training/signin/scan')
  AND NOT EXISTS (
    SELECT 1 FROM casbin_rule c
    WHERE c.p_type = 'p' AND c.v0 = rm.role_uuid AND c.v1 = 'v1:hrm:signin:doSignin'
  );
