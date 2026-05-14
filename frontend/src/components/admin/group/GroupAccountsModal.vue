<template>
  <BaseDialog
    :show="show"
    :title="modalTitle"
    width="extra-wide"
    @close="handleClose"
  >
    <div v-if="group" class="space-y-5">
      <div
        class="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-gray-200 bg-gray-50 px-4 py-3 dark:border-dark-600 dark:bg-dark-700/50"
      >
        <div class="min-w-0">
          <div class="flex flex-wrap items-center gap-2">
            <span class="font-medium text-gray-900 dark:text-white">
              {{ group.name }}
            </span>
            <span
              class="inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium"
              :class="platformBadgeClass(group.platform)"
            >
              <PlatformIcon :platform="group.platform" size="xs" />
              {{ platformLabel(group.platform) }}
            </span>
            <span
              :class="[
                'badge',
                group.status === 'active' ? 'badge-success' : 'badge-danger',
              ]"
            >
              {{ group.status === "active" ? "启用" : "停用" }}
            </span>
          </div>
          <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
            这里编辑的是账号和当前分组的绑定关系，不会修改账号密钥或代理配置。
          </p>
        </div>
        <div class="flex items-center gap-4 text-xs">
          <div>
            <span class="text-gray-500 dark:text-gray-400">总数</span>
            <span class="ml-1 font-semibold text-gray-900 dark:text-white">
              {{ group.account_count || 0 }}
            </span>
          </div>
          <div>
            <span class="text-gray-500 dark:text-gray-400">可用</span>
            <span class="ml-1 font-semibold text-emerald-600 dark:text-emerald-400">
              {{
                (group.active_account_count || 0) -
                (group.rate_limited_account_count || 0)
              }}
            </span>
          </div>
          <div>
            <span class="text-gray-500 dark:text-gray-400">限流</span>
            <span class="ml-1 font-semibold text-amber-600 dark:text-amber-400">
              {{ group.rate_limited_account_count || 0 }}
            </span>
          </div>
        </div>
      </div>

      <section class="space-y-3">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h4 class="text-sm font-medium text-gray-900 dark:text-white">
              当前分组账号
            </h4>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              点击移除只会解除这个分组绑定，账号本身仍然保留。
            </p>
          </div>
          <div class="flex w-full flex-wrap items-center gap-2 sm:w-auto">
            <div class="relative w-full sm:w-72">
              <Icon
                name="search"
                size="sm"
                class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 dark:text-gray-500"
              />
              <input
                v-model="boundSearch"
                type="text"
                class="input pl-9"
                placeholder="搜索当前分组账号"
                @input="handleBoundSearchInput"
              />
            </div>
            <button
              type="button"
              class="btn btn-secondary"
              :disabled="boundLoading"
              @click="loadBoundAccounts"
            >
              <Icon
                name="refresh"
                size="sm"
                :class="boundLoading ? 'animate-spin' : ''"
              />
            </button>
          </div>
        </div>

        <div
          class="overflow-hidden rounded-lg border border-gray-200 dark:border-dark-600"
        >
          <div
            v-if="boundLoading"
            class="flex items-center justify-center py-10 text-sm text-gray-500 dark:text-gray-400"
          >
            加载中...
          </div>
          <div
            v-else-if="boundAccounts.length === 0"
            class="flex flex-col items-center justify-center gap-2 py-10 text-sm text-gray-500 dark:text-gray-400"
          >
            <Icon name="inbox" size="lg" />
            <span>这个分组还没有绑定账号</span>
          </div>
          <div v-else class="overflow-x-auto">
            <table class="w-full min-w-[820px] divide-y divide-gray-100 dark:divide-dark-700">
              <thead class="bg-gray-50 dark:bg-dark-700/60">
                <tr>
                  <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400">
                    账号
                  </th>
                  <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400">
                    类型
                  </th>
                  <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400">
                    状态
                  </th>
                  <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400">
                    调度
                  </th>
                  <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400">
                    错误
                  </th>
                  <th class="px-4 py-2 text-right text-xs font-medium text-gray-500 dark:text-gray-400">
                    操作
                  </th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 bg-white dark:divide-dark-700 dark:bg-dark-800">
                <tr
                  v-for="account in boundAccounts"
                  :key="account.id"
                  class="hover:bg-gray-50 dark:hover:bg-dark-700/50"
                >
                  <td class="px-4 py-2">
                    <div class="max-w-xs truncate text-sm font-medium text-gray-900 dark:text-white">
                      {{ account.name }}
                    </div>
                    <div class="text-xs text-gray-400">#{{ account.id }}</div>
                  </td>
                  <td class="px-4 py-2 text-xs text-gray-600 dark:text-gray-300">
                    {{ formatAccountType(account.type) }}
                  </td>
                  <td class="px-4 py-2">
                    <span :class="['badge', accountStatusClass(account.status)]">
                      {{ accountStatusLabel(account.status) }}
                    </span>
                  </td>
                  <td class="px-4 py-2">
                    <span
                      :class="[
                        'badge',
                        account.schedulable ? 'badge-success' : 'badge-danger',
                      ]"
                    >
                      {{ account.schedulable ? "可调度" : "不可调度" }}
                    </span>
                  </td>
                  <td class="px-4 py-2">
                    <div class="max-w-xs truncate text-xs text-gray-500 dark:text-gray-400">
                      {{
                        account.temp_unschedulable_reason ||
                        account.error_message ||
                        "-"
                      }}
                    </div>
                  </td>
                  <td class="px-4 py-2 text-right">
                    <button
                      type="button"
                      class="inline-flex items-center gap-1.5 rounded-lg px-2.5 py-1.5 text-xs font-medium text-red-600 transition-colors hover:bg-red-50 disabled:cursor-not-allowed disabled:opacity-60 dark:text-red-400 dark:hover:bg-red-900/20"
                      :disabled="isUpdating(account.id)"
                      @click="removeAccount(account)"
                    >
                      <Icon name="x" size="xs" />
                      {{ isUpdating(account.id) ? "处理中" : "移除" }}
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <div
          v-if="boundPagination.pages > 1"
          class="flex flex-wrap items-center justify-between gap-3 text-xs text-gray-500 dark:text-gray-400"
        >
          <span>
            第 {{ boundPagination.page }} / {{ boundPagination.pages }} 页，共
            {{ boundPagination.total }} 个账号
          </span>
          <div class="flex items-center gap-2">
            <button
              type="button"
              class="btn btn-secondary"
              :disabled="boundPagination.page <= 1 || boundLoading"
              @click="changeBoundPage(boundPagination.page - 1)"
            >
              上一页
            </button>
            <button
              type="button"
              class="btn btn-secondary"
              :disabled="
                boundPagination.page >= boundPagination.pages || boundLoading
              "
              @click="changeBoundPage(boundPagination.page + 1)"
            >
              下一页
            </button>
          </div>
        </div>
      </section>

      <section
        class="space-y-3 rounded-lg border border-dashed border-gray-300 p-4 dark:border-dark-600"
      >
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div>
            <h4 class="text-sm font-medium text-gray-900 dark:text-white">
              添加账号
            </h4>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              只搜索 {{ platformLabel(group.platform) }} 平台账号，避免跨平台误绑。
            </p>
          </div>
          <button
            type="button"
            class="btn btn-secondary"
            :disabled="candidateLoading"
            @click="searchCandidates"
          >
            <Icon
              name="search"
              size="sm"
              :class="candidateLoading ? 'animate-spin' : ''"
            />
            搜索
          </button>
        </div>
        <input
          v-model="candidateSearch"
          type="text"
          class="input"
          placeholder="输入账号名称或 ID 搜索可添加账号"
          @input="handleCandidateSearchInput"
          @keydown.enter.prevent="searchCandidates"
        />

        <div
          v-if="candidateLoading"
          class="py-4 text-sm text-gray-500 dark:text-gray-400"
        >
          搜索中...
        </div>
        <div
          v-else-if="candidateSearched && candidateAccounts.length === 0"
          class="py-4 text-sm text-gray-500 dark:text-gray-400"
        >
          没有找到可添加账号
        </div>
        <div v-else-if="candidateAccounts.length > 0" class="grid gap-2 md:grid-cols-2">
          <div
            v-for="account in candidateAccounts"
            :key="account.id"
            class="flex items-center justify-between gap-3 rounded-lg border border-gray-200 bg-white px-3 py-2 dark:border-dark-600 dark:bg-dark-800"
          >
            <div class="min-w-0">
              <div class="truncate text-sm font-medium text-gray-900 dark:text-white">
                {{ account.name }}
              </div>
              <div class="text-xs text-gray-500 dark:text-gray-400">
                #{{ account.id }} · {{ formatAccountType(account.type) }} ·
                {{ accountStatusLabel(account.status) }}
              </div>
            </div>
            <button
              type="button"
              class="inline-flex shrink-0 items-center gap-1.5 rounded-lg px-2.5 py-1.5 text-xs font-medium text-primary-600 transition-colors hover:bg-primary-50 disabled:cursor-not-allowed disabled:opacity-60 dark:text-primary-400 dark:hover:bg-primary-900/20"
              :disabled="isUpdating(account.id)"
              @click="addAccount(account)"
            >
              <Icon name="plus" size="xs" />
              {{ isUpdating(account.id) ? "处理中" : "添加" }}
            </button>
          </div>
        </div>
      </section>
    </div>

    <template #footer>
      <div class="flex justify-end gap-3 pt-4">
        <button type="button" class="btn btn-secondary" @click="handleClose">
          关闭
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from "vue";
import { adminAPI } from "@/api/admin";
import { useAppStore } from "@/stores/app";
import type { Account, AdminGroup, GroupPlatform } from "@/types";
import BaseDialog from "@/components/common/BaseDialog.vue";
import Icon from "@/components/icons/Icon.vue";
import PlatformIcon from "@/components/common/PlatformIcon.vue";

