# Base images are pinned by tag and digest (Dependabot refreshes the digest weekly). The go
# directive in go.mod is the minimum Go patch: with GOTOOLCHAIN=local a stale builder fails the
# build instead of compiling against an older standard library.
# Stage 1: build a static binary
FROM golang:1.27-alpine@sha256:b8c9b82affbbb82a020bbcf998dad2ba9516944ac7e883524642a481a9505e9c AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# Stage 2: minimal runtime (CA certificates included for SMTP/TLS, runs as non-root).
# The binary lives outside /app because docker-compose mounts the project there
# so the process can read its CONFIG_FILE.
FROM gcr.io/distroless/static-debian13:nonroot@sha256:18f4cc26c9ffa7f20a359298be7deadac5c607c8e03a02b8119f1a5fc2f82c99
WORKDIR /app
COPY --from=build /out/server /server
EXPOSE 8081
USER nonroot:nonroot
ENTRYPOINT ["/server"]
