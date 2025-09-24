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
    -o fixpanic \
    main.go

# Runtime stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache ca-certificates bash curl

# Create non-root user
RUN addgroup -g 1000 fixpanic && \
    adduser -D -u 1000 -G fixpanic fixpanic

# Set working directory
WORKDIR /home/fixpanic

# Copy binary from builder
COPY --from=builder /app/fixpanic /usr/local/bin/fixpanic

# Make binary executable
RUN chmod +x /usr/local/bin/fixpanic

# Create necessary directories
RUN mkdir -p /home/fixpanic/.config/fixpanic && \
    mkdir -p /home/fixpanic/.local/lib/fixpanic && \
    mkdir -p /home/fixpanic/.local/log/fixpanic && \
    chown -R fixpanic:fixpanic /home/fixpanic

# Switch to non-root user
USER fixpanic

# Set environment variables
ENV PATH="/home/fixpanic/.local/bin:${PATH}"

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD fixpanic agent status || exit 1

# Default command
ENTRYPOINT ["fixpanic"]

# Default arguments (show help)
CMD ["--help"]