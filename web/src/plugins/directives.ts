import type { App, Directive } from 'vue';
import { useAuth } from '@/hooks/business/auth';

/**
 * v-auth 权限指令：当前用户不持有 codes 中任一权限码时移除元素。
 *
 * 用法： <NButton v-auth="'hrm:punch:add'">新增</NButton> <NButton v-auth="['hrm:punch:add',
 * 'hrm:punch:edit']">任一命中即显示</NButton>
 *
 * codes 为 string 或 string[]；权限判定复用 useAuth().hasAuth（authStore.userInfo.buttons）。
 *
 * 注意：useAuth() 必须在 mounted 钩子内调用（组件上下文，pinia/router 注入可用）。 若在注册指令时调用，会在 setupRouter 之前创建 auth store，useRoute() 拿不到
 * 路由注入，导致 `injection "Symbol(route location)" not found`。
 */
const authDirective: Directive<HTMLElement, string | string[]> = {
  mounted(el, binding) {
    const { hasAuth } = useAuth();

    if (!hasAuth(binding.value)) {
      el.parentNode?.removeChild(el);
    }
  }
};

export function setupAuthDirectives(app: App) {
  app.directive('auth', authDirective);
}
