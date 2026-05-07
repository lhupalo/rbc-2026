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
COPY data/normalization.json   /data/normalization.json
COPY data/mcc_risk.json        /data/mcc_risk.json
COPY data/ivf_config.json      /data/ivf_config.json
COPY data/centroids.bin        /data/centroids.bin
COPY data/vectors.bin          /data/vectors.bin
COPY data/ivf_structure.bin    /data/ivf_structure.bin
ENV DATA_DIR=/data
EXPOSE 8080
USER nobody
ENTRYPOINT ["/api"]
