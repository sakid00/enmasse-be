# syntax=docker/dockerfile:1

FROM golang:1.27-bookworm AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/enmasse-api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=builder /out/enmasse-api /app/enmasse-api
COPY --from=builder /src/migrations /app/migrations
EXPOSE 8081
ENTRYPOINT ["/app/enmasse-api"]
