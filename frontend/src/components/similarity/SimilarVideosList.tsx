import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { videoService } from '@/services/videoService'
import { VideoCard } from '@/components/video/VideoCard'
import { Card, CardContent } from '@/components/ui/card'
import { useToast } from '@/components/ui/toast-provider'
import { Loader2 } from 'lucide-react'
import { useState } from 'react'

interface SimilarVideosListProps {
  videoId: string
  limit?: number
  minScore?: number
}

export function SimilarVideosList({
  videoId,
  limit = 10,
  minScore = 0.5,
}: SimilarVideosListProps) {
  const navigate = useNavigate()
  const { showToast } = useToast()
  const queryClient = useQueryClient()
  const [addingVideoId, setAddingVideoId] = useState<string | null>(null)

  const { data, isLoading } = useQuery({
    queryKey: ['similar-videos', videoId, limit, minScore],
    queryFn: () => videoService.getSimilar(videoId, { limit, min_score: minScore }),
    enabled: !!videoId,
  })

  const createMutation = useMutation({
    mutationFn: (url: string) => videoService.create(url),
    onSuccess: (newVideo) => {
      queryClient.invalidateQueries({ queryKey: ['videos'] })
      setAddingVideoId(null)
      showToast('Video added successfully!', 'success')
      // Navigate to the new video detail page
      navigate(`/videos/${newVideo.id}`)
    },
    onError: (error: Error) => {
      setAddingVideoId(null)
      showToast(error.message || 'Failed to add video', 'error')
    },
  })

  const handleVideoClick = async (video: any) => {
    // Check if video is already in database
    // Backend returns "00000000-0000-0000-0000-000000000000" for videos not in database
    const isEmptyUUID = video.id === '00000000-0000-0000-0000-000000000000' || !video.id
    
    // Always add video if it has empty UUID (not in database)
    if (isEmptyUUID) {
      // Video is not in database, add it first
      const youtubeUrl = `https://www.youtube.com/watch?v=${video.youtubeId}`
      setAddingVideoId(video.youtubeId)
      createMutation.mutate(youtubeUrl)
    } else {
      // Video is already in database, navigate directly
      navigate(`/videos/${video.id}`)
    }
  }

  if (isLoading) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        Loading similar videos...
      </div>
    )
  }

  const similarVideos = data?.similar_videos || []

  if (similarVideos.length === 0) {
    return (
      <Card>
        <CardContent className="py-8 text-center text-muted-foreground">
          No similar videos found
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="space-y-4">
      <div>
        <h3 className="text-xl font-semibold mb-2">Similar Videos</h3>
        <p className="text-sm text-muted-foreground">
          Found {similarVideos.length} similar video{similarVideos.length !== 1 ? 's' : ''}
        </p>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
        {similarVideos.map((item) => (
          <div key={item.video.id || item.video.youtubeId} className="relative">
            {addingVideoId === item.video.youtubeId && (
              <div className="absolute inset-0 bg-background/80 z-10 flex items-center justify-center rounded-lg">
                <Loader2 className="w-6 h-6 animate-spin text-primary" />
              </div>
            )}
            <VideoCard
              video={item.video}
              showSimilarity
              similarityScore={item.similarityScore}
              onClick={handleVideoClick}
            />
          </div>
        ))}
      </div>
    </div>
  )
}

