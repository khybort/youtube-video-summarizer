# 🎬 YouTube Video Summarizer & Analyzer

A full-stack AI-powered platform for analyzing, summarizing, and discovering similar YouTube videos. Built with modern technologies and designed for scalability.

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-1.22+-00ADD8)
![Node Version](https://img.shields.io/badge/node-20+-339933)

## ✨ Features

### 🎥 Video Management
- **YouTube Integration**: Add videos by URL with automatic metadata extraction
- **Video Library**: Organize and browse your video collection
- **Search & Filter**: Find videos by title, channel, or status
- **Video Player**: Embedded YouTube player with transcript synchronization
- **Status Tracking**: Real-time processing status (pending, processing, completed, error)

### 📝 Transcription
- **Multiple Providers**: Choose from YouTube captions, Groq Whisper, Local Whisper, or Hugging Face
- **Automatic Detection**: Automatically fetches YouTube captions when available
- **Fallback Support**: Seamlessly falls back to Whisper if captions unavailable
- **Interactive Transcript**: Click on transcript segments to jump to video timestamps
- **Multi-language Support**: Support for various languages

### 🤖 AI-Powered Summarization
- **Multiple LLM Providers**: 
  - **Google Gemini** (cloud) - Fast and accurate
  - **Ollama/Llama** (local) - Privacy-focused, offline capable
- **Summary Types**:
  - **Short Summary**: Concise overview
  - **Detailed Summary**: Comprehensive analysis
  - **Bullet Points**: Key takeaways
- **Audio-Based Summarization**: Generate summaries directly from audio using Whisper
- **Dynamic Model Selection**: Automatically selects best available Gemini model
- **Regeneration**: Regenerate summaries with different models or from audio

### 🔍 Similarity Search
- **Vector-Based Similarity**: Uses pgvector for efficient similarity calculations
- **Multi-Modal Embeddings**: Combines title, description, and transcript embeddings
- **YouTube Integration**: Fetches similar videos directly from YouTube API
- **Similarity Scores**: Shows percentage match between videos
- **One-Click Addition**: Click similar videos to automatically add them to your library

### 💰 Cost Analysis
- **Token Usage Tracking**: Monitor input/output tokens for all operations
- **Cost Calculation**: Real-time cost tracking per provider and model
- **Usage Breakdown**: 
  - By provider (Gemini, Ollama, Groq, Local)
  - By operation (transcription, summarization, embedding)
  - By model
- **Time Periods**: View costs for today, week, month, or all time
- **Per-Video Costs**: Track costs for individual videos

### ⚙️ Configurable AI Models
- **Provider Selection**: Choose LLM and Whisper providers per operation
- **Model Configuration**: 
  - Select specific Gemini models or use auto-detection
  - Configure Ollama models and URLs
  - Set Whisper models and local service URLs
- **Health Checks**: Automatic detection of local service availability
- **Dynamic Settings**: Change providers without restarting

### 🚀 Event-Driven Architecture
- **Kafka Integration**: Asynchronous video processing pipeline
- **Background Workers**: 
  - Transcript worker
  - Embedding worker
  - Similarity worker
- **Fault Tolerance**: Dead letter queue for failed messages
- **Scalable**: Horizontal scaling with consumer groups

### 🎨 Modern UI/UX
- **Responsive Design**: Works seamlessly on desktop, tablet, and mobile
- **Dark Mode**: Beautiful dark theme with system preference detection
- **Real-time Updates**: Live status updates during video processing
- **Interactive Dashboard**: Statistics, quick actions, and recent videos
- **Toast Notifications**: User-friendly feedback for all actions
- **Loading States**: Skeleton loaders and spinners for better UX
- **Empty States**: Helpful messages when no data is available

## 🏗️ Architecture

### Tech Stack

#### Backend
- **Go 1.22+** with **Gin** framework
- **PostgreSQL 16** with **pgvector** extension for vector similarity
- **Redis** for caching and future job queues
- **Kafka** for event-driven processing
- **GORM** for database ORM
- **Zap** for structured logging

#### Frontend
- **React 18** with **TypeScript**
- **Vite** for fast development and building
- **Tailwind CSS** for styling
- **shadcn/ui** for component library
- **TanStack Query** for data fetching and caching
- **React Router DOM** for routing
- **Zustand** for state management
- **Playwright** for E2E testing

#### AI/ML Services
- **Google Gemini API** - Cloud LLM for summarization
- **Ollama** - Local LLM (Llama models)
- **Groq Whisper API** - Cloud speech-to-text
- **Local Whisper** - Python service using faster-whisper
- **Hugging Face** - Alternative Whisper provider

#### Infrastructure
- **Docker** & **Docker Compose** for containerization
- **Nginx** for frontend serving and API proxying
- **Multi-stage builds** for optimized images
- **Health checks** for all services
- **Prometheus** + **Grafana** + **cAdvisor** for monitoring
- **pprof** for performance profiling

### System Architecture

```
┌─────────────┐
│   Frontend  │ (React + TypeScript)
│  (Port 3000) │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   Nginx     │ (Reverse Proxy)
└──────┬──────┘
       │
       ▼
┌─────────────┐      ┌──────────────┐
│   Backend   │◄────►│  PostgreSQL  │
│  (Port 8080)│      │  (pgvector)  │
│  + pprof    │      └──────────────┘
└──────┬──────┘
       │
       ├──────────────┐
       │              │
       ▼              ▼
┌─────────────┐  ┌─────────────┐
│    Redis    │  │    Kafka    │
│  (Cache)    │  │  (Events)   │
└─────────────┘  └─────────────┘
       │
       ▼
┌─────────────────────────────┐
│      AI Services            │
│  ┌────────┐  ┌──────────┐   │
│  │ Gemini │  │  Ollama  │   │
│  └────────┘  └──────────┘   │
│  ┌────────┐  ┌──────────┐   │
│  │  Groq  │  │  Local   │   │
│  │ Whisper│  │ Whisper  │   │
│  └────────┘  └──────────┘   │
└─────────────────────────────┘
       │
       ▼
┌─────────────────────────────┐
│    Monitoring Stack         │
│  ┌──────────┐ ┌──────────┐  │
│  │ cAdvisor │ │Prometheus│  │
│  │  :8082   │ │  :9090   │  │
│  └──────────┘ └────┬─────┘  │
│                    │         │
│              ┌─────▼─────┐  │
│              │  Grafana  │  │
│              │   :3001   │  │
│              └───────────┘  │
└─────────────────────────────┘
```

### Data Flow

1. **Video Addition**: User adds YouTube URL → Backend fetches metadata → Video created
2. **Analysis Trigger**: User clicks "Analyze" → Kafka event published → Workers process
3. **Transcript Generation**: 
   - Try YouTube captions first
   - Fallback to configured Whisper provider
   - Store segments with timestamps
4. **Summary Generation**:
   - Use transcript or audio (with Whisper)
   - Generate with configured LLM provider
   - Extract key points
5. **Embedding Generation**: Create multi-modal embeddings → Store in pgvector
6. **Similarity Calculation**: Compare embeddings → Find similar videos → Store results

## 🚀 Quick Start

### Prerequisites

- **Docker** and **Docker Compose** (recommended)
- **Go 1.22+** (for local development)
- **Node.js 20+** (for local development)
- **API Keys**:
  - YouTube Data API v3 key
  - Google Gemini API key (optional, for cloud LLM)
  - Groq API key (optional, for cloud Whisper)

### Using Docker (Recommended)

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd youtube-video-summarizer
   ```

2. **Set up environment files**
   
   Create `.env.development`:
   ```bash
   cp .env.example .env.development
   ```
   
   Edit `.env.development` and add your API keys:
   ```env
   YOUTUBE_API_KEY=your_youtube_api_key
   GEMINI_API_KEY=your_gemini_api_key
   GROQ_API_KEY=your_groq_api_key
   ```

3. **Start all services**
   ```bash
   make dev
   # or
   make up APP_ENV=development
   ```

4. **Access the application**
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080
   - Health Check: http://localhost:8080/health
   - Kafka UI: http://localhost:8081 (optional)
   - Performance Dashboard (Grafana): http://localhost:3001 (admin/admin)
   - Prometheus: http://localhost:9090
   - cAdvisor: http://localhost:8082
   - pprof: http://localhost:8080/debug/pprof/

### Local Development

#### Backend
```bash
cd backend
go mod download
go run cmd/api/main.go
```

#### Frontend
```bash
cd frontend
npm install
npm run dev
```

#### Whisper Service (Optional)
```bash
cd whisper-service
pip install -r requirements.txt
uvicorn app:app --port 8001
```

## 📖 Usage Guide

### Adding a Video

1. Navigate to Dashboard
2. Paste a YouTube URL in the "Quick Add Video" section
3. Click "Add Video"
4. The video will be added and you'll be redirected to its detail page

### Analyzing a Video

1. Go to the video detail page
2. Click "Analyze Video" button
3. The system will:
   - Generate transcript (YouTube captions or Whisper)
   - Create embeddings
   - Find similar videos
   - Update status in real-time

### Generating a Summary

1. On the video detail page, go to the "Summary" tab
2. Choose summary type (Short, Detailed, or Bullet Points)
3. Optionally check "Generate from Audio" to use Whisper
4. Click "Generate Summary"
5. View the summary with key points

### Finding Similar Videos

1. On the video detail page, go to the "Similar" tab
2. View similar videos with similarity scores
3. Click on a similar video to add it to your library or view details

### Configuring AI Models

1. Go to Settings page
2. Configure providers:
   - **Transcript Provider**: YouTube, Groq, Local Whisper, or Hugging Face
   - **Summary Provider**: Gemini or Ollama
   - **Embedding Provider**: Gemini or Ollama
   - **Audio Analysis Provider**: Gemini or Ollama
3. Set model-specific options (Ollama URL, Whisper model, etc.)
4. Save settings

### Viewing Cost Analysis

1. Navigate to "Cost Analysis" page
2. Select time period (Today, Week, Month, All Time)
3. View:
   - Total cost and tokens
   - Breakdown by provider
   - Breakdown by operation
   - Breakdown by model
   - Per-video costs

## 🔌 API Endpoints

### Videos
- `POST /api/v1/videos` - Add video by URL
- `GET /api/v1/videos` - List videos (with pagination)
- `GET /api/v1/videos/:id` - Get video details
- `DELETE /api/v1/videos/:id` - Delete video
- `POST /api/v1/videos/:id/analyze` - Start full analysis

### Transcripts
- `GET /api/v1/videos/:id/transcript` - Get or create transcript
- `GET /api/v1/videos/:id/transcript/languages` - Get available caption languages

### Summaries
- `GET /api/v1/videos/:id/summary` - Get summary (creates if not exists)
- `POST /api/v1/videos/:id/summarize` - Generate summary
  - Body: `{ type: "short" | "detailed" | "bullet_points", from_audio?: boolean }`

### Similarity
- `GET /api/v1/videos/:id/similar` - Find similar videos
  - Query params: `threshold` (0-1, default: 0.7)

### Settings
- `GET /api/v1/settings` - Get user settings
- `PUT /api/v1/settings` - Update settings
- `GET /api/v1/settings/health/local-whisper` - Check local Whisper health

### Cost Analysis
- `GET /api/v1/costs/summary?period=today|week|month|all` - Get cost summary
- `GET /api/v1/costs/usage?period=today|week|month|all` - Get usage details
- `GET /api/v1/costs/video/:id` - Get video-specific costs

### Health
- `GET /health` - Health check endpoint
- `GET /ready` - Readiness check endpoint

### Performance Profiling (pprof)
- `GET /debug/pprof/` - pprof index page
- `GET /debug/pprof/heap` - Memory heap profile
- `GET /debug/pprof/profile?seconds=30` - CPU profile (30 seconds)
- `GET /debug/pprof/goroutine` - Goroutine profile
- `GET /debug/pprof/block` - Block profile

## 🧪 Testing

### Run All Tests
```bash
make test
```

### Backend Tests
```bash
make test-backend
# or
cd backend && go test ./... -v
```

### Frontend E2E Tests
```bash
make test-frontend
# or
cd frontend && npm run test:e2e
```

### Pre-commit Hooks
Tests run automatically before commits. All tests must pass to commit.

## 🛠️ Development

### Project Structure

```
youtube-video-summarizer/
├── backend/                 # Go API server
│   ├── cmd/api/            # Application entry point
│   ├── internal/           # Internal packages
│   │   ├── handlers/       # HTTP handlers
│   │   ├── services/       # Business logic
│   │   │   ├── video/      # Video service
│   │   │   ├── transcript/ # Transcript service
│   │   │   ├── summary/    # Summary service
│   │   │   ├── embedding/  # Embedding service
│   │   │   ├── similarity/ # Similarity service
│   │   │   ├── cost/        # Cost tracking service
│   │   │   ├── settings/   # Settings service
│   │   │   └── provider/   # Provider factory
│   │   ├── repository/     # Data access layer
│   │   ├── models/         # Domain models
│   │   ├── middleware/     # HTTP middleware
│   │   ├── jobs/           # Background jobs
│   │   └── workers/        # Kafka workers
│   └── pkg/                # Public packages
│       ├── llm/            # LLM providers (Gemini, Ollama)
│       ├── whisper/        # Whisper providers (Groq, Local, HF)
│       ├── youtube/        # YouTube API client
│       ├── embeddings/     # Embedding generation
│       ├── kafka/          # Kafka integration
│       ├── pricing/        # Cost calculation
│       └── errors/         # Error handling
├── frontend/               # React application
│   ├── src/
│   │   ├── components/     # React components
│   │   │   ├── layout/     # Layout components
│   │   │   ├── video/      # Video components
│   │   │   ├── transcript/ # Transcript viewer
│   │   │   ├── summary/    # Summary display
│   │   │   ├── similarity/ # Similar videos
│   │   │   ├── cost/       # Cost analysis
│   │   │   └── ui/         # UI components
│   │   ├── pages/          # Page components
│   │   ├── hooks/          # Custom React hooks
│   │   ├── services/       # API clients
│   │   └── types/          # TypeScript types
│   └── e2e/                # Playwright E2E tests
├── whisper-service/        # Python Whisper service
│   ├── app.py              # FastAPI application
│   ├── Dockerfile          # Docker configuration
│   └── requirements.txt    # Python dependencies
├── docker-compose.yml      # Docker Compose configuration
├── Makefile                # Development commands
├── scripts/                # Utility scripts
└── monitoring/             # Monitoring stack configuration
    ├── prometheus/         # Prometheus config
    ├── grafana/            # Grafana dashboards & provisioning
    ├── DASHBOARD_ACCESS.md # Dashboard access guide
    └── CADVISOR_QUERIES.md # Service-based query examples
```

### Services Architecture

Backend services separate business logic from the repository layer, providing a modular and testable structure. Each service has a specific responsibility and is used across different layers:

#### Video Service (`services/video`)
- **Where Used**: 
  - **Handlers** (`video_handler.go`): Video CRUD operations, adding, listing, and deleting videos
  - **Workers** (`transcript_worker.go`, `embedding_worker.go`): Updating video status (pending → processing → completed)
  - **Jobs** (`analysis_job.go`): Retrieving video information during analysis operations
- **Why**: Centralizes YouTube API integration for fetching video metadata, video status management, and database operations

#### Transcript Service (`services/transcript`)
- **Where Used**:
  - **Handlers** (`video_handler.go`): Creating/getting transcripts on user requests
  - **Workers** (`transcript_worker.go`): Asynchronous transcript processing from Kafka events
  - **Jobs** (`analysis_job.go`): Creating transcripts in the video analysis pipeline
  - **Embedding Service**: Uses transcript content to generate embeddings
- **Why**: Manages different providers (YouTube captions, Groq Whisper, Local Whisper, Hugging Face) and provides fallback mechanisms

#### Summary Service (`services/summary`)
- **Where Used**:
  - **Handlers** (`video_handler.go`): Creating summaries on user requests (short, detailed, bullet points)
- **Why**: Generates summaries from transcripts or audio using LLM providers (Gemini, Ollama) and tracks costs

#### Embedding Service (`services/embedding`)
- **Where Used**:
  - **Handlers** (`video_handler.go`): Creating embeddings during video analysis
  - **Workers** (`embedding_worker.go`): Asynchronous embedding processing from Kafka events
  - **Jobs** (`analysis_job.go`): Creating embeddings in the video analysis pipeline
- **Why**: Generates multi-modal embeddings from video title, description, and transcript, stores them in pgvector, required for similarity search

#### Similarity Service (`services/similarity`)
- **Where Used**:
  - **Handlers** (`video_handler.go`): Processing requests to find similar videos
  - **Workers** (`similarity_worker.go`): Asynchronous similarity calculation from Kafka events
- **Why**: Calculates cosine similarity between embeddings, finds similar videos via YouTube API, caches results

#### Cost Service (`services/cost`)
- **Where Used**:
  - **Handlers** (`cost_handler.go`): Cost analysis endpoints (summary, usage, per-video)
  - **Transcript Service**: Records transcript creation costs
  - **Summary Service**: Records summary generation costs
  - **Embedding Service**: Records embedding generation costs
- **Why**: Tracks token usage and costs for all AI operations, provides provider/model-based analysis

#### Settings Service (`services/settings`)
- **Where Used**:
  - **Handlers** (`settings_handler.go`): Getting/updating user settings
  - **Provider Factory**: Reads settings to determine which AI provider to use
- **Why**: Manages user preferences (LLM provider, Whisper provider, model selections) and enables dynamic provider switching

#### Provider Factory (`services/provider`)
- **Where Used**:
  - **Summary Service**: To get LLM provider (Gemini/Ollama)
  - **Transcript Service**: To get Whisper provider (Groq/Local/HuggingFace)
  - **Embedding Service**: To get LLM provider for embeddings
- **Why**: Creates the correct provider instance based on settings, centrally manages provider changes, performs health checks

### Environment Variables

#### Backend
```env
# Application
APP_ENV=development
PORT=8080
GIN_MODE=debug

# Database
DATABASE_URL=postgres://user:pass@localhost:5432/youtube_analyzer?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379

# Kafka
KAFKA_BROKERS=localhost:9092
KAFKA_ENABLED=true
KAFKA_CONSUMER_GROUP=youtube-analyzer

# YouTube
YOUTUBE_API_KEY=your_key

# Gemini
GEMINI_API_KEY=your_key

# Ollama
OLLAMA_URL=http://localhost:11434
OLLAMA_MODEL=llama3.2

# Groq
GROQ_API_KEY=your_key

# Local Whisper
LOCAL_WHISPER_URL=http://localhost:8001
WHISPER_MODEL=base
```

#### Frontend
```env
VITE_API_URL=http://localhost:8080
```

### Makefile Commands

```bash
make help          # Show all available commands
make build         # Build all Docker images
make up            # Start all services
make down          # Stop all services
make dev           # Start development environment
make prod          # Start production environment
make logs          # Show logs from all services
make logs-backend  # Show backend logs
make logs-frontend # Show frontend logs
make test          # Run all tests
make test-backend  # Run backend tests
make test-frontend # Run frontend E2E tests
make clean         # Remove all containers and volumes
make restart       # Restart all services
make status        # Show status of all services
```

## 🔒 Security

- Environment variables for sensitive data
- API keys stored securely (never committed)
- CORS configuration for frontend
- Input validation on all endpoints
- SQL injection protection via GORM
- Error messages don't expose sensitive information

## 📊 Database Schema

### Core Tables
- `videos` - Video metadata
- `transcripts` - Video transcripts with segments (JSONB)
- `summaries` - AI-generated summaries
- `video_embeddings` - Vector embeddings (pgvector)
- `video_similarities` - Pre-computed similarity scores
- `token_usage` - Cost tracking
- `settings` - User configuration

### Indexes
- HNSW indexes on embeddings for fast similarity search
- Indexes on video status, created_at for efficient queries
- Composite indexes for common query patterns

## ⚡ Performance Optimizations

### CPU & Memory Optimizations

The system includes several optimizations to reduce CPU and memory usage:

#### Kafka Consumer Optimizations
- **ReadMessage** instead of FetchMessage for better backoff handling
- **Increased MaxWait** (10s) to reduce polling frequency
- **Exponential backoff** when no messages available
- **ReadBackoffMin/Max** set to 2s/10s for efficient resource usage

#### Kafka Broker Optimizations
- Reduced thread counts (network: 3, IO: 4)
- Increased log retention check interval (10 minutes)
- Optimized compression (snappy)
- Limited background threads

#### Whisper Service Optimizations
- CPU thread limiting (2 threads)
- Single worker for transcriptions
- Reduced beam size for faster processing
- Docker CPU limits (max 2 CPUs)
- Thread pool executor to prevent concurrent overload

#### Producer Optimizations
- Batch processing (batch size: 10)
- Increased batch timeout (100ms)
- RequiredAcks set to One for lower CPU usage

### Monitoring & Profiling

All performance metrics are available through:
- **Grafana Dashboard**: Real-time CPU, Memory, Network metrics
- **pprof**: Detailed Go profiling (CPU, Memory, Goroutines)
- **Prometheus**: Queryable metrics with PromQL
- **cAdvisor**: Container-level metrics

See [Performance Monitoring Dashboard](#-performance-monitoring-dashboard) section for details.

## 🚢 Deployment

### Production Build

```bash
make build APP_ENV=production
make prod
```

### Environment Setup

1. Create `.env.production` with production values
2. Use secret management (AWS Secrets Manager, HashiCorp Vault, etc.)
3. Set `APP_ENV=production`
4. Configure production database and Redis
5. Set up monitoring and logging
6. Enable monitoring stack: `docker compose up -d cadvisor prometheus grafana`

### Docker Compose Production

```bash
docker compose --env-file .env.production up -d
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make test`
5. Commit with clear messages
6. Push and create a pull request

## 📝 License

MIT License - see LICENSE file for details

## 🙏 Acknowledgments

- **Google Gemini** for LLM capabilities
- **Ollama** for local LLM support
- **Groq** for fast Whisper transcription
- **faster-whisper** for local speech-to-text
- **pgvector** for vector similarity search
- **shadcn/ui** for beautiful components

## 📊 Performance Monitoring Dashboard

### Grafana Dashboard (Real-time Metrics)

**Access**: http://localhost:3001
- **Username**: `admin`
- **Password**: `admin` (change on first login)

**Features**:
- Real-time CPU and Memory usage for all containers
- Network I/O metrics
- Auto-refresh every 5 seconds
- Customizable dashboards
- Service-based filtering and comparison

**Start Monitoring Stack**:
```bash
docker compose up -d cadvisor prometheus grafana
```

**Dashboard Includes**:
- CPU Usage by Container (real-time graph)
- Memory Usage by Container
- Backend CPU/Memory (stat panels)
- Kafka CPU monitoring
- Network I/O (RX/TX) for all services
- Service comparison views

### pprof Endpoints (Backend Performance Profiling)

**Access**: http://localhost:8080/debug/pprof/

**Available Profiles**:
- `/debug/pprof/` - Index page (list of all profiles)
- `/debug/pprof/heap` - Memory heap profile
- `/debug/pprof/profile?seconds=30` - CPU profile (30 seconds)
- `/debug/pprof/goroutine` - Goroutine profile
- `/debug/pprof/block` - Block profile

**Usage Examples**:
```bash
# Get CPU profile (30 seconds)
go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30

# Get memory profile
go tool pprof http://localhost:8080/debug/pprof/heap

# View in web UI
go tool pprof -http=:8080 http://localhost:8080/debug/pprof/heap
```

### Other Monitoring Services

- **Prometheus**: http://localhost:9090 (Metric queries with PromQL)
- **cAdvisor**: http://localhost:8082 (Container metrics UI)

**Service-Based Queries** (Prometheus):
```promql
# Backend CPU
rate(container_cpu_usage_seconds_total{name="youtube-analyzer-backend"}[5m]) * 100

# All services CPU comparison
rate(container_cpu_usage_seconds_total{name=~"youtube-analyzer-.*"}[5m]) * 100
```

**Detailed Documentation**:
- [monitoring/DASHBOARD_ACCESS.md](./monitoring/DASHBOARD_ACCESS.md) - Complete access guide
- [monitoring/CADVISOR_QUERIES.md](./monitoring/CADVISOR_QUERIES.md) - Service-based queries

## 📞 Support

For issues, questions, or contributions, please open an issue on GitHub.

---

**Built with ❤️ using Go, React, and AI**
