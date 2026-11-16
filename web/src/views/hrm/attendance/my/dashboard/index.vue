<script setup lang="tsx">
import { computed, onMounted, ref } from 'vue';
import { NAlert, NButton, NCalendar, NCard, NGi, NGrid, NResult, NSpace, NStatistic, NTag } from 'naive-ui';
import { GetMyDashboard } from '@/service/api';
import { $t } from '@/locales';
import { useRouterPush } from '@/hooks/common/router';

defineOptions({
  name: 'AttendanceMyDashboard'
});

const { routerPushByKey } = useRouterPush();

// ---- 数据加载 ----
const loading = ref(false);
const dashboard = ref<Api.Hrm.MyDashboard | null>(null);

async function loadData() {
  loading.value = true;
  const { data, error } = await GetMyDashboard();
  loading.value = false;
  if (error) return;
  dashboard.value = data;
}

onMounted(() => {
  loadData();
});

// ---- 今日卡片 ----
const today = computed(() => dashboard.value?.today);

const settleStatusMeta = computed<
  Record<string, { label: string; type: 'success' | 'warning' | 'error' | 'info' | 'default' }>
>(() => ({
  normal: { label: $t('page.hrm.attendance.myDashboard.calendarStatus.normal'), type: 'success' },
  late: { label: $t('page.hrm.attendance.myDashboard.calendarStatus.late'), type: 'warning' },
  early: { label: $t('page.hrm.attendance.myDashboard.calendarStatus.early'), type: 'warning' },
  missing: { label: $t('page.hrm.attendance.myDashboard.calendarStatus.missing'), type: 'error' },
  rest: { label: $t('page.hrm.attendance.myDashboard.calendarStatus.rest'), type: 'default' },
  no_rule: { label: $t('page.hrm.attendance.myDashboard.calendarStatus.noRule'), type: 'info' }
}));

function renderSettleTag(status: string) {
  if (!status) return null;
  const meta = settleStatusMeta.value[status];
  if (!meta) return null;
  return (
    <NTag type={meta.type} size="small">
      {meta.label}
    </NTag>
  );
}

function punchStatusLabel(status: string) {
  if (status === 'normal') return $t('page.hrm.attendance.myDashboard.calendarStatus.normal');
  if (status === 'field') return $t('page.hrm.attendance.punchRecord.statusField');
  return $t('page.hrm.attendance.myDashboard.calendarStatus.missing');
}

// ---- 本月统计卡 ----
const stats = computed(() => dashboard.value?.stats);

const statItems = computed(() => {
  const s = stats.value;
  if (!s) return [];
  return [
    { key: 'expectedDays', value: `${s.expectedDays}${$t('page.hrm.attendance.myDashboard.daysUnit')}` },
    { key: 'attendedDays', value: `${s.attendedDays}${$t('page.hrm.attendance.myDashboard.daysUnit')}` },
    { key: 'lateTimes', value: `${s.lateTimes}${$t('page.hrm.attendance.myDashboard.timesUnit')}` },
    { key: 'earlyTimes', value: `${s.earlyTimes}${$t('page.hrm.attendance.myDashboard.timesUnit')}` },
    { key: 'missingTimes', value: `${s.missingTimes}${$t('page.hrm.attendance.myDashboard.timesUnit')}` },
    { key: 'absentTimes', value: `${s.absentTimes}${$t('page.hrm.attendance.myDashboard.timesUnit')}` },
    { key: 'overtimeTimes', value: `${s.overtimeTimes}${$t('page.hrm.attendance.myDashboard.timesUnit')}` },
    { key: 'overtimeHours', value: `${s.overtimeHours}${$t('page.hrm.attendance.myDashboard.hoursUnit')}` },
    { key: 'annualLeaveAvailable', value: `${s.annualLeaveAvailable}${$t('page.hrm.attendance.myDashboard.daysUnit')}` }
  ] as const;
});

// ---- 考勤日历（NCalendar + 自定义逐日渲染）----
type CalendarDayMeta = {
  type: 'success' | 'warning' | 'error' | 'info' | 'default';
  label: string;
};

