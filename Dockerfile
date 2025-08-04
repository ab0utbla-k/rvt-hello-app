# Build stage
FROM golang:1.24-alpine3.22 AS builder

# Build args
ARG BUILD_REF=""
ENV CGO_ENABLED=0

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Copy migrations to cmd/api for embed
COPY migrations/ ./cmd/api/migrations/

# Build the binary with version injection
RUN go build -ldflags "-X main.build=${BUILD_REF}" -o main ./cmd/api


FROM alpine:3.22

# Build args for metadata
ARG BUILD_REF
ARG BUILD_DATE

# Add a non-root user
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

WORKDIR /app

# Copy binary and migrations directory
COPY --from=builder --chown=appuser:appgroup /app/main .

USER appuser

EXPOSE 4000
CMD ["./main"]

LABEL org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.title="birthday-api" \
      org.opencontainers.image.authors="Ihar Statkevich <ihast.online@pm.me>" \
      org.opencontainers.image.source="https://github.com/ab0utbla-k/rvt-hello-app" \
      org.opencontainers.image.revision="${BUILD_REF}"