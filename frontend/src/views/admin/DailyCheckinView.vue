<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">签到管理</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">查看用户每日签到记录，并在本页直接调整签到奖励范围。</p>
        </div>
      </div>

      <div class="grid gap-4 lg:grid-cols-3">
        <div class="card p-4 lg:col-span-2">
          <div class="mb-3 flex items-center justify-between">
            <div>
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">记录筛选</h2>
              <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">可按邮箱、用户名、用户ID、签到日期范围查询。</p>
            </div>
          </div>
          <div class="grid gap-3 md:grid-cols-5">
            <input
              v-model.trim="filters.search"
              type="text"
              class="input md:col-span-2"
              placeholder="搜索邮箱 / 用户名 / 用户ID"
              @keyup.enter="applyFilters"
            />
            <input v-model="filters.start_date" type="date" class="input" />
            <input v-model="filters.end_date" type="date" class="input" />
            <div class="flex gap-2">
              <button type="button" class="btn btn-primary flex-1" @click="applyFilters">查询</button>
              <button type="button" class="btn btn-ghost" @click="resetFilters">重置</button>
            </div>
          </div>
        </div>

        <div class="card p-4">
          <div class="mb-3">
            <h2 class="text-base font-semibold text-gray-900 dark:text-white">签到奖励配置</h2>
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">用户每日签到随机获得该范围内的 $ 余额。</p>
          </div>
          <div v-if="settingsLoading" class="py-8 text-center text-sm text-gray-500 dark:text-gray-400">配置加载中...</div>
          <div v-else class="space-y-3">
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">最小奖励 $</label>
                <input v-model.number="rewardForm.min" type="number" min="0" step="1" class="input" placeholder="10" />
              </div>
              <div>
                <label class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">最大奖励 $</label>
                <input v-model.number="rewardForm.max" type="number" min="0" step="1" class="input" placeholder="29" />
              </div>
            </div>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              当前生效：${{ formatAmount(currentReward.min) }} - ${{ formatAmount(currentReward.max) }}
            </p>
            <button type="button" class="btn btn-primary w-full" :disabled="savingSettings" @click="saveRewardSettings">
              {{ savingSettings ? '保存中...' : '保存签到配置' }}
            </button>
          </div>
        </div>
      </div>

      <div class="card overflow-hidden">
        <div v-if="loading" class="flex items-center justify-center py-12">
          <div class="h-8 w-8 animate-spin rounded-full border-b-2 border-primary-600"></div>
        </div>

        <div v-else-if="records.length === 0" class="py-12 text-center text-sm text-gray-500 dark:text-gray-400">
          暂无签到记录
        </div>

        <div v-else class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200 dark:divide-dark-700">
            <thead class="bg-gray-50 dark:bg-dark-700">
              <tr>
                <th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-300">记录ID</th>
                <th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-300">用户</th>
                <th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-300">签到日期</th>
                <th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-300">奖励</th>
                <th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-300">签到后余额</th>
                <th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-300">创建时间</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200 bg-white dark:divide-dark-700 dark:bg-dark-800">
              <tr v-for="record in records" :key="record.id" class="hover:bg-gray-50 dark:hover:bg-dark-700/50">
                <td class="whitespace-nowrap px-6 py-4 text-sm text-gray-500 dark:text-gray-400">#{{ record.id }}</td>
                <td class="whitespace-nowrap px-6 py-4">
                  <div class="text-sm font-medium text-gray-900 dark:text-white">
                    {{ record.username || record.user_email || `用户 ${record.user_id}` }}
                  </div>
                  <div class="text-xs text-gray-500 dark:text-gray-400">
                    ID: {{ record.user_id }}<span v-if="record.user_email"> · {{ record.user_email }}</span>
                  </div>
                </td>
                <td class="whitespace-nowrap px-6 py-4 text-sm text-gray-700 dark:text-gray-300">{{ formatCheckinDate(record.checkin_date) }}</td>
                <td class="whitespace-nowrap px-6 py-4 text-sm font-semibold text-emerald-600 dark:text-emerald-400">+${{ formatAmount(record.reward) }}</td>
                <td class="whitespace-nowrap px-6 py-4 text-sm text-gray-700 dark:text-gray-300">${{ formatAmount(record.balance_after) }}</td>
                <td class="whitespace-nowrap px-6 py-4 text-sm text-gray-500 dark:text-gray-400">{{ formatDateTime(record.created_at) }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <Pagination
          v-if="pagination.total > 0"
          :page="pagination.page"
          :total="pagination.total"
          :page-size="pagination.page_size"
          @update:page="handlePageChange"
          @update:pageSize="handlePageSizeChange"
        />
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import Pagination from '@/components/common/Pagination.vue'
import { adminDailyCheckinAPI, type AdminDailyCheckinRecord } from '@/api/admin/dailyCheckin'
import { settingsAPI } from '@/api/admin/settings'
import { useAppStore } from '@/stores'
import { formatDateTime } from '@/utils/format'

const appStore = useAppStore()
const loading = ref(false)
const settingsLoading = ref(false)
const savingSettings = ref(false)
const records = ref<AdminDailyCheckinRecord[]>([])

const filters = reactive({
  search: '',
  start_date: '',
  end_date: '',
})

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
})

