<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">签到记录</h2>
            <p class="text-sm text-gray-500 dark:text-gray-400">查看每日签到奖励发放明细</p>
          </div>
          <button @click="loadRecords" :disabled="loading" class="btn btn-secondary">
            <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
          </button>
        </div>
      </template>

      <template #table>
        <DataTable :columns="columns" :data="records" :loading="loading">
          <template #cell-user="{ row }">
            <div class="text-sm">
              <div class="font-medium text-gray-900 dark:text-white">{{ row.user_email || row.username || `用户 #${row.user_id}` }}</div>
              <div class="text-xs text-gray-500 dark:text-gray-400">ID: {{ row.user_id }}</div>
            </div>
          </template>
          <template #cell-reward="{ value }">
            <span class="font-medium text-emerald-600 dark:text-emerald-400">+${{ Number(value || 0).toFixed(0) }}</span>
          </template>
          <template #cell-balance_after="{ value }">
            <span class="text-sm text-gray-700 dark:text-gray-300">{{ Number(value || 0).toFixed(2) }}</span>
          </template>
          <template #cell-checkin_date="{ value }">
            <span class="text-sm text-gray-700 dark:text-gray-300">{{ formatDate(value) }}</span>
          </template>
          <template #cell-created_at="{ value }">
            <span class="text-sm text-gray-500 dark:text-gray-400">{{ formatDateTime(value) }}</span>
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <Pagination
          :page="page"
          :total="total"
          :page-size="limit"
          @update:page="handlePageChange"
          @update:page-size="handlePageSizeChange"
        />
      </template>
    </TablePageLayout>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import Icon from '@/components/icons/Icon.vue'
import { adminAPI } from '@/api/admin'
import type { AdminDailyCheckinRecord } from '@/api/admin/dailyCheckins'

const records = ref<AdminDailyCheckinRecord[]>([])
const loading = ref(false)
const page = ref(1)
const limit = ref(20)
const total = ref(0)

const columns = [
  { key: 'id', label: 'ID', sortable: false },
  { key: 'user', label: '用户', sortable: false },
  { key: 'reward', label: '奖励', sortable: false },
  { key: 'balance_after', label: '签到后余额', sortable: false },
  { key: 'checkin_date', label: '签到日期', sortable: false },
  { key: 'created_at', label: '签到时间', sortable: false },
]

function formatDate(value: string) {
  if (!value) return '-'
  return new Date(value).toLocaleDateString('zh-CN')
}

function formatDateTime(value: string) {
  if (!value) return '-'
  return new Date(value).toLocaleString('zh-CN')
}

async function loadRecords() {
  loading.value = true
  try {
    const response = await adminAPI.dailyCheckins.list({ page: page.value, limit: limit.value })
    const result = response.data
    records.value = result.records || []
    total.value = result.total || 0
    page.value = result.page || page.value
    limit.value = result.limit || limit.value
  } finally {
    loading.value = false
  }
}

function handlePageChange(nextPage: number) {
  page.value = nextPage
  loadRecords()
}

function handlePageSizeChange(nextLimit: number) {
  limit.value = nextLimit
  page.value = 1
  loadRecords()
}

onMounted(loadRecords)
</script>
