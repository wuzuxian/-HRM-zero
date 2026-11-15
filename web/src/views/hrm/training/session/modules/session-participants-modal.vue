<script setup lang="ts">
import { ref, watch } from 'vue';
import { useLoading } from '@sa/hooks';
import { GetSessionParticipantList, GetUserList, SetSessionParticipants } from '@/service/api';
import { $t } from '@/locales';

defineOptions({
  name: 'SessionParticipantsModal'
});

interface Props {
  /** session uuid */
  sessionUuid: string;
  /** session name for the modal title */
  sessionName: string;
}

const props = defineProps<Props>();

interface Emits {
  (e: 'submitted'): void;
}

const emit = defineEmits<Emits>();

const visible = defineModel<boolean>('visible', {
  default: false
});

const { loading: confirmLoading, startLoading: confirmStartLoading, endLoading: confirmEndLoading } = useLoading();

/** selected 学员 uuids (full replace on save) */
const selectedUserUuids = ref<string[]>([]);

// current bindings (回显)
async function loadParticipants() {
  if (!props.sessionUuid) return;
  const { data, error } = await GetSessionParticipantList(props.sessionUuid);
  if (!error && data) {
    selectedUserUuids.value = (data.records ?? []).map(p => p.userUuid);
  }
}

// ---- 学员 picker (manage_user list, paginated remote search) ----
interface UserOption {
  label: string;
  value: string;
}

const userOptions = ref<UserOption[]>([]);
const userLoading = ref(false);

async function searchUsers(query: string) {
  userLoading.value = true;
  const { data, error } = await GetUserList({ current: 1, size: 50, username: query || undefined });
  userLoading.value = false;
  if (!error && data) {
    userOptions.value = (data.records ?? []).map(u => ({
      label: u.nickName ? `${u.nickName}（${u.username}）` : u.username,
      value: u.uuid
    }));
  }
}

function handleInit() {
  loadParticipants();
  searchUsers('');
}

async function handleSubmit() {
  confirmStartLoading();
  const { error } = await SetSessionParticipants({
    sessionUuid: props.sessionUuid,
    userUuids: selectedUserUuids.value
  });
  confirmEndLoading();
  if (error) return;
  window.$message?.success($t('common.updateSuccess'));
  visible.value = false;
  emit('submitted');
}

watch(visible, val => {
  if (val) {
    handleInit();
  }
});
</script>

<template>
  <NModal
    v-model:show="visible"
    :title="`${$t('page.hrm.training.session.participants')} - ${props.sessionName}`"
    preset="card"
    class="w-560px"
  >
    <NSelect
      v-model:value="selectedUserUuids"
      multiple
      filterable
      remote
      clearable
      :options="userOptions"
      :loading="userLoading"
      :placeholder="$t('page.hrm.training.session.form.participantsPlaceholder')"
      @search="searchUsers"
    />
    <div class="mt-12px text-12px text-gray-500">{{ $t('page.hrm.training.session.participantsTip') }}</div>
    <template #footer>
      <NSpace justify="end" :size="16">
        <NButton @click="visible = false">{{ $t('common.cancel') }}</NButton>
        <NButton type="primary" :loading="confirmLoading" @click="handleSubmit">{{ $t('common.save') }}</NButton>
      </NSpace>
    </template>
  </NModal>
</template>

<style scoped></style>
