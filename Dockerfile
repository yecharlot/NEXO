FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/nexo ./cmd/nexo

FROM alpine:3.21
RUN apk add --no-cache ca-certificates ffmpeg && addgroup -S nexo && adduser -S -G nexo nexo
WORKDIR /app
COPY --from=build /out/nexo /usr/local/bin/nexo
COPY web ./web
RUN mkdir -p /var/lib/nexo && chown -R nexo:nexo /var/lib/nexo /app
USER nexo
ENV NEXO_ADDR=:8090 NEXO_DATA_DIR=/var/lib/nexo NEXO_WEB_DIR=/app/web
EXPOSE 8090
VOLUME ["/var/lib/nexo"]
ENTRYPOINT ["/usr/local/bin/nexo"]
