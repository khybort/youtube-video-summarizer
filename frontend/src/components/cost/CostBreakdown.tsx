import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { formatCurrency } from '@/lib/utils'
import type { CostSummary } from '@/types/cost'

interface CostBreakdownProps {
  summary: CostSummary
}

export function CostBreakdown({ summary }: CostBreakdownProps) {
  const totalCost = summary.totalCost

  return (
    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
      {/* By Provider */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">By Provider</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {Object.entries(summary.byProvider || {})
              .sort(([, a], [, b]) => b - a)
              .map(([provider, cost]) => (
                <div key={provider} className="space-y-1">
                  <div className="flex items-center justify-between text-sm">
                    <span className="capitalize font-medium">{provider}</span>
                    <span className="font-semibold">{formatCurrency(cost)}</span>
                  </div>
                  <div className="w-full bg-muted rounded-full h-2">
                    <div
                      className="bg-primary h-2 rounded-full transition-all"
                      style={{
                        width: `${totalCost > 0 ? (cost / totalCost) * 100 : 0}%`,
                      }}
                    />
                  </div>
                  <div className="text-xs text-muted-foreground">
                    {totalCost > 0 ? ((cost / totalCost) * 100).toFixed(1) : 0}% of total
                  </div>
                </div>
              ))}
          </div>
        </CardContent>
      </Card>

      {/* By Operation */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">By Operation</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {Object.entries(summary.byOperation || {})
              .sort(([, a], [, b]) => b - a)
              .map(([operation, cost]) => (
                <div key={operation} className="space-y-1">
                  <div className="flex items-center justify-between text-sm">
                    <span className="capitalize font-medium">{operation}</span>
                    <span className="font-semibold">{formatCurrency(cost)}</span>
                  </div>
                  <div className="w-full bg-muted rounded-full h-2">
                    <div
                      className="bg-primary h-2 rounded-full transition-all"
                      style={{
                        width: `${totalCost > 0 ? (cost / totalCost) * 100 : 0}%`,
                      }}
                    />
                  </div>
                  <div className="text-xs text-muted-foreground">
                    {totalCost > 0 ? ((cost / totalCost) * 100).toFixed(1) : 0}% of total
                  </div>
                </div>
              ))}
          </div>
        </CardContent>
      </Card>

      {/* By Model */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">By Model</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {Object.entries(summary.byModel || {})
              .sort(([, a], [, b]) => b - a)
              .slice(0, 5)
              .map(([model, cost]) => (
                <div key={model} className="space-y-1">
                  <div className="flex items-center justify-between text-sm">
                    <span className="font-medium truncate">{model}</span>
                    <span className="font-semibold">{formatCurrency(cost)}</span>
                  </div>
                  <div className="w-full bg-muted rounded-full h-2">
                    <div
                      className="bg-primary h-2 rounded-full transition-all"
                      style={{
                        width: `${totalCost > 0 ? (cost / totalCost) * 100 : 0}%`,
                      }}
                    />
                  </div>
                </div>
              ))}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

