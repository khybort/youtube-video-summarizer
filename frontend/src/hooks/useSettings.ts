import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { settingsService } from '@/services/settingsService'
import type { Settings } from '@/types/video'

export function useSettings() {
  return useQuery<Settings>({
    queryKey: ['settings'],
    queryFn: () => settingsService.get(),
  })
}

export function useUpdateSettings() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (settings: Partial<Settings>) => settingsService.update(settings),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['settings'] })
    },
  })
}

