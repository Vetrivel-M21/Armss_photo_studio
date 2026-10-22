# ----------------------------------------------------
# ARMSS Photo Studio — React/Vite Frontend Dockerfile
# Two-stage build: Node build -> nginx static serve
# ----------------------------------------------------
FROM node:20-alpine AS builder

WORKDIR /app

# Install deps (separate layer for caching)
COPY package.json package-lock.json ./
RUN npm ci --legacy-peer-deps

# Copy source and build
COPY . .

# VITE_API_BASE_URL is injected at build time so the compiled JS bundle
# knows where to call the backend API
ARG VITE_API_BASE_URL=/api/v1
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL

RUN npm run build

# ----------------------------------------------------
# Serve compiled static files via nginx
# ----------------------------------------------------
FROM nginx:alpine

# Remove default nginx config
RUN rm /etc/nginx/conf.d/default.conf

# Our custom SPA nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy built assets from the builder stage
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
