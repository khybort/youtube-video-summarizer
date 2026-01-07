import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { VideoCard } from '@/components/video/VideoCard'
import { VideoCardSkeleton } from '@/components/video/VideoCardSkeleton'
import { videoService } from '@/services/videoService'
import { Search as SearchIcon, Filter, X } from 'lucide-react'
import { EmptyState } from '@/components/ui/empty-state'

export function Search() {
  const [searchTerm, setSearchTerm] = useState('')
  const [query, setQuery] = useState('')
  const [filters, setFilters] = useState({
    status: '',
    channel: '',
  })

  const { data, isLoading } = useQuery({
    queryKey: ['videos', 'search', query, filters],
    queryFn: () => videoService.getAll({ limit: 50 }), // Backend doesn't support search yet, fetch all and filter client-side
    enabled: true, // Always enabled, filtering will be done client-side
  })

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    setQuery(searchTerm)
  }

  const clearFilters = () => {
    setFilters({ status: '', channel: '' })
    setQuery('')
    setSearchTerm('')
  }

  const hasActiveFilters = Object.values(filters).some((v) => v !== '') || query !== ''

  // Client-side filtering (backend doesn't support search yet)
  const allVideos = data?.videos || []
  const videos = allVideos.filter((video) => {
    const matchesSearch = !query || 
      video.title.toLowerCase().includes(query.toLowerCase()) ||
      video.description?.toLowerCase().includes(query.toLowerCase()) ||
      video.channelName?.toLowerCase().includes(query.toLowerCase())
    const matchesStatus = !filters.status || video.status === filters.status
    const matchesChannel = !filters.channel || video.channelName?.toLowerCase().includes(filters.channel.toLowerCase())
    return matchesSearch && matchesStatus && matchesChannel
  })

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Search Videos</h1>
        <p className="text-muted-foreground mt-2">
          Find videos by title, description, or channel
        </p>
      </div>

      {/* Search Bar */}
      <Card>
        <CardContent className="p-4">
          <form onSubmit={handleSearch} className="flex gap-2">
            <div className="relative flex-1">
              <SearchIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-muted-foreground" />
              <Input
                type="text"
                placeholder="Search videos..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-10"
              />
            </div>
            <Button type="submit">Search</Button>
            {hasActiveFilters && (
              <Button type="button" variant="outline" onClick={clearFilters}>
                <X className="w-4 h-4 mr-2" />
                Clear
              </Button>
            )}
          </form>
        </CardContent>
      </Card>

      {/* Filters */}
      <Card>
        <CardContent className="p-4">
          <div className="flex items-center gap-4">
            <Filter className="w-4 h-4 text-muted-foreground" />
            <span className="text-sm font-medium">Filters:</span>
            <select
              value={filters.status}
              onChange={(e) => setFilters({ ...filters, status: e.target.value })}
              className="flex h-10 rounded-md border border-input bg-background px-3 py-2 text-sm"
            >
              <option value="">All Status</option>
              <option value="pending">Pending</option>
              <option value="processing">Processing</option>
              <option value="completed">Completed</option>
              <option value="error">Error</option>
            </select>
          </div>
        </CardContent>
      </Card>

      {/* Results */}
      {isLoading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
          {[...Array(8)].map((_, i) => (
            <VideoCardSkeleton key={i} />
          ))}
        </div>
      ) : videos.length > 0 ? (
        <>
          <div className="text-sm text-muted-foreground">
            Found {videos.length} video{videos.length !== 1 ? 's' : ''}
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {videos.map((video) => (
              <VideoCard key={video.id} video={video} />
            ))}
          </div>
        </>
      ) : hasActiveFilters ? (
        <EmptyState
          title="No videos found"
          description="Try adjusting your search or filters"
        />
      ) : (
        <EmptyState
          title="Start searching"
          description="Enter a search term or apply filters to find videos"
        />
      )}
    </div>
  )
}

