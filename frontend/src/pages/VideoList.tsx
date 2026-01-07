import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { VideoCard } from '@/components/video/VideoCard'
import { VideoCardSkeleton } from '@/components/video/VideoCardSkeleton'
import { videoService } from '@/services/videoService'
import { Search, Filter, X } from 'lucide-react'

export function VideoList() {
  const navigate = useNavigate()
  const [searchQuery, setSearchQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState<string>('all')

  const { data, isLoading } = useQuery({
    queryKey: ['videos'],
    queryFn: () => videoService.getAll(),
  })

  // Filter videos
  const filteredVideos = data?.videos?.filter(video => {
    const matchesSearch = !searchQuery || 
      video.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
      video.channelName.toLowerCase().includes(searchQuery.toLowerCase())
    const matchesStatus = statusFilter === 'all' || video.status === statusFilter
    return matchesSearch && matchesStatus
  }) || []

  const statusOptions = [
    { value: 'all', label: 'All Status' },
    { value: 'completed', label: 'Completed' },
    { value: 'processing', label: 'Processing' },
    { value: 'pending', label: 'Pending' },
    { value: 'error', label: 'Error' },
  ]

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-4xl font-bold bg-gradient-to-r from-primary to-primary/60 bg-clip-text text-transparent">
          All Videos
        </h1>
        <p className="text-muted-foreground mt-2 text-lg">
          {data?.videos?.length || 0} video{(data?.videos?.length || 0) !== 1 ? 's' : ''} in your library
        </p>
      </div>

      {/* Filters */}
      <Card>
        <CardContent className="p-4">
          <div className="flex flex-col sm:flex-row gap-4">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-muted-foreground" />
              <Input
                type="text"
                placeholder="Search by title or channel..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-10"
              />
              {searchQuery && (
                <Button
                  variant="ghost"
                  size="icon"
                  className="absolute right-1 top-1/2 transform -translate-y-1/2 h-7 w-7"
                  onClick={() => setSearchQuery('')}
                >
                  <X className="w-4 h-4" />
                </Button>
              )}
            </div>
            <div className="flex gap-2">
              {statusOptions.map((option) => (
                <Button
                  key={option.value}
                  variant={statusFilter === option.value ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setStatusFilter(option.value)}
                  className="capitalize"
                >
                  {option.label}
                </Button>
              ))}
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Results Count */}
      {!isLoading && (
        <div className="text-sm text-muted-foreground">
          Showing {filteredVideos.length} of {data?.videos?.length || 0} videos
          {(searchQuery || statusFilter !== 'all') && (
            <Button
              variant="ghost"
              size="sm"
              className="ml-2 h-auto p-0 text-xs"
              onClick={() => {
                setSearchQuery('')
                setStatusFilter('all')
              }}
            >
              Clear filters
            </Button>
          )}
        </div>
      )}

      {/* Video Grid */}
      {isLoading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
          {[...Array(8)].map((_, i) => (
            <VideoCardSkeleton key={i} />
          ))}
        </div>
      ) : filteredVideos.length === 0 ? (
        <Card className="border-2 border-dashed">
          <CardContent className="py-16 text-center">
            {searchQuery || statusFilter !== 'all' ? (
              <>
                <Filter className="w-16 h-16 mx-auto mb-4 text-muted-foreground/50" />
                <h3 className="text-lg font-semibold mb-2">No videos found</h3>
                <p className="text-muted-foreground mb-6 max-w-md mx-auto">
                  Try adjusting your search or filter criteria.
                </p>
                <Button
                  variant="outline"
                  onClick={() => {
                    setSearchQuery('')
                    setStatusFilter('all')
                  }}
                >
                  Clear Filters
                </Button>
              </>
            ) : (
              <>
                <Search className="w-16 h-16 mx-auto mb-4 text-muted-foreground/50" />
                <h3 className="text-lg font-semibold mb-2">No videos yet</h3>
                <p className="text-muted-foreground mb-6 max-w-md mx-auto">
                  Get started by adding videos from the dashboard.
                </p>
                <Button onClick={() => navigate('/')}>
                  Go to Dashboard
                </Button>
              </>
            )}
          </CardContent>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
          {filteredVideos.map((video) => (
            <VideoCard key={video.id} video={video} />
          ))}
        </div>
      )}
    </div>
  )
}
