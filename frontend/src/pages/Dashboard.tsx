import { useState } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { useToast } from '@/components/ui/toast-provider'
import { videoService } from '@/services/videoService'
import { VideoCard } from '@/components/video/VideoCard'
import { VideoCardSkeleton } from '@/components/video/VideoCardSkeleton'
import { Plus, Loader2, Video, CheckCircle, Clock, TrendingUp, Sparkles, Zap } from 'lucide-react'

export function Dashboard() {
  const [url, setUrl] = useState('')
  const { showToast } = useToast()
  const navigate = useNavigate()

  const { data: videos, isLoading, refetch } = useQuery({
    queryKey: ['videos'],
    queryFn: () => videoService.getAll({ limit: 12 }),
  })

  const createMutation = useMutation({
    mutationFn: (url: string) => videoService.create(url),
    onSuccess: (video) => {
      setUrl('')
      refetch()
      showToast('Video added successfully!', 'success')
      // Navigate to video detail after a short delay
      setTimeout(() => {
        navigate(`/videos/${video.id}`)
      }, 500)
    },
    onError: (error: any) => {
      const errorMessage = error?.response?.data?.error || error?.message || 'Failed to add video'
      showToast(errorMessage, 'error')
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (url.trim()) {
      createMutation.mutate(url.trim())
    }
  }

  // Calculate statistics
  const stats = videos?.videos ? {
    total: videos.videos.length,
    completed: videos.videos.filter(v => v.status === 'completed').length,
    processing: videos.videos.filter(v => v.status === 'processing').length,
    pending: videos.videos.filter(v => v.status === 'pending').length,
    error: videos.videos.filter(v => v.status === 'error').length,
    withSummary: videos.videos.filter(v => v.hasSummary).length,
    withTranscript: videos.videos.filter(v => v.hasTranscript).length,
  } : { total: 0, completed: 0, processing: 0, pending: 0, error: 0, withSummary: 0, withTranscript: 0 }

  const recentVideos = videos?.videos?.slice(0, 6) || []

  return (
    <div className="space-y-8">
      {/* Header Section */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-4xl font-bold bg-gradient-to-r from-primary to-primary/60 bg-clip-text text-transparent">
            Dashboard
          </h1>
          <p className="text-muted-foreground mt-2 text-lg">
            Analyze and summarize YouTube videos with AI
          </p>
        </div>
      </div>

      {/* Quick Add Section */}
      <Card className="border-2 border-dashed border-primary/20 bg-gradient-to-br from-primary/5 to-primary/10">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Zap className="w-5 h-5 text-primary" />
            Quick Add Video
          </CardTitle>
          <CardDescription>
            Paste a YouTube URL to start analyzing and summarizing
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="flex gap-3">
              <Input
              id="video-url-input"
              type="url"
              placeholder="https://www.youtube.com/watch?v=..."
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              className="flex-1 h-12 text-base"
              disabled={createMutation.isPending}
            />
            <Button
              type="submit"
              size="lg"
              disabled={createMutation.isPending || !url.trim()}
              className="px-8"
            >
              {createMutation.isPending ? (
                <>
                  <Loader2 className="w-5 h-5 mr-2 animate-spin" />
                  Adding...
                </>
              ) : (
                <>
                  <Plus className="w-5 h-5 mr-2" />
                  Add Video
                </>
              )}
            </Button>
          </form>
        </CardContent>
      </Card>

      {/* Statistics Cards */}
      {!isLoading && videos && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card className="border-l-4 border-l-blue-500">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Videos</CardTitle>
              <Video className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.total}</div>
              <p className="text-xs text-muted-foreground mt-1">
                Videos in your library
              </p>
            </CardContent>
          </Card>

          <Card className="border-l-4 border-l-green-500">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Completed</CardTitle>
              <CheckCircle className="h-4 w-4 text-green-500" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.completed}</div>
              <p className="text-xs text-muted-foreground mt-1">
                {stats.total > 0 ? Math.round((stats.completed / stats.total) * 100) : 0}% of total
              </p>
            </CardContent>
          </Card>

          <Card className="border-l-4 border-l-yellow-500">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Processing</CardTitle>
              <Clock className="h-4 w-4 text-yellow-500" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.processing}</div>
              <p className="text-xs text-muted-foreground mt-1">
                Currently analyzing
              </p>
            </CardContent>
          </Card>

          <Card className="border-l-4 border-l-purple-500">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">With Summary</CardTitle>
              <Sparkles className="h-4 w-4 text-purple-500" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.withSummary}</div>
              <p className="text-xs text-muted-foreground mt-1">
                AI-generated summaries
              </p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Recent Videos Section */}
      <div>
        <div className="flex items-center justify-between mb-6">
          <div>
            <h2 className="text-2xl font-bold flex items-center gap-2">
              <TrendingUp className="w-6 h-6 text-primary" />
              Recent Videos
            </h2>
            <p className="text-muted-foreground mt-1">
              Your latest analyzed videos
            </p>
          </div>
          {videos && videos.videos.length > 0 && (
            <Button
              variant="outline"
              onClick={() => navigate('/videos')}
            >
              View All
            </Button>
          )}
        </div>

        {isLoading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {[...Array(8)].map((_, i) => (
              <VideoCardSkeleton key={i} />
            ))}
          </div>
        ) : recentVideos.length === 0 ? (
          <Card className="border-2 border-dashed">
            <CardContent className="py-16 text-center">
              <Video className="w-16 h-16 mx-auto mb-4 text-muted-foreground/50" />
              <h3 className="text-lg font-semibold mb-2">No videos yet</h3>
              <p className="text-muted-foreground mb-6 max-w-md mx-auto">
                Get started by adding your first YouTube video above. We'll analyze it and generate an AI-powered summary.
              </p>
              <Button onClick={() => document.getElementById('video-url-input')?.focus()}>
                <Plus className="w-4 h-4 mr-2" />
                Add Your First Video
              </Button>
            </CardContent>
          </Card>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {recentVideos.map((video) => (
              <VideoCard key={video.id} video={video} />
            ))}
          </div>
        )}
      </div>

      {/* Quick Actions */}
      {videos && videos.videos.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Quick Actions</CardTitle>
            <CardDescription>
              Common tasks and shortcuts
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <Button
                variant="outline"
                className="h-auto py-4 flex flex-col items-start"
                onClick={() => navigate('/videos')}
              >
                <Video className="w-5 h-5 mb-2" />
                <span className="font-semibold">Browse All Videos</span>
                <span className="text-xs text-muted-foreground mt-1">
                  View and manage all your videos
                </span>
              </Button>
              <Button
                variant="outline"
                className="h-auto py-4 flex flex-col items-start"
                onClick={() => navigate('/costs')}
              >
                <TrendingUp className="w-5 h-5 mb-2" />
                <span className="font-semibold">View Cost Analysis</span>
                <span className="text-xs text-muted-foreground mt-1">
                  Track your API usage and costs
                </span>
              </Button>
              <Button
                variant="outline"
                className="h-auto py-4 flex flex-col items-start"
                onClick={() => navigate('/settings')}
              >
                <Sparkles className="w-5 h-5 mb-2" />
                <span className="font-semibold">Configure Models</span>
                <span className="text-xs text-muted-foreground mt-1">
                  Customize AI model settings
                </span>
              </Button>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
