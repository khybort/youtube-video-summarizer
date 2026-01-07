export interface TokenUsage {
  id: string
  videoId: string
  operation: 'transcription' | 'summarization' | 'embedding'
  provider: string
  model: string
  inputTokens: number
  outputTokens: number
  totalTokens: number
  cost: number
  createdAt: string
}

export interface CostSummary {
  totalCost: number
  totalTokens: number
  byProvider: Record<string, number>
  byOperation: Record<string, number>
  byModel: Record<string, number>
  period: 'today' | 'week' | 'month' | 'all'
  videoCount: number
  averageCostPerVideo: number
}