const props = defineProps<{
  show: boolean;
  group: AdminGroup | null;
}>();

const emit = defineEmits<{
  (e: "close"): void;
  (e: "success"): void;
}>();

const appStore = useAppStore();

const boundAccounts = ref<Account[]>([]);
const candidateResults = ref<Account[]>([]);
const boundLoading = ref(false);
const candidateLoading = ref(false);
const candidateSearched = ref(false);
const boundSearch = ref("");
const candidateSearch = ref("");
const updatingAccountIds = ref<Set<number>>(new Set());
const boundPagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
  pages: 0,
});

let boundSearchTimer: ReturnType<typeof setTimeout> | null = null;
let candidateSearchTimer: ReturnType<typeof setTimeout> | null = null;

const modalTitle = computed(() =>
  props.group ? `管理分组账号：${props.group.name}` : "管理分组账号",
);

const boundAccountIdSet = computed(
  () => new Set(boundAccounts.value.map((account) => account.id)),
);

const candidateAccounts = computed(() => {
  const groupID = props.group?.id;
  if (!groupID) return [];
  return candidateResults.value.filter((account) => {
    if (account.group_ids?.includes(groupID)) return false;
    return !boundAccountIdSet.value.has(account.id);
  });
});

const platformLabel = (platform: GroupPlatform | string) => {
  switch (platform) {
    case "anthropic":
      return "Anthropic";
    case "openai":
      return "OpenAI";
    case "gemini":
      return "Gemini";
    case "antigravity":
      return "Antigravity";
    default:
      return platform;
  }
};

