# Stage 1: Builder Stage (for compiling Go application)
# Using the official Golang Alpine image to keep the size small
FROM golang:1.27-alpine AS builder

# Set working directory in the container to /app
WORKDIR /app

# Copy only the dependency files to cache the 'go mod download' step
COPY go.mod go.sum ./
# Download dependencies and build the Go application in a single step to leverage Docker's cache
RUN go mod download
# Copy the rest of the application code
COPY . .

# Build the application with version information included
RUN CGO_ENABLED=0 \
  GOOS=linux \
  go build \
  -ldflags "-s -w -buildid= -extldflags '-static'" \
  -trimpath \
  -tags netgo \
  -o fugocreative-backend ./cmd/server/

# Stage 2: Final Image (minimal size for deployment)
FROM alpine:3.21

# Copy the pre-built binary from the builder stage
COPY --from=builder /app/fugocreative-backend /

# Copy the migrations directory from the builder stage to the final image
COPY --from=builder /app/migrations /migrations

# Set application version and environment as labels
LABEL app.name="fugocreative-backend"

# Set environment variable for the HTTP port
ENV HTTP_PORT=8080
# Expose the port to the host system
EXPOSE 8080
# Specify the entry point for the container (run the Go application)
ENTRYPOINT ["./fugocreative-backend"]

