import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { costService } from '@/services/costService'
import { DollarSign, Zap } from 'lucide-react'
import { formatCurrency, formatNumber } from '@/lib/utils'
import { Skeleton } from '@/components/ui/skeleton'

interface VideoCostCardProps {
  videoId: string
}

export function VideoCostCard({ videoId }: VideoCostCardProps) {
  const { data, isLoading } = useQuery({
    queryKey: ['video-cost', videoId],
    queryFn: () => costService.getVideoUsage(videoId),
  })

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <Skeleton className="h-5 w-32" />
        </CardHeader>
        <CardContent>
          <Skeleton className="h-4 w-24" />
        </CardContent>
      </Card>
    )
  }

  const usage = data?.usage || []
  const totalCost = usage.reduce((sum, item) => sum + item.cost, 0)
  const totalTokens = usage.reduce((sum, item) => sum + item.totalTokens, 0)

  if (usage.length === 0) {
    return null
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">Cost Breakdown</CardTitle>
        <CardDescription>Token usage and costs for this video</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <DollarSign className="w-4 h-4 text-muted-foreground" />
            <span className="text-sm font-medium">Total Cost</span>
          </div>
          <span className="text-lg font-semibold">{formatCurrency(totalCost)}</span>
        </div>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Zap className="w-4 h-4 text-muted-foreground" />
            <span className="text-sm font-medium">Total Tokens</span>
          </div>
          <span className="text-sm">{formatNumber(totalTokens)}</span>
        </div>
        <div className="pt-2 border-t space-y-2">
          {usage.map((item) => (
            <div key={item.id} className="flex items-center justify-between text-sm">
              <div>
                <span className="font-medium capitalize">{item.operation}</span>
                <span className="text-muted-foreground ml-2">({item.provider})</span>
              </div>
              <div className="text-right">
                <div className="font-semibold">{formatCurrency(item.cost)}</div>
                <div className="text-xs text-muted-foreground">
                  {formatNumber(item.totalTokens)} tokens
                </div>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}