const platformBadgeClass = (platform: string) => {
  if (platform === "anthropic") {
    return "bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400";
  }
  if (platform === "openai") {
    return "bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400";
  }
  if (platform === "antigravity") {
    return "bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400";
  }
  return "bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400";
};

const formatAccountType = (type: string) => {
  switch (type) {
    case "oauth":
      return "OAuth";
    case "setup-token":
      return "Setup Token";
    case "apikey":
      return "API Key";
    case "bedrock":
      return "Bedrock";
    default:
      return type || "-";
  }
};

const accountStatusLabel = (status: string) => {
  switch (status) {
    case "active":
      return "启用";
    case "inactive":
      return "停用";
    case "error":
      return "异常";
    default:
      return status || "-";
  }
};

const accountStatusClass = (status: string) => {
  if (status === "active") return "badge-success";
  if (status === "error") return "badge-danger";
  return "badge-gray";
};

const isUpdating = (accountID: number) => updatingAccountIds.value.has(accountID);

const setUpdating = (accountID: number, updating: boolean) => {
  const next = new Set(updatingAccountIds.value);
  if (updating) {
    next.add(accountID);
  } else {
    next.delete(accountID);
  }
  updatingAccountIds.value = next;
};

const loadBoundAccounts = async () => {
  if (!props.group) return;
  boundLoading.value = true;
  try {
    const response = await adminAPI.accounts.list(
      boundPagination.page,
      boundPagination.page_size,
      {
        group: String(props.group.id),
        search: boundSearch.value.trim() || undefined,
        sort_by: "name",
        sort_order: "asc",
      },
    );
    boundAccounts.value = response.items;
    boundPagination.total = response.total;
    boundPagination.pages = response.pages;
  } catch (error: any) {
    appStore.showError(error.message || "加载分组账号失败");
  } finally {
    boundLoading.value = false;
  }
};

