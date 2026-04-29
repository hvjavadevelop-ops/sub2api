import { apiClient } from '../client'
import type { BasePaginationResponse } from '@/types'

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

export interface AdminDailyCheckinListParams {
  page?: number
  page_size?: number
  search?: string
  user_id?: number | string
  start_date?: string
  end_date?: string
}

export const adminDailyCheckinAPI = {
  async list(params: AdminDailyCheckinListParams = {}): Promise<BasePaginationResponse<AdminDailyCheckinRecord>> {
    const { data } = await apiClient.get<BasePaginationResponse<AdminDailyCheckinRecord>>('/admin/daily-checkins', { params })
    return data
  },
}
