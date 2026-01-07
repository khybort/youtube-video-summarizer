import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { useSettings, useUpdateSettings } from '@/hooks/useSettings'
import { useToast } from '@/components/ui/toast-provider'
import { settingsService } from '@/services/settingsService'
import { Loader2, Save, AlertCircle } from 'lucide-react'
import { useState, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'

export function Settings() {
  const { data: settings, isLoading } = useSettings()
  const updateMutation = useUpdateSettings()
  const { showToast } = useToast()
  
  // Check local whisper health
  const { data: localWhisperHealth } = useQuery({
    queryKey: ['localWhisperHealth'],
    queryFn: () => settingsService.checkLocalWhisperHealth(),
    refetchInterval: 10000, // Check every 10 seconds
  })
  
  const isLocalWhisperAvailable = localWhisperHealth?.available ?? false
  
  const [formData, setFormData] = useState({
    transcript_provider: 'youtube' as 'youtube' | 'groq' | 'local' | 'huggingface',
    summary_provider: 'gemini' as 'gemini' | 'ollama',
    embedding_provider: 'gemini' as 'gemini' | 'ollama',
    audio_analysis_provider: 'gemini' as 'gemini' | 'ollama',
    ollama_model: 'llama3.2',
    whisper_model: 'base',
    gemini_model: '',
    ollama_url: 'http://localhost:11434',
    local_whisper_url: 'http://localhost:8001',
    summary_language: 'auto',
  })
  const [groqApiKey, setGroqApiKey] = useState('')
  const [hfApiKey, setHfApiKey] = useState('')
  const [geminiApiKey, setGeminiApiKey] = useState('')
  const [clearGroqKey, setClearGroqKey] = useState(false)
  const [clearHfKey, setClearHfKey] = useState(false)
  const [clearGeminiKey, setClearGeminiKey] = useState(false)

  // Initialize from loaded settings
  useEffect(() => {
    if (settings) {
      setFormData({
        transcript_provider: settings.transcript_provider || 'youtube',
        summary_provider: settings.summary_provider || 'gemini',
        embedding_provider: settings.embedding_provider || 'gemini',
        audio_analysis_provider: settings.audio_analysis_provider || 'gemini',
        ollama_model: settings.ollama_model || 'llama3.2',
        whisper_model: settings.whisper_model || 'base',
        gemini_model: settings.gemini_model || '',
        ollama_url: settings.ollama_url || 'http://localhost:11434',
        local_whisper_url: settings.local_whisper_url || 'http://localhost:8001',
        summary_language: settings.summary_language || 'auto',
      })
      // Do not prefill API keys (backend doesn't return them). Reset any pending key edits.
      setGroqApiKey('')
      setHfApiKey('')
      setGeminiApiKey('')
      setClearGroqKey(false)
      setClearHfKey(false)
      setClearGeminiKey(false)
    }
  }, [settings])

  const handleSave = () => {
    // Check if user is trying to select local whisper when it's not available
    if (formData.transcript_provider === 'local' && !isLocalWhisperAvailable) {
      showToast('Local Whisper service is not available. Please ensure the service is running or select a different provider.', 'error')
      return
    }
    
    const payload: any = { ...formData }
    // Only send API keys if user explicitly provided one, or explicitly cleared it.
    if (geminiApiKey.trim() !== '' || clearGeminiKey) payload.gemini_api_key = clearGeminiKey ? '' : geminiApiKey.trim()
    if (groqApiKey.trim() !== '' || clearGroqKey) payload.groq_api_key = clearGroqKey ? '' : groqApiKey.trim()
    if (hfApiKey.trim() !== '' || clearHfKey) payload.huggingface_api_key = clearHfKey ? '' : hfApiKey.trim()

    updateMutation.mutate(payload, {
      onSuccess: () => {
        showToast('Settings saved successfully', 'success')
      },
      onError: (error: any) => {
        const errorMessage = error?.response?.data?.error || error?.message || 'Failed to save settings'
        if (error?.response?.data?.code === 'LOCAL_WHISPER_UNAVAILABLE') {
          showToast('Local Whisper service is not available. Please ensure the service is running or select a different provider.', 'error')
        } else {
          showToast(errorMessage, 'error')
        }
      },
    })
  }

  if (isLoading) {
    return (
      <div className="text-center py-12">
        <Loader2 className="w-6 h-6 animate-spin mx-auto mb-2 text-muted-foreground" />
        <p className="text-muted-foreground">Loading settings...</p>
      </div>
    )
  }

  return (
    <div className="space-y-6 max-w-4xl">
      <div>
        <h1 className="text-3xl font-bold">Settings</h1>
        <p className="text-muted-foreground mt-2">
          Configure model selection for different operations
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Model Selection</CardTitle>
          <CardDescription>
            Choose which models to use for each operation
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Transcript Provider */}
          <div className="space-y-2">
            <label className="text-sm font-medium">Transcript Provider</label>
            <Select
              value={formData.transcript_provider}
              onChange={(e) => {
                const newProvider = e.target.value as 'youtube' | 'groq' | 'local' | 'huggingface'
                if (newProvider === 'local' && !isLocalWhisperAvailable) {
                  showToast('Local Whisper service is not available. Please ensure the service is running.', 'error')
                  return
                }
                setFormData({ ...formData, transcript_provider: newProvider })
              }}
              disabled={formData.transcript_provider === 'local' && !isLocalWhisperAvailable}
            >
              <option value="youtube">YouTube Captions (Recommended)</option>
              <option value="groq">Groq Whisper (Cloud)</option>
              <option value="huggingface">Hugging Face Whisper (Cloud)</option>
              <option value="local" disabled={!isLocalWhisperAvailable}>
                Local Whisper {!isLocalWhisperAvailable ? '(Unavailable)' : ''}
              </option>
            </Select>
            {formData.transcript_provider === 'local' && !isLocalWhisperAvailable && (
              <div className="flex items-center gap-2 text-sm text-destructive bg-destructive/10 p-2 rounded">
                <AlertCircle className="w-4 h-4" />
                <span>Local Whisper service is not available. Please ensure the service is running at {localWhisperHealth?.url || 'http://localhost:8001'}.</span>
              </div>
            )}
            <p className="text-xs text-muted-foreground">
              YouTube captions are free and fast. Groq and Hugging Face are cloud-based options. Local requires setup.
            </p>
          </div>

          {/* Cloud Whisper Keys */}
          {(formData.transcript_provider === 'groq' || formData.transcript_provider === 'huggingface') && (
            <div className="space-y-4 p-4 border rounded-lg bg-muted/50">
              <h3 className="text-sm font-semibold">Cloud Whisper Configuration</h3>

              {formData.transcript_provider === 'groq' && (
                <div className="space-y-2">
                  <label className="text-sm font-medium">Groq API Key</label>
                  <Input
                    value={groqApiKey}
                    onChange={(e) => {
                      setGroqApiKey(e.target.value)
                      if (e.target.value.trim() !== '') setClearGroqKey(false)
                    }}
                    placeholder={settings?.has_groq_api_key ? 'Configured (enter to replace)' : 'Enter Groq API key (optional)'}
                  />
                  <div className="flex items-center gap-3">
                    <Button
                      type="button"
                      variant="secondary"
                      onClick={() => {
                        setGroqApiKey('')
                        setClearGroqKey(true)
                      }}
                    >
                      Clear stored key
                    </Button>
                    <p className="text-xs text-muted-foreground">
                      Leave empty to keep current / use env. Backend does not display the key.
                    </p>
                  </div>
                </div>
              )}

              {formData.transcript_provider === 'huggingface' && (
                <div className="space-y-2">
                  <label className="text-sm font-medium">Hugging Face API Key</label>
                  <Input
                    value={hfApiKey}
                    onChange={(e) => {
                      setHfApiKey(e.target.value)
                      if (e.target.value.trim() !== '') setClearHfKey(false)
                    }}
                    placeholder={settings?.has_huggingface_api_key ? 'Configured (enter to replace)' : 'Enter Hugging Face API key (optional)'}
                  />
                  <div className="flex items-center gap-3">
                    <Button
                      type="button"
                      variant="secondary"
                      onClick={() => {
                        setHfApiKey('')
                        setClearHfKey(true)
                      }}
                    >
                      Clear stored key
                    </Button>
                    <p className="text-xs text-muted-foreground">
                      Leave empty to keep current / use env. Backend does not display the key.
                    </p>
                  </div>
                </div>
              )}
            </div>
          )}

          {/* Summary Provider */}
          <div className="space-y-2">
            <label className="text-sm font-medium">Summary Provider</label>
            <Select
              value={formData.summary_provider}
              onChange={(e) =>
                setFormData({ ...formData, summary_provider: e.target.value as 'gemini' | 'ollama' })
              }
            >
              <option value="gemini">Google Gemini (Cloud)</option>
              <option value="ollama">Ollama (Local)</option>
            </Select>
            <p className="text-xs text-muted-foreground">
              Model used for generating video summaries
            </p>
          </div>

          {/* Gemini API Key */}
          {(formData.summary_provider === 'gemini' ||
            formData.embedding_provider === 'gemini' ||
            formData.audio_analysis_provider === 'gemini') && (
            <div className="space-y-4 p-4 border rounded-lg bg-muted/50">
              <h3 className="text-sm font-semibold">Gemini API Configuration</h3>
              <div className="space-y-2">
                <label className="text-sm font-medium">Gemini API Key</label>
                <Input
                  value={geminiApiKey}
                  onChange={(e) => {
                    setGeminiApiKey(e.target.value)
                    if (e.target.value.trim() !== '') setClearGeminiKey(false)
                  }}
                  placeholder={settings?.has_gemini_api_key ? 'Configured (enter to replace)' : 'Enter Gemini API key (optional)'}
                />
                <div className="flex items-center gap-3">
                  <Button
                    type="button"
                    variant="secondary"
                    onClick={() => {
                      setGeminiApiKey('')
                      setClearGeminiKey(true)
                    }}
                  >
                    Clear stored key
                  </Button>
                  <p className="text-xs text-muted-foreground">
                    Leave empty to keep current / use env. Backend does not display the key.
                  </p>
                </div>
              </div>
            </div>
          )}

          {/* Summary Language */}
          <div className="space-y-2">
            <label className="text-sm font-medium">Summary Language</label>
            <Select
              value={formData.summary_language}
              onChange={(e) =>
                setFormData({ ...formData, summary_language: e.target.value })
              }
            >
              <option value="auto">Auto (Same as transcript)</option>
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
            <p className="text-xs text-muted-foreground">
              Language for generated summaries. "Auto" uses the same language as the transcript.
            </p>
          </div>

          {/* Embedding Provider */}
          <div className="space-y-2">
            <label className="text-sm font-medium">Embedding Provider</label>
            <Select
              value={formData.embedding_provider}
              onChange={(e) =>
                setFormData({ ...formData, embedding_provider: e.target.value as 'gemini' | 'ollama' })
              }
            >
              <option value="gemini">Google Gemini (Cloud)</option>
              <option value="ollama">Ollama (Local)</option>
            </Select>
            <p className="text-xs text-muted-foreground">
              Model used for generating video embeddings (for similarity search)
            </p>
          </div>

          {/* Audio Analysis Provider */}
          <div className="space-y-2">
            <label className="text-sm font-medium">Audio Analysis Provider</label>
            <Select
              value={formData.audio_analysis_provider}
              onChange={(e) =>
                setFormData({ ...formData, audio_analysis_provider: e.target.value as 'gemini' | 'ollama' })
              }
            >
              <option value="gemini">Google Gemini (Cloud)</option>
              <option value="ollama">Ollama (Local)</option>
            </Select>
            <p className="text-xs text-muted-foreground">
              Model used for audio analysis (future feature)
            </p>
          </div>

          {/* Gemini Settings */}
          {(formData.summary_provider === 'gemini' || 
            formData.embedding_provider === 'gemini' || 
            formData.audio_analysis_provider === 'gemini') && (
            <div className="space-y-4 p-4 border rounded-lg bg-muted/50">
              <h3 className="text-sm font-semibold">Gemini Configuration</h3>
              
              <div className="space-y-2">
                <label className="text-sm font-medium">Gemini Model</label>
                <Input
                  value={formData.gemini_model}
                  onChange={(e) => setFormData({ ...formData, gemini_model: e.target.value })}
                  placeholder="Auto-detect (leave empty)"
                />
                <p className="text-xs text-muted-foreground">
                  Model name (e.g., gemini-1.5-flash, gemini-1.5-pro, gemini-pro). Leave empty to auto-detect from API.
                </p>
              </div>
            </div>
          )}

          {/* Ollama Settings */}
          {(formData.summary_provider === 'ollama' || 
            formData.embedding_provider === 'ollama' || 
            formData.audio_analysis_provider === 'ollama') && (
            <div className="space-y-4 p-4 border rounded-lg bg-muted/50">
              <h3 className="text-sm font-semibold">Ollama Configuration</h3>
              
              <div className="space-y-2">
                <label className="text-sm font-medium">Ollama Model</label>
                <Input
                  value={formData.ollama_model}
                  onChange={(e) => setFormData({ ...formData, ollama_model: e.target.value })}
                  placeholder="llama3.2"
                />
                <p className="text-xs text-muted-foreground">
                  Model name (e.g., llama3.2, llama3.1, mistral)
                </p>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium">Ollama URL</label>
                <Input
                  value={formData.ollama_url}
                  onChange={(e) => setFormData({ ...formData, ollama_url: e.target.value })}
                  placeholder="http://localhost:11434"
                />
                <p className="text-xs text-muted-foreground">
                  URL where Ollama is running
                </p>
              </div>
            </div>
          )}

          {/* Whisper Settings */}
          {(formData.transcript_provider === 'local') && (
            <div className="space-y-4 p-4 border rounded-lg bg-muted/50">
              <h3 className="text-sm font-semibold">Local Whisper Configuration</h3>
              
              <div className="space-y-2">
                <label className="text-sm font-medium">Whisper Model</label>
                <Select
                  value={formData.whisper_model}
                  onChange={(e) =>
                    setFormData({ ...formData, whisper_model: e.target.value })
                  }
                >
                  <option value="base">Base</option>
                  <option value="small">Small</option>
                  <option value="medium">Medium</option>
                  <option value="large-v3">Large v3</option>
                </Select>
                <p className="text-xs text-muted-foreground">
                  Larger models are more accurate but slower
                </p>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium">Local Whisper URL</label>
                <Input
                  value={formData.local_whisper_url}
                  onChange={(e) => setFormData({ ...formData, local_whisper_url: e.target.value })}
                  placeholder="http://localhost:8001"
                />
                <p className="text-xs text-muted-foreground">
                  URL where local Whisper service is running
                </p>
              </div>
            </div>
          )}

          <Button
            onClick={handleSave}
            disabled={updateMutation.isPending}
            className="w-full"
          >
            {updateMutation.isPending ? (
              <>
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                Saving...
              </>
            ) : (
              <>
                <Save className="w-4 h-4 mr-2" />
                Save Changes
              </>
            )}
          </Button>
        </CardContent>
      </Card>
    </div>
  )
}
