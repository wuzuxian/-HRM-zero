<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import { NButton, NCard, NResult, NSpin } from 'naive-ui';
import { DoSignin } from '@/service/api';
import { $t } from '@/locales';

// 学员扫码签到页（HRM-29）：课堂动态码 URL 落地页。
// jwt+token 双验：学员端已登录（manage_user JWT），token 来自二维码 URL 参数。
// 幂等：重复扫码返回首次签到时间。
const route = useRoute();

const token = computed(() => String(route.query.token ?? ''));
const state = ref<'idle' | 'loading' | 'success' | 'error'>('idle');
const resultName = ref('');
const resultTime = ref('');
const errorMsg = ref('');

onMounted(() => {
  if (!token.value) {
    state.value = 'error';
    errorMsg.value = $t('page.hrm.training.signin.scan.noToken');
    return;
  }
  state.value = 'loading';
  DoSignin(token.value).then(({ data, error }) => {
    if (error || !data) {
      // 后端错误信息已由全局 onError toast 展示；本页只展示结果态
      state.value = 'error';
      return;
    }
    resultName.value = data.sessionName;
    resultTime.value = data.signInTime;
    state.value = 'success';
  });
});
</script>

<template>
  <div class="min-h-400px flex-col-center">
    <NCard :title="$t('page.hrm.training.signin.scan.title')" :bordered="false" size="small" class="w-460px card-wrapper">
      <NSpin v-if="state === 'loading'" class="w-full py-24px" :show="true" />
      <NResult v-else-if="state === 'success'" status="success" :title="$t('page.hrm.training.signin.scan.successTitle')">
        <template #default>
          <div class="flex-col gap-4px">
            <span>{{ resultName }}</span>
            <span class="font-medium">{{ resultTime }}</span>
          </div>
        </template>
      </NResult>
      <template v-else-if="state === 'error'">
        <NResult status="error" :title="errorMsg || $t('page.hrm.training.signin.scan.errorTitle')" />
      </template>
      <div v-if="state === 'success'" class="flex justify-center mt-8px">
        <NButton type="primary" ghost @click="$router.back()">{{ $t('common.back') }}</NButton>
      </div>
    </NCard>
  </div>
</template>

<style scoped></style>
