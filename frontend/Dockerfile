# ==================== BUILD STAGE ====================
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files
COPY package.json package-lock.json* ./

    # Install dependencies
    RUN npm install --silent

# Copy source
COPY . .

# Build args for environment
ARG VITE_API_URL
ENV VITE_API_URL=$VITE_API_URL

# Build production bundle
RUN npm run build

# ==================== RUNTIME STAGE ====================
FROM nginx:alpine AS runtime

# Remove default nginx config
RUN rm /etc/nginx/conf.d/default.conf

# Copy custom nginx config
COPY nginx.conf /etc/nginx/conf.d/

# Copy built assets from builder
COPY --from=builder /app/dist /usr/share/nginx/html

# Add healthcheck
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:80/health || exit 1

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]

