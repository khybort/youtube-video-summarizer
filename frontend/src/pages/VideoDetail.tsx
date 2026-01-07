import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import { videoService } from '@/services/videoService'
import { formatNumber, formatDuration } from '@/lib/utils'
import { ArrowLeft, Clock, Eye, ThumbsUp, Loader2, Play, FileText, Sparkles, CheckCircle, AlertCircle } from 'lucide-react'
import ReactPlayer from 'react-player'
import { SimilarVideosList } from '@/components/similarity/SimilarVideosList'
import { TranscriptViewer } from '@/components/transcript/TranscriptViewer'
import { SummaryDisplay } from '@/components/summary/SummaryDisplay'
import { VideoCostCard } from '@/components/cost/VideoCostCard'
import { useState, useEffect } from 'react'
import { cn } from '@/lib/utils'

export function VideoDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [activeTab, setActiveTab] = useState<'overview' | 'transcript' | 'summary' | 'similar'>('overview')
  const [playerRef, setPlayerRef] = useState<any>(null)

  const { data: video, isLoading, refetch } = useQuery({
    queryKey: ['video', id],
    queryFn: () => videoService.getById(id!),
    enabled: !!id,
    refetchOnWindowFocus: true,
    refetchOnMount: true,
    select: (data) => {
      const videosCache = queryClient.getQueryData<{ videos: any[] }>(['videos'])
      if (videosCache) {
        const cachedVideo = videosCache.videos.find((v) => v.id === id)
        if (cachedVideo && cachedVideo.status !== data.status) {
          return { ...data, status: cachedVideo.status }
        }
      }
      return data
    },
  })

  useEffect(() => {
    if (id && video) {
      if (video.status === 'processing' || video.status === 'pending') {
        const interval = setInterval(() => {
          refetch()
        }, 5000)
        return () => clearInterval(interval)
      }
    }
  }, [id, video?.status, refetch])

  const analyzeMutation = useMutation({
    mutationFn: (videoId: string) => videoService.analyze(videoId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['video', id] })
      queryClient.invalidateQueries({ queryKey: ['videos'] })
      refetch()
    },
  })

  const handleAnalyze = () => {
    if (id) {
      analyzeMutation.mutate(id)
    }
  }

  const handleSeek = (time: number) => {
    if (playerRef) {
      playerRef.seekTo(time)
    }
  }

  if (isLoading) {
    return (
      <div className="text-center py-12">
        <Loader2 className="w-8 h-8 animate-spin mx-auto mb-4 text-muted-foreground" />
        <p className="text-muted-foreground">Loading video details...</p>
      </div>
    )
  }

  if (!video) {
    return (
      <div className="text-center py-12">
        <AlertCircle className="w-16 h-16 mx-auto mb-4 text-muted-foreground/50" />
        <h2 className="text-2xl font-bold mb-2">Video not found</h2>
        <p className="text-muted-foreground mb-6">The video you're looking for doesn't exist.</p>
        <Button onClick={() => navigate('/videos')}>
          <ArrowLeft className="w-4 h-4 mr-2" />
          Back to Videos
        </Button>
      </div>
    )
  }

  const statusConfig: Record<string, { icon: React.ComponentType<{ className?: string }>, color: string, label: string }> = {
    completed: { icon: CheckCircle, color: 'bg-green-500/10 text-green-600 dark:text-green-400 border-green-500/20', label: 'Completed' },
    processing: { icon: Loader2, color: 'bg-yellow-500/10 text-yellow-600 dark:text-yellow-400 border-yellow-500/20', label: 'Processing' },
    pending: { icon: Clock, color: 'bg-blue-500/10 text-blue-600 dark:text-blue-400 border-blue-500/20', label: 'Pending' },
    error: { icon: AlertCircle, color: 'bg-red-500/10 text-red-600 dark:text-red-400 border-red-500/20', label: 'Error' },
  }
  const statusInfo = statusConfig[video.status] || statusConfig.pending
  const StatusIcon = statusInfo.icon

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <Button
          variant="ghost"
          onClick={() => navigate('/videos')}
          className="gap-2"
        >
          <ArrowLeft className="w-4 h-4" />
          Back to Videos
        </Button>
        <Badge className={cn("px-3 py-1", statusInfo.color)}>
          <StatusIcon className={cn("w-3 h-3 mr-1", video.status === 'processing' && "animate-spin")} />
          {statusInfo.label}
        </Badge>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Main Content */}
        <div className="lg:col-span-2 space-y-6">
          {/* Video Player */}
          <Card className="overflow-hidden">
            <CardContent className="p-0">
              <div className="aspect-video bg-black">
                <ReactPlayer
                  ref={setPlayerRef}
                  url={`https://www.youtube.com/watch?v=${video.youtubeId}`}
                  width="100%"
                  height="100%"
                  controls
                  playing={false}
                />
              </div>
            </CardContent>
          </Card>

          {/* Video Info */}
          <Card>
            <CardHeader>
              <CardTitle className="text-2xl leading-tight">{video.title}</CardTitle>
              <CardDescription className="text-base mt-2">{video.channelName}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex flex-wrap items-center gap-6 text-sm">
                <div className="flex items-center gap-2 text-muted-foreground">
                  <Eye className="w-4 h-4" />
                  <span className="font-medium">{formatNumber(video.viewCount)}</span>
                  <span>views</span>
                </div>
                <div className="flex items-center gap-2 text-muted-foreground">
                  <ThumbsUp className="w-4 h-4" />
                  <span className="font-medium">{formatNumber(video.likeCount)}</span>
                  <span>likes</span>
                </div>
                <div className="flex items-center gap-2 text-muted-foreground">
                  <Clock className="w-4 h-4" />
                  <span>{formatDuration(video.duration)}</span>
                </div>
                <div className="text-muted-foreground">
                  Published {new Date(video.publishedAt).toLocaleDateString()}
                </div>
              </div>

              {video.description && (
                <div className="pt-4 border-t">
                  <h3 className="font-semibold mb-2">Description</h3>
                  <p className="text-sm text-muted-foreground whitespace-pre-wrap line-clamp-6">
                    {video.description}
                  </p>
                </div>
              )}
            </CardContent>
          </Card>

          {/* Action Button */}
          <Card>
            <CardContent className="pt-6">
              <Button
                onClick={handleAnalyze}
                disabled={analyzeMutation.isPending || video.status === 'processing'}
                size="lg"
                className="w-full"
              >
                {analyzeMutation.isPending || video.status === 'processing' ? (
                  <>
                    <Loader2 className="w-5 h-5 mr-2 animate-spin" />
                    Analyzing...
                  </>
                ) : (
                  <>
                    <Sparkles className="w-5 h-5 mr-2" />
                    Analyze Video
                  </>
                )}
              </Button>
            </CardContent>
          </Card>

          {/* Tabs */}
          <Card>
            <CardContent className="pt-6">
              <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as any)}>
                <TabsList className="grid w-full grid-cols-4">
                  <TabsTrigger value="overview">Overview</TabsTrigger>
                  <TabsTrigger value="transcript" className="gap-2">
                    <FileText className="w-4 h-4" />
                    Transcript
                  </TabsTrigger>
                  <TabsTrigger value="summary" className="gap-2">
                    <Sparkles className="w-4 h-4" />
                    Summary
                  </TabsTrigger>
                  <TabsTrigger value="similar" className="gap-2">
                    <Play className="w-4 h-4" />
                    Similar
                  </TabsTrigger>
                </TabsList>

                <TabsContent value="overview" className="mt-6">
                  {video.description && (
                    <div className="prose prose-sm max-w-none dark:prose-invert">
                      <h3 className="font-semibold mb-3">Full Description</h3>
                      <p className="text-muted-foreground whitespace-pre-wrap">
                        {video.description}
                      </p>
                    </div>
                  )}
                  {!video.description && (
                    <div className="text-center py-8 text-muted-foreground">
                      No description available
                    </div>
                  )}
                </TabsContent>

                <TabsContent value="transcript" className="mt-6">
                  <TranscriptViewer videoId={video.id} hasTranscript={video.hasTranscript} onSeek={handleSeek} />
                </TabsContent>

                <TabsContent value="summary" className="mt-6">
                  <SummaryDisplay videoId={video.id} />
                </TabsContent>

                <TabsContent value="similar" className="mt-6">
                  <SimilarVideosList videoId={video.id} />
                </TabsContent>
              </Tabs>
            </CardContent>
          </Card>
        </div>

        {/* Sidebar */}
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Video Information</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-3">
                <div className="flex justify-between items-center">
                  <span className="text-sm text-muted-foreground">Status</span>
                  <Badge className={cn("px-2 py-1 text-xs", statusInfo.color)}>
                    <StatusIcon className={cn("w-3 h-3 mr-1", video.status === 'processing' && "animate-spin")} />
                    {statusInfo.label}
                  </Badge>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm text-muted-foreground">Published</span>
                  <span className="text-sm font-medium">{new Date(video.publishedAt).toLocaleDateString()}</span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm text-muted-foreground">Transcript</span>
                  {video.hasTranscript ? (
                    <Badge variant="secondary" className="gap-1">
                      <CheckCircle className="w-3 h-3" />
                      Available
                    </Badge>
                  ) : (
                    <span className="text-sm text-muted-foreground">Not available</span>
                  )}
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm text-muted-foreground">Summary</span>
                  {video.hasSummary ? (
                    <Badge variant="secondary" className="gap-1">
                      <Sparkles className="w-3 h-3" />
                      Available
                    </Badge>
                  ) : (
                    <span className="text-sm text-muted-foreground">Not available</span>
                  )}
                </div>
              </div>
            </CardContent>
          </Card>

          <VideoCostCard videoId={video.id} />
        </div>
      </div>
    </div>
  )
}