const rewardForm = reactive({
  min: 10,
  max: 29,
})

const currentReward = reactive({
  min: 10,
  max: 29,
})

function formatAmount(value: number | null | undefined): string {
  const n = Number(value ?? 0)
  return Number.isFinite(n) ? n.toFixed(2).replace(/\.00$/, '') : '0'
}

function formatCheckinDate(value: string): string {
  if (!value) return '-'
  return value.slice(0, 10)
}

async function loadRewardSettings() {
  settingsLoading.value = true
  try {
    const settings = await settingsAPI.getSettings()
    const min = Number(settings.daily_checkin_reward_min ?? 10)
    const max = Number(settings.daily_checkin_reward_max ?? 29)
    rewardForm.min = min
    rewardForm.max = max
    currentReward.min = min
    currentReward.max = max
  } catch (error) {
    appStore.showError(error instanceof Error ? error.message : '加载签到配置失败')
  } finally {
    settingsLoading.value = false
  }
}

async function saveRewardSettings() {
  const min = Math.floor(Number(rewardForm.min) || 0)
  const max = Math.floor(Number(rewardForm.max) || 0)
  if (min < 0 || max < 0 || min > max) {
    appStore.showError('签到奖励配置不合法：最小值和最大值必须大于等于 0，且最小值不能大于最大值')
    return
  }

  savingSettings.value = true
  try {
    const settings = await settingsAPI.updateSettings({
      daily_checkin_reward_min: min,
      daily_checkin_reward_max: max,
    })
    currentReward.min = Number(settings.daily_checkin_reward_min ?? min)
    currentReward.max = Number(settings.daily_checkin_reward_max ?? max)
    rewardForm.min = currentReward.min
    rewardForm.max = currentReward.max
    appStore.showSuccess('签到配置已保存')
  } catch (error) {
    appStore.showError(error instanceof Error ? error.message : '保存签到配置失败')
  } finally {
    savingSettings.value = false
  }
}

async function loadRecords() {
  loading.value = true
  try {
    const res = await adminDailyCheckinAPI.list({
      page: pagination.page,
      page_size: pagination.page_size,
      search: filters.search || undefined,
      start_date: filters.start_date || undefined,
      end_date: filters.end_date || undefined,
    })
    records.value = res.items || []
    pagination.total = res.total || 0
  } catch (error) {
    appStore.showError(error instanceof Error ? error.message : '加载签到记录失败')
  } finally {
    loading.value = false
  }
}

function applyFilters() {
  pagination.page = 1
  loadRecords()
}

function resetFilters() {
  filters.search = ''
  filters.start_date = ''
  filters.end_date = ''
  pagination.page = 1
  loadRecords()
}

function handlePageChange(page: number) {
  pagination.page = page
  loadRecords()
}

function handlePageSizeChange(pageSize: number) {
  pagination.page_size = pageSize
  pagination.page = 1
  loadRecords()
}

onMounted(() => {
  loadRecords()
  loadRewardSettings()
})
</script>
