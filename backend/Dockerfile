# FROM node:20-alpine
# WORKDIR /app
# RUN apk add --no-cache bash git
# RUN npm i -g pnpm@10.18.3
# COPY package.json pnpm-lock.yaml ./
# RUN pnpm install --frozen-lockfile
# COPY . .
# EXPOSE 4000
# CMD ["pnpm", "dev"]


# ---- Build the Go CLI binary ----
FROM golang:1.25-alpine AS cli-builder
WORKDIR /app
COPY codeecho-cli/ .
RUN go mod download && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o codeecho ./main.go

# ---- Backend Node.js environment ----
FROM node:20-alpine
WORKDIR /app

RUN apk add --no-cache bash git
RUN npm i -g pnpm@10.18.3

COPY backend/package.json backend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

# Copy source code and .env file
COPY . .

# Copy CLI binary from Go builder
COPY --from=cli-builder /app/codeecho /usr/local/bin/codeecho

EXPOSE 4000
CMD ["pnpm", "dev"]



# Production file
# ===== Base Stage =====
# FROM node:20-alpine AS base
# WORKDIR /app
# COPY package.json pnpm-lock.yaml ./
# RUN apk add --no-cache python3 make g++ bash git
# RUN npm i -g pnpm@10.18.3
# RUN pnpm install --frozen-lockfile

# ===== Build Stage =====
# FROM base AS builder
# COPY . .
# RUN pnpm run build

# ===== Runtime Stage =====
# FROM node:20-alpine AS runtime
# WORKDIR /app
# ENV NODE_ENV=production
# COPY --from=builder /app/package.json ./
# COPY --from=builder /app/pnpm-lock.yaml ./
# COPY --from=builder /app/dist ./dist
# RUN npm i -g pnpm@10.18.3
# RUN pnpm install --prod --frozen-lockfile
# EXPOSE 4000
# CMD ["node", "dist/index.js"]

