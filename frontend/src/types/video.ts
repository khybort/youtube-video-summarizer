export interface Video {
  id: string
  youtubeId: string
  title: string
  description: string
  channelId: string
  channelName: string
  duration: number
  viewCount: number
  likeCount: number
  publishedAt: string
  thumbnailUrl: string
  tags: string[]
  status: 'pending' | 'processing' | 'completed' | 'error'
  hasTranscript: boolean
  hasSummary: boolean
  createdAt: string
  updatedAt: string
}

export interface Transcript {
  id: string
  videoId: string
  language: string
  source: 'youtube' | 'whisper'
  content: string
  segments: TranscriptSegment[]
}

export interface TranscriptSegment {
  start: number
  end: number
  text: string
}

export interface Summary {
  id: string
  videoId: string
  modelUsed: string
  summaryType: 'short' | 'detailed' | 'bullet_points'
  content: string
  keyPoints: string[]
  createdAt?: string
}

export interface SimilarVideo {
  video: Video
  similarityScore: number
  comparisonType: string
}

export interface Settings {
  id?: string
  transcript_provider: 'youtube' | 'groq' | 'local' | 'huggingface'
  summary_provider: 'gemini' | 'ollama'
  embedding_provider: 'gemini' | 'ollama'
  audio_analysis_provider: 'gemini' | 'ollama'
  ollama_model: string
  whisper_model: string
  gemini_model?: string
  ollama_url?: string
  local_whisper_url?: string
  summary_language?: string
  // API keys (only sent on update; GET /settings returns only has_* flags)
  gemini_api_key?: string
  groq_api_key?: string
  huggingface_api_key?: string
  has_gemini_api_key?: boolean
  has_groq_api_key?: boolean
  has_huggingface_api_key?: boolean
  created_at?: string
  updated_at?: string
}

