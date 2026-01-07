import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { VideoDeleteDialog } from './VideoDeleteDialog'
import { useAnalyzeVideo } from '@/hooks/useVideos'
import { useToast } from '@/components/ui/toast-provider'
import { MoreVertical, Trash2, Sparkles, Loader2 } from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import type { Video } from '@/types/video'

interface VideoActionsProps {
  video: Video
  onDeleted?: () => void
}

export function VideoActions({ video, onDeleted }: VideoActionsProps) {
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const analyzeMutation = useAnalyzeVideo()
  const { showToast } = useToast()

  const handleAnalyze = () => {
    analyzeMutation.mutate(video.id, {
      onSuccess: () => {
        showToast('Analysis started', 'success')
      },
      onError: (error: Error) => {
        showToast(error.message || 'Failed to start analysis', 'error')
      },
    })
  }

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="icon">
            <MoreVertical className="w-4 h-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem onClick={handleAnalyze} disabled={analyzeMutation.isPending || video.status === 'processing'}>
            {analyzeMutation.isPending || video.status === 'processing' ? (
              <>
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                Analyzing...
              </>
            ) : (
              <>
                <Sparkles className="w-4 h-4 mr-2" />
                Analyze
              </>
            )}
          </DropdownMenuItem>
          <DropdownMenuItem
            onClick={() => setDeleteDialogOpen(true)}
            className="text-destructive"
          >
            <Trash2 className="w-4 h-4 mr-2" />
            Delete
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <VideoDeleteDialog
        open={deleteDialogOpen}
        onOpenChange={setDeleteDialogOpen}
        videoId={video.id}
        videoTitle={video.title}
        onDeleted={onDeleted}
      />
    </>
  )
}

