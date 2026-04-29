<template>
  <BaseDialog :show="show" title="限时权益" width="wide" @close="handleClose">
    <div v-if="user" class="space-y-5">
      <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-700">
        <div class="font-medium text-gray-900 dark:text-white">{{ user.email }}</div>
        <div class="mt-1 text-sm text-gray-500 dark:text-gray-400">
          当前余额 ${{ formatAmount(user.balance) }} · 当前并发 {{ user.concurrency }}
        </div>
      </div>

      <form id="timed-grant-form" class="grid gap-4 rounded-xl border border-gray-200 p-4 dark:border-dark-600" @submit.prevent="handleSubmit">
        <div class="flex flex-wrap items-end gap-3">
          <div class="min-w-36 flex-1">
            <label class="input-label">权益类型</label>
            <select v-model="form.grant_type" class="input">
              <option value="balance">限时余额</option>
              <option value="concurrency">限时并发</option>
            </select>
          </div>
          <div class="min-w-32 flex-1">
            <label class="input-label">数量</label>
            <input v-model.number="form.amount" :step="form.grant_type === 'balance' ? '0.01' : '1'" min="0" type="number" required class="input" />
          </div>
          <div class="min-w-36 flex-1">
            <label class="input-label">有效期</label>
            <select v-model.number="durationPreset" class="input" @change="applyDurationPreset">
              <option :value="3600">1 小时</option>
              <option :value="86400">1 天</option>
              <option :value="604800">7 天</option>
              <option :value="2592000">30 天</option>
              <option :value="0">自定义</option>
            </select>
          </div>
          <div v-if="durationPreset === 0" class="min-w-32 flex-1">
            <label class="input-label">秒数</label>
            <input v-model.number="form.duration_seconds" min="1" type="number" required class="input" />
          </div>
        </div>
        <div>
          <label class="input-label">备注</label>
          <textarea v-model="form.notes" rows="2" class="input" placeholder="例如：活动赠送，首次使用后开始倒计时"></textarea>
        </div>
        <div class="rounded-lg bg-amber-50 px-3 py-2 text-sm text-amber-700 dark:bg-amber-900/20 dark:text-amber-300">
          创建后先显示为“待激活”，用户下一次调用 API 时自动加到账户，并从那一刻开始倒计时；到期后系统自动扣回，扣减记录会进入余额/并发历史。
        </div>
      </form>

      <div>
        <div class="mb-2 flex items-center justify-between">
          <h3 class="text-sm font-medium text-gray-900 dark:text-white">历史记录</h3>
          <button class="btn btn-secondary px-2 py-1 text-xs" :disabled="loading" @click="loadGrants">刷新</button>
        </div>
        <div class="max-h-80 overflow-auto rounded-xl border border-gray-200 dark:border-dark-600">
          <table class="min-w-full divide-y divide-gray-200 text-sm dark:divide-dark-600">
            <thead class="bg-gray-50 text-left text-xs uppercase text-gray-500 dark:bg-dark-700 dark:text-dark-400">
              <tr>
                <th class="px-3 py-2">类型</th>
                <th class="px-3 py-2">数量</th>
                <th class="px-3 py-2">状态</th>
                <th class="px-3 py-2">激活/到期</th>
                <th class="px-3 py-2">备注</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr v-if="loading"><td colspan="5" class="px-3 py-6 text-center text-gray-500">加载中...</td></tr>
              <tr v-else-if="grants.length === 0"><td colspan="5" class="px-3 py-6 text-center text-gray-500">暂无限时权益</td></tr>
              <tr v-for="grant in grants" v-else :key="grant.id" class="text-gray-700 dark:text-gray-300">
                <td class="px-3 py-2">{{ grant.grant_type === 'balance' ? '限时余额' : '限时并发' }}</td>
                <td class="px-3 py-2">{{ grant.grant_type === 'balance' ? '$' : '' }}{{ formatAmount(grant.amount) }}</td>
                <td class="px-3 py-2"><span :class="statusClass(grant.status)">{{ statusText(grant.status) }}</span></td>
                <td class="px-3 py-2 text-xs">
                  <div>激活：{{ grant.activated_at ? formatDateTime(grant.activated_at) : '-' }}</div>
                  <div>到期：{{ grant.expires_at ? formatDateTime(grant.expires_at) : '-' }}</div>
                </td>
                <td class="max-w-48 truncate px-3 py-2" :title="grant.notes || ''">{{ grant.notes || '-' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button class="btn btn-secondary" @click="handleClose">关闭</button>
        <button type="submit" form="timed-grant-form" :disabled="submitting || !canSubmit" class="btn btn-primary">
          {{ submitting ? '保存中...' : '创建限时权益' }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import type { AdminUser } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { formatDateTime } from '@/utils/format'

interface TimedGrant {
  id: number
  user_id: number
  grant_type: 'balance' | 'concurrency'
  amount: number
  duration_seconds: number
  status: 'pending' | 'active' | 'expired'
  activated_at?: string | null
  expires_at?: string | null
  expired_at?: string | null
  deducted_amount: number
  notes: string
  created_at: string
  updated_at: string
}

const props = defineProps<{ show: boolean; user: AdminUser | null }>()
const emit = defineEmits(['close', 'success'])
const appStore = useAppStore()

const loading = ref(false)
const submitting = ref(false)
const grants = ref<TimedGrant[]>([])
const durationPreset = ref(86400)
const form = reactive({
  grant_type: 'balance' as 'balance' | 'concurrency',
  amount: 0,
  duration_seconds: 86400,
  notes: ''
})

const canSubmit = computed(() => form.amount > 0 && form.duration_seconds > 0 && (form.grant_type === 'balance' || Number.isInteger(form.amount)))

watch(() => props.show, (show) => {
  if (show) {
    form.grant_type = 'balance'
    form.amount = 0
    form.duration_seconds = 86400
    form.notes = ''
    durationPreset.value = 86400
    loadGrants()
  }
})

const applyDurationPreset = () => {
  if (durationPreset.value > 0) form.duration_seconds = durationPreset.value
}

const loadGrants = async () => {
  if (!props.user) return
  loading.value = true
  try {
    grants.value = await adminAPI.users.listTimedGrants(props.user.id) as TimedGrant[]
  } catch (e: any) {
    console.error('Failed to load timed grants:', e)
    appStore.showError(e.response?.data?.message || '加载限时权益失败')
  } finally {
    loading.value = false
  }
}

const handleSubmit = async () => {
  if (!props.user || !canSubmit.value) return
  submitting.value = true
  try {
    await adminAPI.users.createTimedGrant(props.user.id, {
      grant_type: form.grant_type,
      amount: form.grant_type === 'concurrency' ? Math.trunc(form.amount) : form.amount,
      duration_seconds: Math.trunc(form.duration_seconds),
      notes: form.notes
    })
    appStore.showSuccess('限时权益已创建，用户首次使用 API 时自动激活')
    await loadGrants()
    emit('success')
    form.amount = 0
    form.notes = ''
  } catch (e: any) {
    console.error('Failed to create timed grant:', e)
    appStore.showError(e.response?.data?.message || '创建限时权益失败')
  } finally {
    submitting.value = false
  }
}

const handleClose = () => emit('close')

const formatAmount = (value: number) => Number(value || 0).toFixed(8).replace(/\.?0+$/, '') || '0'
const statusText = (status: string) => status === 'pending' ? '待激活' : status === 'active' ? '生效中' : '已到期'
const statusClass = (status: string) => [
  'inline-flex rounded-full px-2 py-0.5 text-xs font-medium',
  status === 'pending' ? 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300' : status === 'active' ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300' : 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-dark-300'
]
</script>
