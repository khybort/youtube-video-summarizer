import axios from 'axios'

// Use relative path when running through nginx, absolute URL for development
const getBaseURL = () => {
  // If VITE_API_URL is set and it's a full URL, use it
  if (import.meta.env.VITE_API_URL && import.meta.env.VITE_API_URL.startsWith('http')) {
    return import.meta.env.VITE_API_URL
  }
  
  // Always use relative path - nginx will proxy to backend
  // This works both in Docker (nginx) and development (if nginx is running)
  // IMPORTANT: Must start with / to be relative to the current origin
  const baseURL = '/api/v1'
  console.log('[API] Base URL:', baseURL) // Debug log
  return baseURL
}

const api = axios.create({
  baseURL: getBaseURL(),
  timeout: 30000, // Default timeout for most requests (30 seconds)
  headers: {
    'Content-Type': 'application/json',
  },
})

// Extended timeout for long-running operations (10 minutes)
export const apiWithExtendedTimeout = axios.create({
  baseURL: getBaseURL(),
  timeout: 10 * 60 * 1000, // 10 minutes for summarize, analyze, etc.
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor
api.interceptors.request.use(
  (config) => {
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor for default API
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response) {
      // Server responded with error
      const message = error.response.data?.error || error.response.data?.message || 'An error occurred'
      return Promise.reject(new Error(message))
    } else if (error.request) {
      // Request made but no response
      return Promise.reject(new Error('Network error. Please check your connection.'))
    } else {
      // Something else happened
      return Promise.reject(error)
    }
  }
)

// Response interceptor for extended timeout API (same error handling)
apiWithExtendedTimeout.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response) {
      // Server responded with error
      const message = error.response.data?.error || error.response.data?.message || 'An error occurred'
      return Promise.reject(new Error(message))
    } else if (error.request) {
      // Request made but no response
      return Promise.reject(new Error('Network error. Please check your connection.'))
    } else {
      // Something else happened
      return Promise.reject(error)
    }
  }
)

export default api

