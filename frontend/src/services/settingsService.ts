import api from './api'
import type { Settings } from '@/types/video'

export const settingsService = {
  get: () =>
    api.get<Settings>('/settings').then(res => res.data),

  update: (settings: Partial<Settings>) =>
    api.put<Settings>('/settings', settings).then(res => res.data),
    
  checkLocalWhisperHealth: () =>
    api.get<{ available: boolean; url: string }>('/settings/health/local-whisper').then(res => res.data),
}

