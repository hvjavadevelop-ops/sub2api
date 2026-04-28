<template>
  <div>
    <button
      @click="handleCheckin"
      :disabled="loading || checkedInToday"
      class="relative flex h-9 items-center justify-center rounded-lg px-2 text-sm font-medium transition-all hover:scale-105 disabled:cursor-default disabled:hover:scale-100"
      :class="checkedInToday
        ? 'w-9 bg-emerald-50 text-emerald-600 dark:bg-emerald-900/20 dark:text-emerald-400'
        : 'min-w-[3.25rem] text-gray-600 hover:bg-amber-50 hover:text-amber-600 dark:text-gray-400 dark:hover:bg-amber-900/20 dark:hover:text-amber-400'"
      :title="buttonTitle"
      aria-label="每日签到"
    >
      <span v-if="loading" class="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent"></span>
      <span v-else-if="!checkedInToday">签到</span>
      <img v-else :src="checkmarkIcon" alt="已签到" class="h-5 w-5" />
      <span
        v-if="!checkedInToday && !loading"
        class="absolute right-1 top-1 flex h-1.5 w-1.5"
      >
        <span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-amber-500 opacity-75"></span>
        <span class="relative inline-flex h-1.5 w-1.5 rounded-full bg-amber-500"></span>
      </span>
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { dailyCheckinAPI } from '@/api'
import { useAppStore, useAuthStore } from '@/stores'
import checkmarkIcon from '@/assets/icons/checkmark.svg'

const authStore = useAuthStore()
const appStore = useAppStore()

const loading = ref(false)
const checkedInToday = ref(false)
const minReward = ref(10)
const maxReward = ref(29)

const buttonTitle = computed(() => {
  if (loading.value) return '签到中...'
  if (checkedInToday.value) return '今日已签到'
  return `每日签到，随机领取 $${minReward.value}-$${maxReward.value} 余额`
})

async function loadStatus() {
  try {
    const status = await dailyCheckinAPI.status()
    checkedInToday.value = status.checked_in_today
    minReward.value = status.min_reward
    maxReward.value = status.max_reward
  } catch {
    // 顶栏入口不因状态读取失败打扰用户
  }
}

async function handleCheckin() {
  if (loading.value || checkedInToday.value) return
  loading.value = true
  try {
    const result = await dailyCheckinAPI.checkin()
    checkedInToday.value = true
    if (typeof result.balance_after === 'number' && authStore.user) {
      authStore.user.balance = result.balance_after
      localStorage.setItem('auth_user', JSON.stringify(authStore.user))
    } else {
      await authStore.refreshUser().catch(() => undefined)
    }
    appStore.showSuccess(`签到成功，获得 $${result.reward?.toFixed(0) || ''} 余额`)
  } catch (error: any) {
    if (error?.reason === 'DAILY_CHECKIN_ALREADY_DONE' || error?.message?.includes('already')) {
      checkedInToday.value = true
      appStore.showInfo('今天已经签到过了，明天再来')
    } else {
      appStore.showError(error?.message || '签到失败，请稍后再试')
    }
  } finally {
    loading.value = false
  }
}

onMounted(loadStatus)
</script>
