FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM alpine:3.23

RUN apk add --no-cache ca-certificates su-exec \
    && addgroup -S app \
    && adduser -S -G app app \
    && mkdir -p /data/uploads \
    && chown -R app:app /data/uploads

WORKDIR /app
COPY --from=builder /out/api /app/api
COPY docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod 0755 /app/docker-entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["/app/docker-entrypoint.sh"]

