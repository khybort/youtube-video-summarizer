import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { Layout } from '@/components/layout/Layout'
import { ToastProvider } from '@/components/ui/toast-provider'
import { Dashboard } from '@/pages/Dashboard'
import { VideoList } from '@/pages/VideoList'
import { VideoDetail } from '@/pages/VideoDetail'
import { Search } from '@/pages/Search'
import { CostAnalysis } from '@/pages/CostAnalysis'
import { Settings } from '@/pages/Settings'

function App() {
  // Initialize theme
  const theme = localStorage.getItem('theme') || 'system'
  if (theme === 'dark' || (theme === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
    document.documentElement.classList.add('dark')
  }

  return (
    <ToastProvider>
      <BrowserRouter>
        <Layout>
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/videos" element={<VideoList />} />
            <Route path="/videos/:id" element={<VideoDetail />} />
            <Route path="/search" element={<Search />} />
            <Route path="/costs" element={<CostAnalysis />} />
            <Route path="/settings" element={<Settings />} />
          </Routes>
        </Layout>
      </BrowserRouter>
    </ToastProvider>
  )
}

export default App

