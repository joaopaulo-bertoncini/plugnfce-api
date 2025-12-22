# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.24-alpine AS builder

# Install CGO dependencies and libxml2 for XSD validation
RUN apk add --no-cache \
    gcc \
    musl-dev \
    libxml2-dev \
    pkgconfig \
    git

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build API binary
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o plugnfce-api ./cmd/api

# Build Worker binary
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o plugnfce-worker ./cmd/worker

# Runtime stage
FROM alpine:3.19

# Install ca-certificates for HTTPS requests and runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    libxml2 \
    tzdata

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/plugnfce-api .
COPY --from=builder /app/plugnfce-worker .

# Copy schemas directory
COPY --from=builder /app/internal/infrastructure/sefaz/schemas ./schemas

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Change ownership of app directory
RUN chown -R appuser:appgroup /app
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["sh", "-c", "if [ \"$APP_COMPONENT\" = \"api\" ]; then wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1; else echo 'Worker health check not implemented'; fi"]

EXPOSE 8080

# Default to API, can be overridden
ENV APP_COMPONENT=api
CMD ["sh", "-c", "if [ \"$APP_COMPONENT\" = \"worker\" ]; then ./plugnfce-worker; else ./plugnfce-api; fi"] 