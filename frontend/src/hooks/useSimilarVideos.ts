import { useQuery } from '@tanstack/react-query'
import { videoService } from '@/services/videoService'
import type { SimilarVideo } from '@/types/video'

export function useSimilarVideos(
  videoId: string,
  options?: { limit?: number; minScore?: number }
) {
  return useQuery<{ similar_videos: SimilarVideo[] }>({
    queryKey: ['similar-videos', videoId, options],
    queryFn: () => videoService.getSimilar(videoId, options),
    enabled: !!videoId,
  })
}

