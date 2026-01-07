import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Select } from '@/components/ui/select'
import { useSummary, useSummarizeVideo } from '@/hooks/useSummary'
import { useToast } from '@/components/ui/toast-provider'
import { Loader2, FileText, Download, Volume2, RefreshCw } from 'lucide-react'
import { useState } from 'react'

interface SummaryDisplayProps {
  videoId: string
}

export function SummaryDisplay({ videoId }: SummaryDisplayProps) {
  const [summaryLanguage, setSummaryLanguage] = useState<string>('auto')
  const { data: summary, isLoading, refetch } = useSummary(videoId)
  const summarizeMutation = useSummarizeVideo()
  const { showToast } = useToast()
  const [summaryType, setSummaryType] = useState<'short' | 'detailed' | 'bullet_points'>('short')
  const [fromAudio, setFromAudio] = useState(false)
  const [isRegenerating, setIsRegenerating] = useState(false)

  const handleSummarize = () => {
    const isRegenerate = !!summary
    setIsRegenerating(isRegenerate)
    
    summarizeMutation.mutate(
      { videoId, type: summaryType, fromAudio, language: summaryLanguage },
      {
        onSuccess: () => {
          setIsRegenerating(false)
          refetch()
          if (isRegenerate) {
            showToast(
              fromAudio 
                ? 'Summary regenerated from audio successfully' 
                : 'Summary regenerated successfully',
              'success'
            )
          } else {
            showToast('Summary generated successfully', 'success')
          }
        },
        onError: (error: any) => {
          setIsRegenerating(false)
          const errorMessage = error?.response?.data?.error || error?.message || 'Failed to generate summary'
          showToast(errorMessage, 'error')
        },
      }
    )
  }

  const handleExport = () => {
    if (!summary) return

    const content = `Summary for Video ${videoId}\n\n${summary.content}\n\nKey Points:\n${summary.keyPoints.map((p) => `- ${p}`).join('\n')}`
    const blob = new Blob([content], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `summary-${videoId}.txt`
    a.click()
    URL.revokeObjectURL(url)
  }

  if (isLoading) {
    return (
      <Card>
        <CardContent className="py-8 text-center">
          <Loader2 className="w-6 h-6 animate-spin mx-auto mb-2 text-muted-foreground" />
          <p className="text-sm text-muted-foreground">Loading summary...</p>
        </CardContent>
      </Card>
    )
  }

  if (!summary) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Summary</CardTitle>
          <CardDescription>Generate an AI-powered summary of this video</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-3">
            <div className="grid grid-cols-2 gap-2">
              <select
                value={summaryType}
                onChange={(e) => setSummaryType(e.target.value as any)}
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              >
                <option value="short">Short Summary</option>
                <option value="detailed">Detailed Summary</option>
                <option value="bullet_points">Bullet Points</option>
              </select>
              <Select
                value={summaryLanguage}
                onChange={(e) => setSummaryLanguage(e.target.value)}
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              >
                <option value="auto">Auto (Transcript Language)</option>
                <option value="en">English</option>
                <option value="tr">Turkish</option>
                <option value="es">Spanish</option>
                <option value="fr">French</option>
                <option value="de">German</option>
                <option value="it">Italian</option>
                <option value="pt">Portuguese</option>
                <option value="ru">Russian</option>
                <option value="ja">Japanese</option>
                <option value="ko">Korean</option>
                <option value="zh">Chinese</option>
                <option value="ar">Arabic</option>
                <option value="hi">Hindi</option>
                <option value="nl">Dutch</option>
                <option value="pl">Polish</option>
                <option value="sv">Swedish</option>
                <option value="da">Danish</option>
                <option value="no">Norwegian</option>
                <option value="fi">Finnish</option>
              </Select>
            </div>
            
            <div className="flex items-center gap-2">
              <input
                type="checkbox"
                id="fromAudio"
                checked={fromAudio}
                onChange={(e) => setFromAudio(e.target.checked)}
                className="h-4 w-4 rounded border-gray-300"
              />
              <label htmlFor="fromAudio" className="text-sm font-medium cursor-pointer flex items-center gap-2">
                <Volume2 className="w-4 h-4" />
                Generate from Audio (uses Whisper from settings)
              </label>
            </div>
            <p className="text-xs text-muted-foreground">
              {fromAudio 
                ? "Summary will be generated from audio using the Whisper provider selected in settings (local/groq)"
                : "Summary will be generated from transcript (YouTube captions or Whisper)"}
            </p>
            
            <Button
              onClick={handleSummarize}
              disabled={summarizeMutation.isPending}
              className="w-full"
            >
              {summarizeMutation.isPending ? (
                <>
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  Generating...
                </>
              ) : (
                <>
                  {fromAudio ? (
                    <>
                      <Volume2 className="w-4 h-4 mr-2" />
                      Generate from Audio
                    </>
                  ) : (
                    <>
                      <FileText className="w-4 h-4 mr-2" />
                      Generate Summary
                    </>
                  )}
                </>
              )}
            </Button>
          </div>
        </CardContent>
      </Card>
    )
  }

  const isProcessing = summarizeMutation.isPending || isRegenerating

  return (
    <Card className="relative">
      {/* Loading overlay when regenerating */}
      {isProcessing && summary && (
        <div className="absolute inset-0 bg-background/80 backdrop-blur-sm z-10 rounded-lg flex items-center justify-center">
          <div className="text-center space-y-2">
            <Loader2 className="w-8 h-8 animate-spin mx-auto text-primary" />
            <p className="text-sm font-medium">
              {fromAudio ? 'Regenerating from audio...' : 'Regenerating summary...'}
            </p>
            <p className="text-xs text-muted-foreground">
              This may take a few moments
            </p>
          </div>
        </div>
      )}

      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2">
              Summary
              {summary && summary.createdAt && (
                <span className="text-xs font-normal text-muted-foreground">
                  • Last updated: {new Date(summary.createdAt).toLocaleString()}
                </span>
              )}
            </CardTitle>
            <CardDescription>
              {summary?.summaryType || summaryType} • {summary?.modelUsed || 'N/A'}
            </CardDescription>
          </div>
          <div className="flex gap-2">
            {summary && (
              <Button 
                variant="outline" 
                size="sm" 
                onClick={handleExport}
                disabled={isProcessing}
              >
                <Download className="w-4 h-4 mr-2" />
                Export
              </Button>
            )}
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {summary && (
          <>
            <div className="prose prose-sm max-w-none dark:prose-invert relative">
              <div className={`whitespace-pre-wrap text-sm ${isProcessing ? 'opacity-50' : ''}`}>
                {summary.content}
              </div>
            </div>

            {summary.keyPoints && summary.keyPoints.length > 0 && (
              <div className={isProcessing ? 'opacity-50' : ''}>
                <h4 className="font-semibold mb-2">Key Points</h4>
                <ul className="list-disc list-inside space-y-1 text-sm">
                  {summary.keyPoints.map((point, idx) => (
                    <li key={idx}>{point}</li>
                  ))}
                </ul>
              </div>
            )}
          </>
        )}

        <div className="pt-4 border-t space-y-3">
          <div className="grid grid-cols-2 gap-2">
            <select
              value={summaryType}
              onChange={(e) => setSummaryType(e.target.value as any)}
              disabled={isProcessing}
              className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm disabled:opacity-50"
            >
              <option value="short">Short Summary</option>
              <option value="detailed">Detailed Summary</option>
              <option value="bullet_points">Bullet Points</option>
            </select>
            <Select
              value={summaryLanguage}
              onChange={(e) => setSummaryLanguage(e.target.value)}
              disabled={isProcessing}
            >
              <option value="auto">Auto (Transcript Language)</option>
              <option value="en">English</option>
              <option value="tr">Turkish</option>
              <option value="es">Spanish</option>
              <option value="fr">French</option>
              <option value="de">German</option>
              <option value="it">Italian</option>
              <option value="pt">Portuguese</option>
              <option value="ru">Russian</option>
              <option value="ja">Japanese</option>
              <option value="ko">Korean</option>
              <option value="zh">Chinese</option>
              <option value="ar">Arabic</option>
              <option value="hi">Hindi</option>
              <option value="nl">Dutch</option>
              <option value="pl">Polish</option>
              <option value="sv">Swedish</option>
              <option value="da">Danish</option>
              <option value="no">Norwegian</option>
              <option value="fi">Finnish</option>
            </Select>
          </div>
          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="fromAudioRegenerate"
              checked={fromAudio}
              onChange={(e) => setFromAudio(e.target.checked)}
              disabled={isProcessing}
              className="h-4 w-4 rounded border-gray-300 disabled:opacity-50"
            />
            <label 
              htmlFor="fromAudioRegenerate" 
              className={`text-sm font-medium cursor-pointer flex items-center gap-2 ${isProcessing ? 'opacity-50 cursor-not-allowed' : ''}`}
            >
              <Volume2 className="w-4 h-4" />
              {summary 
                ? 'Regenerate from Audio (uses Whisper from settings)' 
                : 'Generate from Audio (uses Whisper from settings)'}
            </label>
          </div>
          <div className="flex gap-2">
            <Button
              variant={summary ? "outline" : "default"}
              size="sm"
              onClick={handleSummarize}
              disabled={isProcessing}
              className="flex-1"
            >
              {isProcessing ? (
                <>
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  {summary ? 'Regenerating...' : 'Generating...'}
                </>
              ) : (
                <>
                  {summary ? (
                    <>
                      <RefreshCw className="w-4 h-4 mr-2" />
                      {fromAudio ? 'Regenerate from Audio' : 'Regenerate Summary'}
                    </>
                  ) : (
                    <>
                      {fromAudio ? (
                        <>
                          <Volume2 className="w-4 h-4 mr-2" />
                          Generate from Audio
                        </>
                      ) : (
                        <>
                          <FileText className="w-4 h-4 mr-2" />
                          Generate Summary
                        </>
                      )}
                    </>
                  )}
                </>
              )}
            </Button>
            {summary && !fromAudio && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  setFromAudio(true)
                  handleSummarize()
                }}
                disabled={isProcessing}
                title="Quick regenerate from audio"
              >
                <Volume2 className="w-4 h-4" />
              </Button>
            )}
          </div>
          {summary && (
            <p className="text-xs text-muted-foreground">
              {fromAudio 
                ? '⚠️ Regenerating will replace the current summary with a new one generated from audio.'
                : '⚠️ Regenerating will replace the current summary with a new one.'}
            </p>
          )}
        </div>
      </CardContent>
    </Card>
  )
}

