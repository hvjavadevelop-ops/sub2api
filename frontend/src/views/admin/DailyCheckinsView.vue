<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">签到管理</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">查看用户每日签到记录，奖励范围可在系统设置中调整。</p>
        </div>
        <router-link to="/admin/settings" class="btn btn-primary">签到奖励配置</router-link>
      </div>

      <div class="card p-4">
        <div class="grid gap-3 md:grid-cols-5">
          <input v-model="filters.search" class="input" placeholder="搜索邮箱/用户名" @keyup.enter="loadRecords" />
          <input v-model="filters.user_id" class="input" placeholder="用户 ID" @keyup.enter="loadRecords" />
          <input v-model="filters.start_date" type="date" class="input" />
          <input v-model="filters.end_date" type="date" class="input" />
          <div class="flex gap-2">
            <button class="btn btn-primary flex-1" @click="handleSearch">查询</button>
            <button class="btn btn-secondary" @click="resetFilters">重置</button>
          </div>
        </div>
      </div>

      <div class="card overflow-hidden">
        <div v-if="loading" class="p-8 text-center text-sm text-gray-500 dark:text-gray-400">加载中...</div>
        <div v-else-if="records.length === 0" class="p-8 text-center text-sm text-gray-500 dark:text-gray-400">暂无签到记录</div>
        <div v-else class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200 dark:divide-dark-600">
            <thead class="bg-gray-50 dark:bg-dark-700/50">
              <tr>
                <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">ID</th>
                <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">用户</th>
                <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">签到日期</th>
                <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">奖励</th>
                <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">签到后余额</th>
                <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">创建时间</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200 bg-white dark:divide-dark-600 dark:bg-dark-800">
              <tr v-for="record in records" :key="record.id" class="hover:bg-gray-50 dark:hover:bg-dark-700/50">
                <td class="whitespace-nowrap px-4 py-3 text-sm text-gray-700 dark:text-gray-300">{{ record.id }}</td>
                <td class="px-4 py-3 text-sm">
                  <div class="font-medium text-gray-900 dark:text-white">{{ record.user_email || record.username || '-' }}</div>
                  <div class="text-xs text-gray-500 dark:text-gray-400">ID: {{ record.user_id }}</div>
                </td>
                <td class="whitespace-nowrap px-4 py-3 text-sm text-gray-700 dark:text-gray-300">{{ formatDate(record.checkin_date) }}</td>
                <td class="whitespace-nowrap px-4 py-3 text-sm font-semibold text-green-600 dark:text-green-400">${{ formatAmount(record.reward) }}</td>
                <td class="whitespace-nowrap px-4 py-3 text-sm text-gray-700 dark:text-gray-300">${{ formatAmount(record.balance_after) }}</td>
                <td class="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{{ formatDateTime(record.created_at) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <div class="flex flex-wrap items-center justify-between gap-3 border-t border-gray-200 px-4 py-3 dark:border-dark-600">
          <div class="text-sm text-gray-500 dark:text-gray-400">共 {{ total }} 条，第 {{ page }} / {{ totalPages }} 页</div>
          <div class="flex gap-2">
            <button class="btn btn-secondary" :disabled="page <= 1 || loading" @click="changePage(page - 1)">上一页</button>
            <button class="btn btn-secondary" :disabled="page >= totalPages || loading" @click="changePage(page + 1)">下一页</button>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import { adminDailyCheckinAPI, type AdminDailyCheckinRecord } from '@/api/admin/dailyCheckin'
import { useAppStore } from '@/stores'

const appStore = useAppStore()
const loading = ref(false)
const records = ref<AdminDailyCheckinRecord[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const filters = reactive({ search: '', user_id: '', start_date: '', end_date: '' })
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize)))

async function loadRecords() {
  loading.value = true
  try {
    const response = await adminDailyCheckinAPI.list({
      page: page.value,
      page_size: pageSize,
      search: filters.search || undefined,
      user_id: filters.user_id || undefined,
      start_date: filters.start_date || undefined,
      end_date: filters.end_date || undefined,
    })
    records.value = response.items || []
    total.value = response.total || 0
  } catch (error) {
    appStore.showError(error instanceof Error ? error.message : '加载签到记录失败')
  } finally {
    loading.value = false
  }
}

function handleSearch() { page.value = 1; loadRecords() }
function resetFilters() { filters.search = ''; filters.user_id = ''; filters.start_date = ''; filters.end_date = ''; page.value = 1; loadRecords() }
function changePage(nextPage: number) { page.value = nextPage; loadRecords() }
function formatAmount(value: number) { return Number(value || 0).toFixed(2) }
function formatDate(value: string) { return value ? new Date(value).toLocaleDateString('zh-CN') : '-' }
function formatDateTime(value: string) { return value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '-' }

onMounted(loadRecords)
</script>
