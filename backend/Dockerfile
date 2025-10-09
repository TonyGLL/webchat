# Stage 1: build
FROM golang:1.24-alpine AS build
WORKDIR /src
RUN apk add --no-cache gcc musl-dev
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -tags musl -ldflags '-extldflags "-static"' -o /main ./cmd/server

# Final: imagen mínima
FROM scratch
WORKDIR /app
COPY --from=build /main /main
COPY --from=build /main /app/main

ARG CONFIG_FILE=dev.env
COPY ${CONFIG_FILE} .env

EXPOSE 3000
CMD ["/main"]