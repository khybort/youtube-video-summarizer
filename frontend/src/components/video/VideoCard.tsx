import React from 'react'
import { useNavigate } from 'react-router-dom'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { VideoActions } from './VideoActions'
import { formatNumber, formatDuration } from '@/lib/utils'
import { Play, Clock, CheckCircle, FileText, Sparkles, AlertCircle, Loader2 } from 'lucide-react'
import type { Video } from '@/types/video'
import { cn } from '@/lib/utils'

interface VideoCardProps {
  video: Video
  showSimilarity?: boolean
  similarityScore?: number
  onClick?: (video: Video) => void | Promise<void>
}

const statusConfig: Record<string, {
  icon: React.ComponentType<{ className?: string }>
  color: string
  label: string
  animate?: boolean
}> = {
  completed: { 
    icon: CheckCircle, 
    color: 'bg-green-500/10 text-green-600 dark:text-green-400 border-green-500/20',
    label: 'Completed'
  },
  processing: { 
    icon: Loader2, 
    color: 'bg-yellow-500/10 text-yellow-600 dark:text-yellow-400 border-yellow-500/20',
    label: 'Processing',
    animate: true
  },
  pending: { 
    icon: Clock, 
    color: 'bg-blue-500/10 text-blue-600 dark:text-blue-400 border-blue-500/20',
    label: 'Pending'
  },
  error: { 
    icon: AlertCircle, 
    color: 'bg-red-500/10 text-red-600 dark:text-red-400 border-red-500/20',
    label: 'Error'
  },
}

export function VideoCard({ video, showSimilarity, similarityScore, onClick }: VideoCardProps) {
  const navigate = useNavigate()
  const statusInfo = statusConfig[video.status as keyof typeof statusConfig] || statusConfig.pending
  const StatusIcon = statusInfo.icon

  const handleClick = async () => {
    if (onClick) {
      await onClick(video)
    } else {
      navigate(`/videos/${video.id}`)
    }
  }

  return (
    <Card
      className="overflow-hidden hover:shadow-xl transition-all duration-300 cursor-pointer group border-2 hover:border-primary/20"
      onClick={handleClick}
    >
      <div className="relative aspect-video bg-muted overflow-hidden">
        {video.thumbnailUrl ? (
          <img
            src={video.thumbnailUrl}
            alt={video.title}
            className="w-full h-full object-cover group-hover:scale-110 transition-transform duration-500"
            onError={(e) => {
              const target = e.target as HTMLImageElement
              target.src = `https://i.ytimg.com/vi/${video.youtubeId}/hqdefault.jpg`
            }}
          />
        ) : (
          <div className="w-full h-full bg-gradient-to-br from-primary/20 via-primary/10 to-primary/5 flex items-center justify-center">
            <Play className="w-16 h-16 text-primary/40" />
          </div>
        )}
        
        {/* Overlay on hover */}
        <div className="absolute inset-0 bg-black/0 group-hover:bg-black/30 transition-colors flex items-center justify-center">
          <div className="w-16 h-16 rounded-full bg-white/90 dark:bg-black/90 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity shadow-lg">
            <Play className="w-8 h-8 text-primary ml-1" fill="currentColor" />
          </div>
        </div>

        {/* Duration badge */}
        <div className="absolute bottom-2 right-2 bg-black/80 backdrop-blur-sm text-white text-xs px-2 py-1 rounded-md flex items-center gap-1 font-medium">
          <Clock className="w-3 h-3" />
          {formatDuration(video.duration)}
        </div>

        {/* Similarity badge */}
        {showSimilarity && similarityScore !== undefined && (
          <div className="absolute top-2 left-2 bg-primary text-primary-foreground text-xs px-2 py-1 rounded-md font-semibold shadow-lg">
            {Math.round(similarityScore * 100)}% Match
          </div>
        )}

        {/* Status badge */}
        <div className={cn(
          "absolute top-2 right-2 px-2 py-1 rounded-md text-xs font-medium border backdrop-blur-sm",
          statusInfo.color,
          statusInfo.animate && "animate-pulse"
        )}>
          <div className="flex items-center gap-1">
            <StatusIcon className={cn("w-3 h-3", statusInfo.animate && "animate-spin")} />
            <span>{statusInfo.label}</span>
          </div>
        </div>
      </div>

      <CardHeader className="pb-3">
        <div className="flex items-start justify-between gap-2">
          <div className="flex-1 min-w-0">
            <CardTitle className="line-clamp-2 text-base leading-tight group-hover:text-primary transition-colors">
              {video.title}
            </CardTitle>
            <CardDescription className="line-clamp-1 mt-1">
              {video.channelName}
            </CardDescription>
          </div>
          <div onClick={(e) => e.stopPropagation()}>
            <VideoActions video={video} />
          </div>
        </div>
      </CardHeader>

      <CardContent className="pt-0">
        <div className="flex items-center justify-between text-sm">
          <div className="flex items-center gap-4 text-muted-foreground">
            <span className="flex items-center gap-1">
              <Play className="w-3 h-3" />
              {formatNumber(video.viewCount)}
            </span>
          </div>
          <div className="flex items-center gap-2">
            {video.hasSummary && (
              <Badge variant="secondary" className="text-xs">
                <Sparkles className="w-3 h-3 mr-1" />
                Summary
              </Badge>
            )}
            {video.hasTranscript && (
              <Badge variant="secondary" className="text-xs">
                <FileText className="w-3 h-3 mr-1" />
                Transcript
              </Badge>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
