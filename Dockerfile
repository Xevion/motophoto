# Stage 1: Build Go backend
FROM golang:1.26-alpine AS backend-builder
WORKDIR /build
RUN apk update && apk add --no-cache upx
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -a -ldflags="-w -s" -o motophoto . \
	&& upx -q motophoto

# Stage 2: Build SvelteKit frontend
FROM oven/bun:1 AS frontend-builder
WORKDIR /build
COPY web/package.json web/bun.lock ./
RUN bun install --frozen-lockfile
COPY web/ ./
RUN bun --smol run build

# Stage 3: Runtime
FROM oven/bun:1-slim
WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates wget \
    && rm -rf /var/lib/apt/lists/*

# Copy Go binary
COPY --from=backend-builder /build/motophoto ./motophoto

# Copy SvelteKit build output and runtime node_modules
COPY --from=frontend-builder /build/build ./web/build
COPY --from=frontend-builder /build/node_modules ./web/node_modules

# Copy entrypoint
COPY web/entrypoint.ts ./web/entrypoint.ts

ENV PORT=8080 \
    BACKEND_HOST=127.0.0.1 \
    BACKEND_PORT=3001 \
    TZ=Etc/UTC

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=30s --retries=3 \
    CMD wget -q --spider http://localhost:${PORT}/health || exit 1

ENTRYPOINT ["bun", "run", "/app/web/entrypoint.ts"]
