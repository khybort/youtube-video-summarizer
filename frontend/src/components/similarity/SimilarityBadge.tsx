import { cn } from '@/lib/utils'

interface SimilarityBadgeProps {
  score: number // 0-1
}

export function SimilarityBadge({ score }: SimilarityBadgeProps) {
  const percentage = Math.round(score * 100)

  const colorClass =
    percentage >= 80
      ? 'bg-green-500'
      : percentage >= 60
      ? 'bg-yellow-500'
      : percentage >= 40
      ? 'bg-orange-500'
      : 'bg-red-500'

  return (
    <span
      className={cn(
        'px-2 py-1 rounded text-white text-sm font-semibold',
        colorClass
      )}
    >
      {percentage}% Match
    </span>
  )
}

