import api, { apiWithExtendedTimeout } from './api'
import type { Video, Transcript, Summary } from '@/types/video'

// Transform backend snake_case to frontend camelCase
function transformVideo(data: any): Video {
  return {
    id: data.id,
    youtubeId: data.youtube_id || data.youtubeId,
    title: data.title || '',
    description: data.description || '',
    channelId: data.channel_id || data.channelId,
    channelName: data.channel_name || data.channelName,
    duration: data.duration || 0,
    viewCount: data.view_count || data.viewCount || 0,
    likeCount: data.like_count || data.likeCount || 0,
    publishedAt: data.published_at || data.publishedAt,
    thumbnailUrl: data.thumbnail_url || data.thumbnailUrl || '',
    tags: data.tags || [],
    status: data.status || 'pending',
    hasTranscript: data.has_transcript || data.hasTranscript || false,
    hasSummary: data.has_summary || data.hasSummary || false,
    createdAt: data.created_at || data.createdAt,
    updatedAt: data.updated_at || data.updatedAt,
  }
}

export const videoService = {
  getAll: (params?: { page?: number; limit?: number; offset?: number }) => {
    // Convert page to offset if page is provided (backend uses offset, not page)
    const queryParams: any = { ...params }
    if (queryParams.page && !queryParams.offset) {
      const limit = queryParams.limit || 20
      queryParams.offset = (queryParams.page - 1) * limit
      delete queryParams.page
    }
    // Remove search if present (backend doesn't support it yet)
    delete queryParams.search
    return api.get<{ videos: any[]; total: number; limit: number; offset: number }>('/videos', { params: queryParams }).then(res => ({
      videos: res.data.videos.map(transformVideo),
      total: res.data.total,
      limit: res.data.limit,
      offset: res.data.offset,
    }))
  },

  getById: (id: string) =>
    api.get<any>(`/videos/${id}`).then(res => transformVideo(res.data)),

  create: (url: string) =>
    api.post<any>('/videos', { url }).then(res => transformVideo(res.data)),

  delete: (id: string) =>
    api.delete(`/videos/${id}`).then(res => res.data),

  analyze: (id: string) =>
    apiWithExtendedTimeout.post(`/videos/${id}/analyze`).then(res => res.data),

  getTranscript: (id: string, language?: string) => {
    const params = language ? { language } : {}
    return api.get<any>(`/videos/${id}/transcript`, { params }).then(res => ({
      ...res.data,
      videoId: res.data.video_id || res.data.videoId,
      createdAt: res.data.created_at || res.data.createdAt,
    } as Transcript))
  },

  getAvailableLanguages: (id: string) =>
    api.get<{ languages: Array<{ code: string; name: string; is_auto_generated: boolean }> }>(`/videos/${id}/transcript/languages`).then(res => res.data.languages),

  getSummary: (id: string, language?: string) => {
    const params = language ? { language } : {}
    // Use extended timeout because summary generation can take time (transcription + LLM)
    return apiWithExtendedTimeout.get<any>(`/videos/${id}/summary`, { params }).then(res => ({
      ...res.data,
      videoId: res.data.video_id || res.data.videoId,
      modelUsed: res.data.model_used || res.data.modelUsed,
      summaryType: res.data.summary_type || res.data.summaryType,
      keyPoints: res.data.key_points || res.data.keyPoints || [],
      createdAt: res.data.created_at || res.data.createdAt || new Date().toISOString(),
    } as Summary))
  },

  summarize: (id: string, opts?: { type?: 'short' | 'detailed' | 'bullet_points'; from_audio?: boolean; language?: string }) =>
    apiWithExtendedTimeout.post(`/videos/${id}/summarize`, opts).then(res => res.data),

  translateSummary: (id: string, language: string) =>
    api.post(`/videos/${id}/summary/translate`, { language }).then(res => res.data),

  getSimilar: (id: string, params?: { limit?: number; min_score?: number }) => {
    // Backend expects min_score, not minScore
    const queryParams: any = {}
    if (params?.limit) queryParams.limit = params.limit
    if (params?.min_score !== undefined) queryParams.min_score = params.min_score
    return api.get<{ similar_videos: any[] }>(`/videos/${id}/similar`, { params: queryParams }).then(res => ({
      similar_videos: res.data.similar_videos.map((item: any) => ({
        video: transformVideo(item.video),
        similarityScore: item.similarity_score || item.similarityScore,
        comparisonType: item.comparison_type || item.comparisonType,
      })),
    }))
  },
}