function calendarDayMeta(day: Api.Hrm.DashboardCalendarDay): CalendarDayMeta | null {
  const statusMap: Record<string, { labelKey: string; type: CalendarDayMeta['type'] }> = {
    normal: { labelKey: 'page.hrm.attendance.myDashboard.calendarStatus.normal', type: 'success' },
    late: { labelKey: 'page.hrm.attendance.myDashboard.calendarStatus.late', type: 'warning' },
    early: { labelKey: 'page.hrm.attendance.myDashboard.calendarStatus.early', type: 'warning' },
    missing: { labelKey: 'page.hrm.attendance.myDashboard.calendarStatus.missing', type: 'error' },
    rest: { labelKey: 'page.hrm.attendance.myDashboard.calendarStatus.rest', type: 'default' },
    no_rule: { labelKey: 'page.hrm.attendance.myDashboard.calendarStatus.noRule', type: 'info' }
  };
  if (!day.status) {
    return { label: $t('page.hrm.attendance.myDashboard.calendarStatus.none'), type: 'default' };
  }
  const meta = statusMap[day.status];
  if (!meta) return null;
  return { label: $t(meta.labelKey as App.I18n.I18nKey), type: meta.type };
}

const calendarDays = computed(() => {
  const map = new Map<string, Api.Hrm.DashboardCalendarDay>();
  for (const d of dashboard.value?.calendar ?? []) {
    map.set(d.date, d);
  }
  return map;
});

