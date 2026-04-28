import { apiClient } from './client'

export interface DailyCheckinRecord {
  id: number
  user_id: number
  checkin_date: string
  reward: number
  balance_after: number
  created_at: string
}

export interface DailyCheckinResult {
  checked_in_today: boolean
  min_reward: number
  max_reward: number
  reward?: number
  balance_after?: number
  today?: DailyCheckinRecord
}

export const dailyCheckinAPI = {
  async status(): Promise<DailyCheckinResult> {
    const response = await apiClient.get('/user/daily-checkin')
    return response.data
  },

  async checkin(): Promise<DailyCheckinResult> {
    const response = await apiClient.post('/user/daily-checkin')
    return response.data
  }
}
