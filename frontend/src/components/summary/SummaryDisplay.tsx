import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Select } from '@/components/ui/select'
import { useSummary, useSummarizeVideo } from '@/hooks/useSummary'
import { useToast } from '@/components/ui/toast-provider'
import { videoService } from '@/services/videoService'
import { Loader2, FileText, Download, Volume2, RefreshCw } from 'lucide-react'
import { useState } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'

interface SummaryDisplayProps {
  videoId: string
}

export function SummaryDisplay({ videoId }: SummaryDisplayProps) {
  const [summaryLanguage, setSummaryLanguage] = useState<string>('auto')
  const { data: summary, isLoading, refetch } = useSummary(videoId, summaryLanguage)
  const summarizeMutation = useSummarizeVideo()
  const { showToast } = useToast()
  const [summaryType, setSummaryType] = useState<'short' | 'detailed' | 'bullet_points'>('short')
  const [fromAudio, setFromAudio] = useState(false)
  const [isRegenerating, setIsRegenerating] = useState(false)
  const [isTranslating, setIsTranslating] = useState(false)

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
      {/* Loading overlay when regenerating or translating */}
      {isProcessing && summary && (
        <div className="absolute inset-0 bg-background/80 backdrop-blur-sm z-10 rounded-lg flex items-center justify-center">
          <div className="text-center space-y-2">
            <Loader2 className="w-8 h-8 animate-spin mx-auto text-primary" />
            <p className="text-sm font-medium">
              {isTranslating ? 'Translating summary...' : fromAudio ? 'Regenerating from audio...' : 'Regenerating summary...'}
            </p>
            <p className="text-xs text-muted-foreground">
              {isTranslating ? 'This will only take a moment' : 'This may take a few moments'}
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
      <CardContent className="space-y-6">
        {summary && (
          <div className={`relative ${isProcessing ? 'opacity-50' : ''}`}>
            {/* Render Markdown content with professional styling */}
            <div className="prose prose-slate dark:prose-invert max-w-none 
                          prose-headings:font-semibold prose-headings:text-foreground
                          prose-h2:text-xl prose-h2:mt-8 prose-h2:mb-4 prose-h2:border-b prose-h2:pb-2 prose-h2:border-border
                          prose-h3:text-lg prose-h3:mt-6 prose-h3:mb-3
                          prose-p:text-base prose-p:leading-7 prose-p:text-foreground prose-p:my-4
                          prose-ul:my-4 prose-ul:space-y-2
                          prose-li:text-base prose-li:text-foreground prose-li:leading-7
                          prose-strong:text-foreground prose-strong:font-semibold
                          prose-code:text-sm prose-code:bg-muted prose-code:px-1.5 prose-code:py-0.5 prose-code:rounded
                          prose-blockquote:border-l-4 prose-blockquote:border-primary prose-blockquote:pl-4 prose-blockquote:italic
                          prose-a:text-primary prose-a:underline hover:prose-a:text-primary/80
                          prose-pre:bg-muted prose-pre:p-4 prose-pre:rounded-lg prose-pre:overflow-x-auto">
              <ReactMarkdown
                remarkPlugins={[remarkGfm]}
                components={{
                  h2: ({node, ...props}) => <h2 className="font-semibold text-xl mt-8 mb-4 pb-2 border-b border-border" {...props} />,
                  h3: ({node, ...props}) => <h3 className="font-semibold text-lg mt-6 mb-3" {...props} />,
                  p: ({node, ...props}) => <p className="text-base leading-7 text-foreground my-4" {...props} />,
                  ul: ({node, ...props}) => <ul className="my-4 space-y-2 list-disc list-inside" {...props} />,
                  li: ({node, ...props}) => <li className="text-base text-foreground leading-7" {...props} />,
                  strong: ({node, ...props}) => <strong className="font-semibold text-foreground" {...props} />,
                  code: ({node, inline, ...props}: any) => 
                    inline ? (
                      <code className="text-sm bg-muted px-1.5 py-0.5 rounded" {...props} />
                    ) : (
                      <code className="block text-sm bg-muted p-4 rounded-lg overflow-x-auto" {...props} />
                    ),
                  blockquote: ({node, ...props}) => (
                    <blockquote className="border-l-4 border-primary pl-4 italic my-4" {...props} />
                  ),
                }}
              >
                {summary.content}
              </ReactMarkdown>
            </div>

            {/* Fallback: If keyPoints exist but weren't in markdown, show them separately */}
            {summary.keyPoints && summary.keyPoints.length > 0 && 
             !summary.content.includes('## Key Points') && 
             !summary.content.includes('## Key Takeaways') && (
              <div className="mt-6 pt-6 border-t border-border">
                <h3 className="font-semibold text-lg mb-3">Key Points</h3>
                <ul className="space-y-2 list-disc list-inside">
                  {summary.keyPoints.map((point, idx) => (
                    <li key={idx} className="text-base text-foreground leading-7">{point}</li>
                  ))}
                </ul>
              </div>
            )}
          </div>
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
              onChange={async (e) => {
                const newLanguage = e.target.value
                setSummaryLanguage(newLanguage)
                
                // If summary exists and language is not "auto", translate it
                if (summary && newLanguage !== 'auto') {
                  setIsTranslating(true)
                  try {
                    await videoService.translateSummary(videoId, newLanguage)
                    refetch()
                    showToast('Summary translated successfully', 'success')
                  } catch (error: any) {
                    showToast(error?.response?.data?.error || error?.message || 'Failed to translate summary', 'error')
                  } finally {
                    setIsTranslating(false)
                  }
                }
              }}
              disabled={isProcessing || isTranslating}
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

