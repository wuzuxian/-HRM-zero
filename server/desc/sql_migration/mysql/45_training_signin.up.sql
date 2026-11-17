-- HRM-29 培训签到（动态码 / 补签 / 明细）
-- 签到数据落 hrm_training_session_participant（最小改动：名册=应到名单，签到状态即名册行上的列），
-- 课堂动态码单独建 hrm_training_signin_token（每班次单活跃 token，刷新=删旧插新，短时效轮换）。
-- 菜单 uuid 段：31xx（培训 3xx 块：页面 307，按钮 381-383；开工时已 grep 三方言迁移确认空闲，声明占用）。

ALTER TABLE `hrm_training_session_participant`
  ADD COLUMN `sign_in_time` datetime NULL COMMENT '签到时间（NULL=未签到）',
  ADD COLUMN `sign_in_method` varchar(1) NOT NULL DEFAULT '' COMMENT '签到方式 1=扫码自签 2=管理员补签（空=未签）',
  ADD COLUMN `sign_in_by` varchar(36) NOT NULL DEFAULT '' COMMENT '签到操作人（自签=本人 user_uuid，补签=管理员 user_uuid）',
  ADD COLUMN `sign_in_reason` varchar(200) NOT NULL DEFAULT '' COMMENT '补签原因';

CREATE TABLE `hrm_training_signin_token` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `session_uuid` varchar(36) NOT NULL COMMENT '班次 uuid',
  `token` varchar(64) NOT NULL COMMENT '课堂动态码 token（轮换）',
  `expire_time` datetime NOT NULL COMMENT '失效时间（短时效轮换）',
  `create_by` varchar(36) NOT NULL DEFAULT '' COMMENT '发起刷新的管理员 user_uuid',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_training_signin_token_session` (`session_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 演示数据：迁移 15 的已结束班次（...0302）的学员行补一条 补签 明细（方式 2 / 操作人 admin / 原因）
UPDATE hrm_training_session_participant
SET sign_in_time = '2026-10-10 08:55:00', sign_in_method = '2',
    sign_in_by = '1c2d3e4f-5a6b-7c8d-9e0f-1a2b3c4d5e6f', sign_in_reason = '现场网络故障，课后管理员补签'
WHERE uuid = 'aaa30000-3333-4000-8000-000000000312';

-- 培训签到 page（挂在 培训管理 目录 aaa10000-...000000000301 下，order 6）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000307','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000301','2','培训签到',0,'',6,'hrm_training_signin','/hrm/training/signin/manage','view.hrm_training_signin_manage','mdi:qrcode-scan','1','route.hrm_training_signin',0,'',0,0,'[]','[{"code":"v1:hrm:signin:getSigninDetailList","desc":"签到明细列表"},{"code":"v1:hrm:signin:getSigninStats","desc":"签到统计"},{"code":"v1:hrm:signin:refreshSigninToken","desc":"刷新课堂动态码"},{"code":"v1:hrm:signin:signinForParticipant","desc":"管理员补签"},{"code":"v1:hrm:signin:exportSigninDetail","desc":"导出签到明细"}]',0,'');

-- 扫码签到落地页（hide_in_menu=1 不进菜单，课堂动态码 URL 直达；同 1_init 用户详情隐藏页先例）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000308','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000307','2','扫码签到',1,'',0,'hrm_training_signin_scan','/hrm/training/signin/scan','view.hrm_training_signin_scan','','1','route.hrm_training_signin_scan',0,'',0,0,'[]','[]',1,'');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000381','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000307','3','刷新动态码',0,'',0,'','','','','1','button.hrm.training.signin.refreshToken',0,'',0,0,'[]','[{"code":"v1:hrm:signin:refreshSigninToken","desc":"刷新课堂动态码"}]',0,'v1:hrm:signin:refreshSigninToken'),
    ('aaa10000-1111-4000-8000-000000000382','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000307','3','管理员补签',0,'',1,'','','','','1','button.hrm.training.signin.signFor',0,'',0,0,'[]','[{"code":"v1:hrm:signin:signinForParticipant","desc":"管理员补签"}]',0,'v1:hrm:signin:signinForParticipant'),
    ('aaa10000-1111-4000-8000-000000000383','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000307','3','导出签到明细',0,'',2,'','','','','1','button.hrm.training.signin.export',0,'',0,0,'[]','[{"code":"v1:hrm:signin:exportSigninDetail","desc":"导出签到明细"}]',0,'v1:hrm:signin:exportSigninDetail');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000000308','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000308',0),
    ('aaa20000-1111-4000-8000-000000000307','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000307',0),
    ('aaa20000-1111-4000-8000-000000000381','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000381',0),
    ('aaa20000-1111-4000-8000-000000000382','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000382',0),
    ('aaa20000-1111-4000-8000-000000000383','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000383',0);
