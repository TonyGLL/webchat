# Stage 1: build a static binary
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# Stage 2: minimal runtime (CA certificates included for SMTP/TLS, runs as non-root).
# The binary lives outside /app because docker-compose mounts the project there
# so the process can read its CONFIG_FILE.
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/server /server
EXPOSE 8081
USER nonroot:nonroot
ENTRYPOINT ["/server"]