const searchCandidates = async () => {
  if (!props.group) return;
  candidateLoading.value = true;
  candidateSearched.value = true;
  try {
    const response = await adminAPI.accounts.list(1, 50, {
      platform: props.group.platform,
      search: candidateSearch.value.trim() || undefined,
      sort_by: "name",
      sort_order: "asc",
    });
    candidateResults.value = response.items;
  } catch (error: any) {
    appStore.showError(error.message || "搜索账号失败");
    candidateResults.value = [];
  } finally {
    candidateLoading.value = false;
  }
};

const handleBoundSearchInput = () => {
  if (boundSearchTimer) {
    clearTimeout(boundSearchTimer);
  }
  boundSearchTimer = setTimeout(() => {
    boundPagination.page = 1;
    loadBoundAccounts();
  }, 300);
};

const handleCandidateSearchInput = () => {
  if (candidateSearchTimer) {
    clearTimeout(candidateSearchTimer);
  }
  candidateSearchTimer = setTimeout(() => {
    searchCandidates();
  }, 300);
};

const changeBoundPage = (page: number) => {
  boundPagination.page = page;
  loadBoundAccounts();
};

const updateAccountGroupIDs = async (
  account: Account,
  nextGroupIDs: (currentGroupIDs: number[]) => number[],
  confirmedMixedChannelRisk = false,
) => {
  if (!props.group) return false;
  setUpdating(account.id, true);
  try {
    const fresh = await adminAPI.accounts.getById(account.id);
    const currentGroupIDs = Array.isArray(fresh.group_ids) ? fresh.group_ids : [];
    const groupIDs = Array.from(new Set(nextGroupIDs(currentGroupIDs)));
    await adminAPI.accounts.update(fresh.id, {
      group_ids: groupIDs,
      ...(confirmedMixedChannelRisk ? { confirm_mixed_channel_risk: true } : {}),
    });
    await loadBoundAccounts();
    if (candidateSearched.value) {
      await searchCandidates();
    }
    emit("success");
    return true;
  } catch (error: any) {
    if (error.status === 409 && error.error === "mixed_channel_warning") {
      const ok = window.confirm(
        `${error.message || "检测到 Anthropic/Antigravity 混合渠道风险"}\n\n仍然继续绑定吗？`,
      );
      if (ok) {
        return await updateAccountGroupIDs(account, nextGroupIDs, true);
      }
      return false;
    }
    appStore.showError(error.message || "更新账号分组失败");
    return false;
  } finally {
    setUpdating(account.id, false);
  }
};

const addAccount = async (account: Account) => {
  if (!props.group) return;
  const updated = await updateAccountGroupIDs(account, (currentGroupIDs) => [
    ...currentGroupIDs,
    props.group!.id,
  ]);
  if (updated) {
    appStore.showSuccess("账号已加入当前分组");
  }
};

const removeAccount = async (account: Account) => {
  if (!props.group) return;
  const updated = await updateAccountGroupIDs(account, (currentGroupIDs) =>
    currentGroupIDs.filter((id) => id !== props.group!.id),
  );
  if (updated) {
    appStore.showSuccess("账号已从当前分组移除");
  }
};

const resetState = () => {
  boundAccounts.value = [];
  candidateResults.value = [];
  candidateSearched.value = false;
  boundSearch.value = "";
  candidateSearch.value = "";
  boundPagination.page = 1;
  boundPagination.total = 0;
  boundPagination.pages = 0;
  updatingAccountIds.value = new Set();
  if (boundSearchTimer) {
    clearTimeout(boundSearchTimer);
    boundSearchTimer = null;
  }
  if (candidateSearchTimer) {
    clearTimeout(candidateSearchTimer);
    candidateSearchTimer = null;
  }
};

const handleClose = () => {
  resetState();
  emit("close");
};

watch(
  () => [props.show, props.group?.id],
  ([show]) => {
    if (show && props.group) {
      resetState();
      loadBoundAccounts();
    }
  },
);
</script>
