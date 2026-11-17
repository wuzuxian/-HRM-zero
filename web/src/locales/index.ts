import type { App } from 'vue';
import { createI18n } from 'vue-i18n';
import { localStg } from '@/utils/storage';
import messages from './locale';

const i18n = createI18n({
  locale: localStg.get('lang') || 'zh-CN',
  fallbackLocale: 'en',
  messages,
  legacy: false
});

/**
 * Setup plugin i18n
 *
 * @param app
 */
export function setupI18n(app: App) {
  app.use(i18n);
}

const rawT = i18n.global.t.bind(i18n.global);

/**
 * 枚举渲染守卫：全仓大量列表列用 `$t(enumRecord[row.x])` 渲染枚举 tag， 脏数据（空串/非法枚举值）会让索引结果为 undefined，vue-i18n 直接抛 "Invalid arguments"
 * 且中断组件渲染，表格 keyed patch 崩溃（__vnode null）。 这里在唯一消费口统一兜底：非字符串 key 返回空串，合法 key 走正常翻译。
 */
export const $t = ((key: unknown, ...args: unknown[]) => {
  if (typeof key !== 'string' || key.length === 0) {
    return '';
  }
  return (rawT as (...a: unknown[]) => string)(key, ...args);
}) as App.I18n.$T;

export function setLocale(locale: App.I18n.LangType) {
  i18n.global.locale.value = locale;
}
