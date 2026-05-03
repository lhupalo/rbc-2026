# Build alvo linux/amd64 (requisito da competição).
FROM --platform=linux/amd64 golang:1.26-bookworm AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /out/api /api
EXPOSE 8080
USER nobody
ENTRYPOINT ["/api"]
