# Build stage
FROM golang:1.22-bullseye AS builder
WORKDIR /app

# Copy go.mod and go.sum, then download deps
COPY go.mod go.sum ./
RUN /usr/local/go/bin/go mod download

# Copy the source
COPY . .

# Build the binary
RUN CGO_ENABLED=1 GOOS=linux /usr/local/go/bin/go build -v -o /forum-app .

# Final lightweight image
FROM debian:bookworm-slim

# Install CA certs and sqlite3 for runtime if needed
RUN apt-get update && apt-get install -y ca-certificates sqlite3 && rm -rf /var/lib/apt/lists/*

# Create app user
RUN useradd --no-create-home --uid 1000 appuser

# Copy binary
COPY --from=builder /forum-app /usr/local/bin/forum-app

# Copy static assets and templates
COPY static /app/static
COPY templates /app/templates

WORKDIR /app
USER appuser

ENV PORT=8080
EXPOSE 8080

CMD ["/usr/local/bin/forum-app"]
