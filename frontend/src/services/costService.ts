import api from './api'
import type { CostSummary, TokenUsage } from '@/types/cost'

export const costService = {
  getSummary: (period: string = 'month') =>
    api.get<CostSummary>('/costs/summary', { params: { period } }).then(res => res.data),

  getUsage: (period: string = 'month') =>
    api.get<{ usage: TokenUsage[] }>('/costs/usage', { params: { period } }).then(res => res.data),

  getVideoUsage: (videoId: string) =>
    api.get<{ usage: TokenUsage[] }>(`/costs/videos/${videoId}/usage`).then(res => res.data),
}

