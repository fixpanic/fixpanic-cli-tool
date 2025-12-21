# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
    -o opssquad \
    main.go

# Runtime stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache ca-certificates bash curl

# Create non-root user
RUN addgroup -g 1000 opssquad && \
    adduser -D -u 1000 -G opssquad opssquad

# Set working directory
WORKDIR /home/opssquad

# Copy binary from builder
COPY --from=builder /app/opssquad /usr/local/bin/opssquad

# Make binary executable
RUN chmod +x /usr/local/bin/opssquad

# Create necessary directories
RUN mkdir -p /home/opssquad/.config/opssquad && \
    mkdir -p /home/opssquad/.local/lib/opssquad && \
    mkdir -p /home/opssquad/.local/log/opssquad && \
    chown -R opssquad:opssquad /home/opssquad

# Switch to non-root user
USER opssquad

# Set environment variables
ENV PATH="/home/opssquad/.local/bin:${PATH}"

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD opssquad node status || exit 1

# Default command
ENTRYPOINT ["opssquad"]

# Default arguments (show help)
CMD ["--help"]