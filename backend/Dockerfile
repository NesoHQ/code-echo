# ================================
# Stage 1: Build Go CLI binary
# ================================
FROM golang:1.25-alpine AS cli-builder
WORKDIR /src
RUN apk add --no-cache git
COPY ../codeecho-cli ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o codeecho ./main.go

# ================================
# Stage 2: Install dependencies
# ================================
FROM node:20-alpine AS deps
WORKDIR /app
RUN apk add --no-cache python3 make g++ bash git
RUN npm i -g pnpm@10.18.3
COPY backend/package.json backend/pnpm-lock.yaml ./
RUN pnpm fetch

# ================================
# Stage 3: Build backend (TypeScript → JS)
# ================================
FROM deps AS builder
WORKDIR /app
COPY backend ./
ENV CI=true
RUN pnpm install --offline --frozen-lockfile
RUN pnpm run build

# Verify build output
RUN ls -la /app/dist || (echo "❌ dist folder not found. Check tsconfig and build script." && exit 1)

# ================================
# Stage 4: Runtime image
# ================================
FROM node:20-alpine AS runtime
WORKDIR /app
RUN apk add --no-cache bash git
RUN npm i -g pnpm@10.18.3

# Copy compiled app + dependencies
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/package.json ./
COPY --from=builder /app/pnpm-lock.yaml ./
ENV NODE_ENV=production
RUN pnpm install --frozen-lockfile

# Copy CLI binary
COPY --from=cli-builder /src/codeecho /usr/local/bin/codeecho
RUN chmod +x /usr/local/bin/codeecho

# Add a non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Create writable directory for outputs
ENV TEMP_DIR=/app/data/outputs
RUN mkdir -p ${TEMP_DIR} && chown -R appuser:appgroup ${TEMP_DIR}

USER appuser
EXPOSE 4000

# Healthcheck for container monitoring
HEALTHCHECK --interval=30s --timeout=10s --start-period=20s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:4000/health || exit 1

# Start backend
CMD ["pnpm", "start"]
