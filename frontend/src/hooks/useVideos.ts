import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { videoService } from '@/services/videoService'

export function useVideos(params?: { page?: number; limit?: number; offset?: number }) {
  return useQuery({
    queryKey: ['videos', params],
    queryFn: () => videoService.getAll(params),
  })
}

export function useVideo(id: string) {
  return useQuery({
    queryKey: ['video', id],
    queryFn: () => videoService.getById(id),
    enabled: !!id,
  })
}

export function useCreateVideo() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (url: string) => videoService.create(url),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['videos'] })
    },
  })
}

export function useDeleteVideo() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => videoService.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['videos'] })
    },
  })
}

export function useAnalyzeVideo() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => videoService.analyze(id),
    onSuccess: (_, videoId) => {
      queryClient.invalidateQueries({ queryKey: ['video', videoId] })
      queryClient.invalidateQueries({ queryKey: ['videos'] })
    },
  })
}

