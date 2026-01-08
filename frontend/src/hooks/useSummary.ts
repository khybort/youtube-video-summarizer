import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { videoService } from '@/services/videoService'
import type { Summary } from '@/types/video'

export function useSummary(videoId: string, language?: string) {
  return useQuery<Summary>({
    queryKey: ['summary', videoId, language],
    queryFn: () => videoService.getSummary(videoId, language),
    enabled: !!videoId,
  })
}

export function useSummarizeVideo() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ videoId, type, fromAudio, language }: { videoId: string; type?: 'short' | 'detailed' | 'bullet_points'; fromAudio?: boolean; language?: string }) =>
      videoService.summarize(videoId, { type, from_audio: fromAudio, language }),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['summary', variables.videoId] })
      queryClient.invalidateQueries({ queryKey: ['video', variables.videoId] })
    },
  })
}

