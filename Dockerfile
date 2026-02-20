# Build frontend
FROM node:22-alpine AS frontend-builder
WORKDIR /app/web
RUN npm install -g bun
COPY web/package.json web/bun.lock ./
RUN bun install --frozen-lockfile
COPY web/ ./
RUN bun run build

# Build backend
FROM golang:1.24-alpine AS backend-builder
WORKDIR /app
RUN apk update && apk add --no-cache upx ca-certificates tzdata
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend-builder /app/web/build ./web/build
RUN CGO_ENABLED=0 go build -a -ldflags="-w -s" -o motophoto . \
	&& upx -q motophoto

# Runtime
FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata
COPY --from=backend-builder /app/motophoto .
COPY --from=backend-builder /app/web/build ./web/build
EXPOSE 8080
ENV PORT=8080
CMD ["./motophoto"]
