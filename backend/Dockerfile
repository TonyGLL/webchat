# Stage 1: Build stage
FROM golang:1.22-alpine AS build

WORKDIR /app

# Install build dependencies for static binary
# Using musl-dev for static linking
RUN apk add --no-cache gcc musl-dev

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application for a static binary
RUN CGO_ENABLED=0 GOOS=linux go build -tags musl -ldflags '-extldflags "-static"' -o /main .

# Stage 2: Final stage
FROM scratch

WORKDIR /app

# Copy the binary from the build stage
COPY --from=build /main .

# Copy the environment configuration file. The build-arg is passed from docker-compose.
# It will be copied as ".env" so the app can load it consistently.
ARG CONFIG_FILE=dev.env
COPY ${CONFIG_FILE} .env

# Expose the port the app runs on
EXPOSE 8080

# Run the executable
CMD ["./main"]