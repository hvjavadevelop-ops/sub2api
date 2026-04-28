import { apiClient } from '../client'

export interface AdminDailyCheckinRecord {
  id: number
  user_id: number
  user_email?: string
  username?: string
  checkin_date: string
  reward: number
  balance_after: number
  created_at: string
}

export interface AdminDailyCheckinListResult {
  records: AdminDailyCheckinRecord[]
  total: number
  page: number
  limit: number
}

export const adminDailyCheckinAPI = {
  list(params: { page?: number; limit?: number } = {}) {
    return apiClient.get<AdminDailyCheckinListResult>('/admin/daily-checkins', { params })
  },
}
