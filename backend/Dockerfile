# Stage 1: build
FROM golang:1.24-alpine AS build
WORKDIR /app
RUN apk add --no-cache gcc musl-dev
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -tags musl -ldflags '-extldflags "-static"' -o /main .

# Final
FROM scratch
WORKDIR /app
# Copiamos el binario en / (fuera de /app) y en /app para compatibilidad
COPY --from=build /main /main
COPY --from=build /main /app/main

ARG CONFIG_FILE=dev.env
COPY ${CONFIG_FILE} .env

EXPOSE 3000
CMD ["/main"]