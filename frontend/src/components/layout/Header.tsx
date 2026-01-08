import { Moon, Sun, Search as SearchIcon } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useState, useEffect, useRef } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'

export function Header() {
  const [isDark, setIsDark] = useState(() => 
    document.documentElement.classList.contains('dark')
  )
  const navigate = useNavigate()
  const location = useLocation()
  const debounceTimerRef = useRef<number | null>(null)
  
  // Get query from URL if on search page
  const urlParams = new URLSearchParams(location.search)
  const urlQuery = urlParams.get('q') || ''
  
  const [searchQuery, setSearchQuery] = useState(urlQuery)
  const isSearchPage = location.pathname === '/search'
  const isUserTypingRef = useRef(false)
  
  // Update search query when URL changes (e.g., when navigating to search page)
  // But skip if user is currently typing
  useEffect(() => {
    if (!isUserTypingRef.current) {
      const urlParams = new URLSearchParams(location.search)
      const urlQuery = urlParams.get('q') || ''
      if (urlQuery !== searchQuery) {
        setSearchQuery(urlQuery)
      }
    }
  }, [location.search])
  
  // Auto-search when on search page and query changes (with debounce)
  useEffect(() => {
    if (isSearchPage && isUserTypingRef.current) {
      // Clear previous timer
      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current)
      }
      
      // Set new timer for debounced search
      debounceTimerRef.current = setTimeout(() => {
        const trimmedQuery = searchQuery.trim()
        const urlParams = new URLSearchParams(location.search)
        const currentUrlQuery = urlParams.get('q') || ''
        
        // Only update if different
        if (trimmedQuery !== currentUrlQuery) {
          isUserTypingRef.current = false
          if (trimmedQuery) {
            navigate(`/search?q=${encodeURIComponent(trimmedQuery)}`, { replace: true })
          } else {
            navigate('/search', { replace: true })
          }
        } else {
          isUserTypingRef.current = false
        }
      }, 500) // 500ms debounce
      
      return () => {
        if (debounceTimerRef.current) {
          clearTimeout(debounceTimerRef.current)
        }
      }
    }
  }, [searchQuery, isSearchPage, navigate, location.search])

  useEffect(() => {
    // Listen for theme changes from other sources (e.g., system preference changes)
    const observer = new MutationObserver(() => {
      setIsDark(document.documentElement.classList.contains('dark'))
    })
    
    observer.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ['class']
    })

    return () => observer.disconnect()
  }, [])

  const toggleTheme = () => {
    const newIsDark = !isDark
    setIsDark(newIsDark)
    
    if (newIsDark) {
      document.documentElement.classList.add('dark')
      localStorage.setItem('theme', 'dark')
    } else {
      document.documentElement.classList.remove('dark')
      localStorage.setItem('theme', 'light')
    }
  }

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    if (searchQuery.trim()) {
      navigate(`/search?q=${encodeURIComponent(searchQuery.trim())}`)
    }
  }

  return (
    <header className="h-16 border-b border-border bg-card/95 backdrop-blur supports-[backdrop-filter]:bg-card/60 flex items-center justify-between px-6 sticky top-0 z-40">
      <div className="flex items-center gap-4 flex-1 w-full max-w-4xl mx-auto">
        <form onSubmit={handleSearch} className="flex-1 w-full">
          <div className="relative">
            <SearchIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-muted-foreground" />
            <Input
              type="text"
              placeholder="Search videos..."
              value={searchQuery}
              onChange={(e) => {
                isUserTypingRef.current = true
                setSearchQuery(e.target.value)
              }}
              className="pl-11 pr-4 h-11 w-full text-base"
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  isUserTypingRef.current = false
                  handleSearch(e)
                }
              }}
            />
          </div>
        </form>
      </div>
      <div className="flex items-center gap-2 flex-shrink-0">
        <Button
          variant="ghost"
          size="icon"
          onClick={toggleTheme}
          aria-label="Toggle theme"
          className="rounded-full"
        >
          {isDark ? <Sun className="w-5 h-5" /> : <Moon className="w-5 h-5" />}
        </Button>
      </div>
    </header>
  )
}
