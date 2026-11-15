# jzero-admin

jzero-admin 后台管理系统的领域术语表（jzero 框架后端 + soybean-admin 前端）。

## Language

### 菜单与权限

**API 权限码**：
菜单（`manage_menu.permissions`）上声明的接口鉴权编码，携带 API 版本前缀（如 `v1:manage:user:add`），由 casbin 用于后端接口授权。
_Avoid_: 按钮编码、接口权限码

**按钮码**：
前端按钮级权限编码，使用裸码形式（如 `manage:user:add`），前端 `hasAuth` 据此控制界面元素显隐；存于 `manage_menu.button_code`，与 API 权限码是两套独立体系。
_Avoid_: 权限码、接口权限码

**按钮菜单**：
`menu_type='3'` 的菜单记录，承载按钮码，不产生路由；目录/菜单分别为 `menu_type='1'`/`'2'`。
_Avoid_: 按钮项
