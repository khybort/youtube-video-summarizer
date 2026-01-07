import { Link, useLocation } from 'react-router-dom'
import { Home, Video, Settings, Search, DollarSign, Sparkles } from 'lucide-react'
import { cn } from '@/lib/utils'

const navigation = [
  { name: 'Dashboard', href: '/', icon: Home, description: 'Overview & quick actions' },
  { name: 'Videos', href: '/videos', icon: Video, description: 'Browse all videos' },
  { name: 'Search', href: '/search', icon: Search, description: 'Find videos' },
  { name: 'Cost Analysis', href: '/costs', icon: DollarSign, description: 'Usage & costs' },
  { name: 'Settings', href: '/settings', icon: Settings, description: 'Configure models' },
]

export function Sidebar() {
  const location = useLocation()

  return (
    <div className="w-72 bg-card border-r border-border h-screen sticky top-0 flex flex-col">
      {/* Logo Section */}
      <div className="p-6 border-b border-border">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-lg bg-gradient-to-br from-primary to-primary/60 flex items-center justify-center">
            <Sparkles className="w-6 h-6 text-primary-foreground" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-foreground">YouTube Analyzer</h1>
            <p className="text-xs text-muted-foreground">AI Video Summarizer</p>
          </div>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 px-4 py-6 space-y-1 overflow-y-auto">
        {navigation.map((item) => {
          const isActive = location.pathname === item.href || 
            (item.href !== '/' && location.pathname.startsWith(item.href))
          return (
            <Link
              key={item.name}
              to={item.href}
              className={cn(
                "flex items-start gap-3 px-4 py-3 rounded-xl text-sm font-medium transition-all duration-200 group",
                isActive
                  ? "bg-primary text-primary-foreground shadow-md shadow-primary/20"
                  : "text-muted-foreground hover:bg-accent hover:text-accent-foreground"
              )}
            >
              <item.icon className={cn(
                "w-5 h-5 mt-0.5 flex-shrink-0 transition-transform",
                isActive ? "scale-110" : "group-hover:scale-105"
              )} />
              <div className="flex-1 min-w-0">
                <div className="font-semibold">{item.name}</div>
                <div className={cn(
                  "text-xs mt-0.5 line-clamp-1",
                  isActive ? "text-primary-foreground/80" : "text-muted-foreground"
                )}>
                  {item.description}
                </div>
              </div>
            </Link>
          )
        })}
      </nav>

      {/* Footer */}
      <div className="p-4 border-t border-border">
        <div className="text-xs text-muted-foreground text-center">
          <p className="font-medium">Powered by AI</p>
          <p className="mt-1">Gemini • Groq • Ollama</p>
        </div>
      </div>
    </div>
  )
}
