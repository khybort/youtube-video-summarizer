import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Select } from '@/components/ui/select'
import { costService } from '@/services/costService'
import { DollarSign, TrendingUp, FileText, Zap, Loader2 } from 'lucide-react'
import { formatCurrency, formatNumber } from '@/lib/utils'

export function CostAnalysis() {
  const [period, setPeriod] = useState<'today' | 'week' | 'month' | 'all'>('month')

  const { data: summary, isLoading } = useQuery({
    queryKey: ['cost-summary', period],
    queryFn: () => costService.getSummary(period),
  })

  const { data: usage } = useQuery({
    queryKey: ['cost-usage', period],
    queryFn: () => costService.getUsage(period),
  })

  if (isLoading) {
    return (
      <div className="text-center py-12">
        <Loader2 className="w-6 h-6 animate-spin mx-auto mb-2 text-muted-foreground" />
        <p className="text-muted-foreground">Loading cost analysis...</p>
      </div>
    )
  }

  const costData = summary

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Cost Analysis</h1>
          <p className="text-muted-foreground mt-2">
            Track token usage and costs across all operations
          </p>
        </div>
        <Select
          value={period}
          onChange={(e) => setPeriod(e.target.value as any)}
          className="w-40"
        >
          <option value="today">Today</option>
          <option value="week">Last 7 Days</option>
          <option value="month">Last 30 Days</option>
          <option value="all">All Time</option>
        </Select>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Cost</CardTitle>
            <DollarSign className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {costData ? formatCurrency(costData.totalCost) : '$0.00'}
            </div>
            <p className="text-xs text-muted-foreground">
              {costData?.videoCount || 0} videos processed
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Tokens</CardTitle>
            <Zap className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {costData ? formatNumber(costData.totalTokens) : '0'}
            </div>
            <p className="text-xs text-muted-foreground">
              {costData ? formatNumber(costData.totalTokens / 1000) : '0'}K tokens
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Avg per Video</CardTitle>
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {costData ? formatCurrency(costData.averageCostPerVideo) : '$0.00'}
            </div>
            <p className="text-xs text-muted-foreground">
              Average cost per video
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Videos</CardTitle>
            <FileText className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {costData?.videoCount || 0}
            </div>
            <p className="text-xs text-muted-foreground">
              Videos analyzed
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Breakdown Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* By Provider */}
        <Card>
          <CardHeader>
            <CardTitle>Cost by Provider</CardTitle>
            <CardDescription>Breakdown by AI provider</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {costData?.byProvider && Object.entries(costData.byProvider).length > 0 ? (
                Object.entries(costData.byProvider)
                  .sort(([, a], [, b]) => b - a)
                  .map(([provider, cost]) => (
                    <div key={provider} className="space-y-2">
                      <div className="flex items-center justify-between">
                        <span className="text-sm font-medium capitalize">{provider}</span>
                        <span className="text-sm font-semibold">{formatCurrency(cost)}</span>
                      </div>
                      <div className="w-full bg-muted rounded-full h-2">
                        <div
                          className="bg-primary h-2 rounded-full"
                          style={{
                            width: `${(cost / costData.totalCost) * 100}%`,
                          }}
                        />
                      </div>
                    </div>
                  ))
              ) : (
                <p className="text-sm text-muted-foreground">No data available</p>
              )}
            </div>
          </CardContent>
        </Card>

        {/* By Operation */}
        <Card>
          <CardHeader>
            <CardTitle>Cost by Operation</CardTitle>
            <CardDescription>Breakdown by operation type</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {costData?.byOperation && Object.entries(costData.byOperation).length > 0 ? (
                Object.entries(costData.byOperation)
                  .sort(([, a], [, b]) => b - a)
                  .map(([operation, cost]) => (
                    <div key={operation} className="space-y-2">
                      <div className="flex items-center justify-between">
                        <span className="text-sm font-medium capitalize">{operation}</span>
                        <span className="text-sm font-semibold">{formatCurrency(cost)}</span>
                      </div>
                      <div className="w-full bg-muted rounded-full h-2">
                        <div
                          className="bg-primary h-2 rounded-full"
                          style={{
                            width: `${(cost / costData.totalCost) * 100}%`,
                          }}
                        />
                      </div>
                    </div>
                  ))
              ) : (
                <p className="text-sm text-muted-foreground">No data available</p>
              )}
            </div>
          </CardContent>
        </Card>

        {/* By Model */}
        <Card>
          <CardHeader>
            <CardTitle>Cost by Model</CardTitle>
            <CardDescription>Breakdown by AI model</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {costData?.byModel && Object.entries(costData.byModel).length > 0 ? (
                Object.entries(costData.byModel)
                  .sort(([, a], [, b]) => b - a)
                  .slice(0, 5)
                  .map(([model, cost]) => (
                    <div key={model} className="space-y-2">
                      <div className="flex items-center justify-between">
                        <span className="text-sm font-medium truncate">{model}</span>
                        <span className="text-sm font-semibold">{formatCurrency(cost)}</span>
                      </div>
                      <div className="w-full bg-muted rounded-full h-2">
                        <div
                          className="bg-primary h-2 rounded-full"
                          style={{
                            width: `${(cost / costData.totalCost) * 100}%`,
                          }}
                        />
                      </div>
                    </div>
                  ))
              ) : (
                <p className="text-sm text-muted-foreground">No data available</p>
              )}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Usage Table */}
      <Card>
        <CardHeader>
          <CardTitle>Recent Usage</CardTitle>
          <CardDescription>Detailed token usage records</CardDescription>
        </CardHeader>
        <CardContent>
          {usage?.usage && usage.usage.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b">
                    <th className="text-left p-2">Date</th>
                    <th className="text-left p-2">Operation</th>
                    <th className="text-left p-2">Provider</th>
                    <th className="text-left p-2">Model</th>
                    <th className="text-right p-2">Tokens</th>
                    <th className="text-right p-2">Cost</th>
                  </tr>
                </thead>
                <tbody>
                  {usage.usage.slice(0, 20).map((item) => (
                    <tr key={item.id} className="border-b">
                      <td className="p-2 text-muted-foreground">
                        {new Date(item.createdAt).toLocaleDateString()}
                      </td>
                      <td className="p-2 capitalize">{item.operation}</td>
                      <td className="p-2 capitalize">{item.provider}</td>
                      <td className="p-2">{item.model}</td>
                      <td className="p-2 text-right">{formatNumber(item.totalTokens)}</td>
                      <td className="p-2 text-right font-semibold">
                        {formatCurrency(item.cost)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <p className="text-sm text-muted-foreground text-center py-8">
              No usage data available
            </p>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

