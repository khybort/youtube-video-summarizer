import { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { useTranscript, useAvailableLanguages } from '@/hooks/useTranscript'
import { Search, Copy, Check, FileText, Loader2, AlertCircle, Languages } from 'lucide-react'
import { Alert, AlertDescription } from '@/components/ui/alert'

interface TranscriptViewerProps {
  videoId: string
  hasTranscript?: boolean
  onSeek?: (time: number) => void
}

export function TranscriptViewer({ videoId, hasTranscript, onSeek }: TranscriptViewerProps) {
  const [selectedLanguage, setSelectedLanguage] = useState<string>('')
  const { data: transcript, isLoading, error, refetch } = useTranscript(videoId, hasTranscript, selectedLanguage || undefined)
  const { data: availableLanguages, isLoading: isLoadingLanguages } = useAvailableLanguages(videoId)
  const [searchTerm, setSearchTerm] = useState('')
  const [activeSegment, setActiveSegment] = useState<number | null>(null)
  const [copied, setCopied] = useState(false)

  // Set initial language from transcript if available
  useEffect(() => {
    if (transcript?.language && !selectedLanguage) {
      setSelectedLanguage(transcript.language)
    }
  }, [transcript, selectedLanguage])

  // When language changes, refetch transcript
  useEffect(() => {
    if (selectedLanguage) {
      refetch()
    }
  }, [selectedLanguage, refetch])

  // If video doesn't have transcript, show message immediately
  if (hasTranscript === false) {
    return (
      <Card>
        <CardContent className="py-12">
          <div className="flex flex-col items-center justify-center text-center space-y-4">
            <div className="rounded-full bg-muted p-4">
              <FileText className="w-8 h-8 text-muted-foreground" />
            </div>
            <div className="space-y-2">
              <h3 className="text-lg font-semibold">Transcript Mevcut Değil</h3>
              <p className="text-sm text-muted-foreground max-w-md">
                Bu video için transcript henüz oluşturulmamış. Video analiz edildiğinde transcript otomatik olarak oluşturulacaktır.
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
    )
  }

  if (isLoading) {
    return (
      <Card>
        <CardContent className="py-12">
          <div className="flex flex-col items-center justify-center text-center space-y-4">
            <Loader2 className="w-8 h-8 animate-spin text-primary" />
            <p className="text-sm text-muted-foreground">Transcript yükleniyor...</p>
          </div>
        </CardContent>
      </Card>
    )
  }

  if (error) {
    return (
      <Card>
        <CardContent className="py-12">
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              Transcript yüklenirken bir hata oluştu. Lütfen tekrar deneyin.
            </AlertDescription>
          </Alert>
        </CardContent>
      </Card>
    )
  }

  if (!transcript) {
    return (
      <Card>
        <CardContent className="py-12">
          <div className="flex flex-col items-center justify-center text-center space-y-4">
            <div className="rounded-full bg-muted p-4">
              <FileText className="w-8 h-8 text-muted-foreground" />
            </div>
            <div className="space-y-2">
              <h3 className="text-lg font-semibold">Transcript Mevcut Değil</h3>
              <p className="text-sm text-muted-foreground max-w-md">
                Bu video için transcript henüz oluşturulmamış. Video analiz edildiğinde transcript otomatik olarak oluşturulacaktır.
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
    )
  }

  const segments = transcript.segments || []
  const filteredSegments = segments.filter((seg) =>
    seg.text.toLowerCase().includes(searchTerm.toLowerCase())
  )

  const handleCopy = async () => {
    const fullText = segments.map((s) => s.text).join(' ')
    await navigator.clipboard.writeText(fullText)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const formatTime = (seconds: number): string => {
    const mins = Math.floor(seconds / 60)
    const secs = Math.floor(seconds % 60)
    return `${mins}:${secs.toString().padStart(2, '0')}`
  }

  const hasSegments = segments.length > 0
  const fullText = segments.map((s) => s.text).join(' ')

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div className="flex-1">
            <CardTitle className="flex items-center gap-2">
              <FileText className="w-5 h-5" />
              Transcript
            </CardTitle>
            <CardDescription className="mt-1">
              {transcript.language?.toUpperCase() || 'N/A'} • {transcript.source === 'youtube' ? 'YouTube Captions' : transcript.source === 'whisper' ? 'Whisper Transcription' : transcript.source}
              {hasSegments && ` • ${segments.length} segment`}
            </CardDescription>
          </div>
          <div className="flex items-center gap-2">
            {availableLanguages && availableLanguages.length > 1 && (
              <div className="flex items-center gap-2">
                <Languages className="w-4 h-4 text-muted-foreground" />
                <Select
                  value={selectedLanguage}
                  onChange={(e) => setSelectedLanguage(e.target.value)}
                  className="w-40"
                  disabled={isLoadingLanguages}
                >
                  <option value="">Otomatik</option>
                  {availableLanguages.map((lang) => (
                    <option key={lang.code} value={lang.code}>
                      {lang.name} {lang.is_auto_generated ? '(Otomatik)' : ''}
                    </option>
                  ))}
                </Select>
              </div>
            )}
            {hasSegments && (
              <Button variant="outline" size="sm" onClick={handleCopy}>
                {copied ? (
                  <>
                    <Check className="w-4 h-4 mr-2" />
                    Kopyalandı!
                  </>
                ) : (
                  <>
                    <Copy className="w-4 h-4 mr-2" />
                    Kopyala
                  </>
                )}
              </Button>
            )}
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {hasSegments ? (
          <>
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-muted-foreground" />
              <Input
                placeholder="Transcript'te ara..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-10"
              />
            </div>

            <div className="max-h-[600px] overflow-y-auto space-y-2 pr-2">
              {filteredSegments.map((segment, idx) => (
                <div
                  key={idx}
                  className={`p-4 rounded-lg border cursor-pointer transition-all ${
                    activeSegment === idx
                      ? 'bg-primary/10 border-primary shadow-sm'
                      : 'bg-card hover:bg-accent/50 border-border'
                  }`}
                  onClick={() => {
                    setActiveSegment(idx)
                    onSeek?.(segment.start)
                  }}
                >
                  <div className="flex items-start gap-4">
                    <button
                      className="text-xs font-mono text-muted-foreground hover:text-primary transition-colors px-2 py-1 rounded bg-muted/50 hover:bg-muted whitespace-nowrap"
                      onClick={(e) => {
                        e.stopPropagation()
                        onSeek?.(segment.start)
                      }}
                    >
                      {formatTime(segment.start)}
                    </button>
                    <p className="text-sm flex-1 leading-relaxed">{segment.text}</p>
                  </div>
                </div>
              ))}
            </div>

            {filteredSegments.length === 0 && searchTerm && (
              <div className="text-center py-8 text-muted-foreground">
                <p className="text-sm">"{searchTerm}" için sonuç bulunamadı</p>
              </div>
            )}
          </>
        ) : (
          <div className="py-8">
            <div className="text-center space-y-2">
              <p className="text-sm text-muted-foreground">
                Transcript içeriği mevcut ancak segment bilgisi bulunamadı.
              </p>
              {fullText && (
                <div className="mt-4 p-4 bg-muted rounded-lg">
                  <p className="text-sm whitespace-pre-wrap">{fullText}</p>
                </div>
              )}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

