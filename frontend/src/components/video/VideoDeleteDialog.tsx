import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { useDeleteVideo } from '@/hooks/useVideos'
import { Loader2, AlertTriangle } from 'lucide-react'
import { useToast } from '@/components/ui/toast-provider'

interface VideoDeleteDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  videoId: string
  videoTitle: string
  onDeleted?: () => void
}

export function VideoDeleteDialog({
  open,
  onOpenChange,
  videoId,
  videoTitle,
  onDeleted,
}: VideoDeleteDialogProps) {
  const deleteMutation = useDeleteVideo()
  const { showToast } = useToast()

  const handleDelete = () => {
    deleteMutation.mutate(videoId, {
      onSuccess: () => {
        showToast('Video deleted successfully', 'success')
        onOpenChange(false)
        onDeleted?.()
      },
      onError: (error: Error) => {
        showToast(error.message || 'Failed to delete video', 'error')
      },
    })
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <div className="flex items-center gap-2">
            <AlertTriangle className="w-5 h-5 text-destructive" />
            <DialogTitle>Delete Video</DialogTitle>
          </div>
          <DialogDescription>
            Are you sure you want to delete "{videoTitle}"? This action cannot be undone.
          </DialogDescription>
        </DialogHeader>
        <div className="flex justify-end gap-2 mt-4">
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={deleteMutation.isPending}
          >
            Cancel
          </Button>
          <Button
            variant="destructive"
            onClick={handleDelete}
            disabled={deleteMutation.isPending}
          >
            {deleteMutation.isPending ? (
              <>
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                Deleting...
              </>
            ) : (
              'Delete'
            )}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}

