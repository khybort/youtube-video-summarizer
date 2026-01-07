import { useQuery } from '@tanstack/react-query'
import { videoService } from '@/services/videoService'
import type { Transcript } from '@/types/video'

export function useTranscript(videoId: string, hasTranscript?: boolean, language?: string) {
  return useQuery<Transcript>({
    queryKey: ['transcript', videoId, language],
    queryFn: () => videoService.getTranscript(videoId, language),
    enabled: !!videoId && hasTranscript !== false, // Only fetch if video has transcript
    retry: false, // Don't retry if transcript doesn't exist
    staleTime: 5 * 60 * 1000, // Consider data fresh for 5 minutes
  })
}

export function useAvailableLanguages(videoId: string) {
  return useQuery<Array<{ code: string; name: string; is_auto_generated: boolean }>>({
    queryKey: ['availableLanguages', videoId],
    queryFn: () => videoService.getAvailableLanguages(videoId),
    enabled: !!videoId,
    retry: false,
    staleTime: 10 * 60 * 1000, // Languages don't change often
  })
}