function dayCellContent(dateTs: number) {
  const date = new Date(dateTs);
  const key = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`;
  const day = calendarDays.value.get(key);
  if (!day) return null;
  const meta = calendarDayMeta(day);
  if (!meta) return null;
  return (
    <div class="flex flex-col items-start gap-2px px-2px text-11px leading-tight">
      <NTag type={meta.type} size="tiny" bordered={false}>
        {meta.label}
      </NTag>
      {day.clockInTime ? <span class="text-gray-500">{day.clockInTime.slice(-8)}</span> : null}
      {day.clockOutTime ? <span class="text-gray-500">{day.clockOutTime.slice(-8)}</span> : null}{' '}
    </div>
  );
}

// ---- 快捷操作 ----
const quickActions = computed(
  () =>
    [
      {
        key: 'leave',
        label: $t('page.hrm.attendance.myDashboard.quick.leave'),
        routeKey: 'hrm_attendance_apply_leave-cancel-list',
        authCode: 'hrm:leave:list'
      },
      {
        key: 'cardReplacement',
        label: $t('page.hrm.attendance.myDashboard.quick.cardReplacement'),
        routeKey: 'hrm_attendance_apply_card-replacement-list',
        authCode: 'hrm:cardreplacement:list'
      },
      {
        key: 'overtime',
        label: $t('page.hrm.attendance.myDashboard.quick.overtime'),
        routeKey: 'hrm_attendance_apply_overtime-bill-list',
        authCode: 'hrm:overtimebill:list'
      },
      {
        key: 'leaveCancel',
        label: $t('page.hrm.attendance.myDashboard.quick.leaveCancel'),
        routeKey: 'hrm_attendance_apply_leave-cancel-list',
        authCode: 'hrm:leave:cancel'
      }
    ] as const
);
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <!-- 未建档引导 -->
    <NAlert v-if="dashboard && !dashboard.registered" type="warning" :show-icon="true">
      <NSpace vertical>
        <span>{{ $t('page.hrm.attendance.myDashboard.registeredGuide') }}</span>
        <NButton size="small" type="primary" @click="routerPushByKey('hrm_employee')">
          {{ $t('page.hrm.attendance.myDashboard.registeredGuideBtn') }}
        </NButton>
      </NSpace>
    </NAlert>

    <!-- 锁定日提示 -->
    <NAlert v-if="dashboard?.lock?.hintText" type="info" :show-icon="true">
      {{ dashboard.lock.hintText }}
    </NAlert>

    <!-- 今日卡片 -->
    <NCard :title="$t('page.hrm.attendance.myDashboard.title')" :bordered="false" size="small" class="card-wrapper">
      <template #header-extra>
        <NButton
          v-auth="'hrm:punch:punch'"
          type="primary"
          size="small"
          @click="routerPushByKey('hrm_attendance_my_attendance-punch')"
        >
          {{ $t('page.hrm.attendance.myDashboard.goPunch') }}
        </NButton>
      </template>
      <template v-if="today">
        <NGrid :x-gap="16" :y-gap="12" responsive="screen" item-responsive cols="2 s:4 m:6">
          <NGi span="1 m:2">
            <NStatistic :label="$t('page.hrm.attendance.myDashboard.todayWork')">
              <span v-if="today.isRestDay" class="text-16px font-bold">
                {{ $t('page.hrm.attendance.myDashboard.todayRest') }}
              </span>
              <span v-else class="text-16px font-bold">
                {{ today.groupName || $t('page.hrm.attendance.myDashboard.noRuleHint') }}
              </span>
            </NStatistic>
          </NGi>
          <NGi span="1">
            <NStatistic :label="$t('page.hrm.attendance.myDashboard.clockInStatus')">
              <div class="flex items-center gap-8px">
                <span>{{ today.clockInTime || '-' }}</span>
                <NTag :type="today.clockInStatus === 'missing' ? 'error' : 'success'" size="tiny">
                  {{ punchStatusLabel(today.clockInStatus) }}
                </NTag>
              </div>
            </NStatistic>
          </NGi>
          <NGi span="1">
            <NStatistic :label="$t('page.hrm.attendance.myDashboard.clockOutStatus')">
              <div class="flex items-center gap-8px">
                <span>{{ today.clockOutTime || '-' }}</span>
                <NTag :type="today.clockOutStatus === 'missing' ? 'error' : 'success'" size="tiny">
                  {{ punchStatusLabel(today.clockOutStatus) }}
                </NTag>
              </div>
            </NStatistic>
          </NGi>
          <NGi span="1">
            <NStatistic :label="$t('page.hrm.attendance.myDashboard.settleStatus')">
              <component :is="renderSettleTag(today.settleStatus)" v-if="renderSettleTag(today.settleStatus)" />
              <span v-else>-</span>
            </NStatistic>
          </NGi>
          <NGi v-if="today.lateMinutes > 0 || today.earlyMinutes > 0" span="1 m:2">
            <NStatistic>
              <template #label>
                <span class="sr-only">deviation</span>
              </template>
              <NSpace size="small">
                <NTag v-if="today.lateMinutes > 0" type="warning" size="small">
                  {{ $t('page.hrm.attendance.myDashboard.lateMinutes', { n: today.lateMinutes }) }}
                </NTag>
                <NTag v-if="today.earlyMinutes > 0" type="warning" size="small">
                  {{ $t('page.hrm.attendance.myDashboard.earlyMinutes', { n: today.earlyMinutes }) }}
                </NTag>
              </NSpace>
            </NStatistic>
          </NGi>
        </NGrid>
      </template>
    </NCard>

    <!-- 本月统计 -->
    <NCard
      :title="$t('page.hrm.attendance.myDashboard.monthStats')"
      :bordered="false"
      size="small"
      class="card-wrapper"
    >
      <NGrid :x-gap="16" :y-gap="12" responsive="screen" item-responsive cols="3 s:5 m:9">
        <NGi v-for="item in statItems" :key="item.key" span="1">
          <NStatistic :label="$t(`page.hrm.attendance.myDashboard.${item.key}` as App.I18n.I18nKey)">
            <span class="text-16px font-bold">{{ item.value }}</span>
          </NStatistic>
        </NGi>
      </NGrid>
    </NCard>

    <!-- 快捷操作 -->
    <NCard
      :title="$t('page.hrm.attendance.myDashboard.quickActions')"
      :bordered="false"
      size="small"
      class="card-wrapper"
    >
      <NSpace>
        <NButton
          v-for="action in quickActions"
          :key="action.key"
          v-auth="action.authCode"
          size="small"
          @click="routerPushByKey(action.routeKey)"
        >
          {{ action.label }}
        </NButton>
      </NSpace>
    </NCard>

    <!-- 考勤日历 -->
    <NCard :title="$t('page.hrm.attendance.myDashboard.calendar')" :bordered="false" size="small" class="card-wrapper">
      <NCalendar v-if="!loading && dashboard?.registered">
        <template #header="{ year, month }">
          <span>{{ year }}-{{ String(month).padStart(2, '0') }}</span>
        </template>
        <template #default="{ ts }">
          <component :is="dayCellContent(ts)" v-if="dayCellContent(ts)" />
        </template>
      </NCalendar>
      <NResult v-else-if="loading" status="info" :description="$t('page.hrm.attendance.myDashboard.title')" />
    </NCard>
  </div>
</template>
