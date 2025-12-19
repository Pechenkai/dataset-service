FROM golang:1.21-alpine AS builder
ENV GOTOOLCHAIN=auto
WORKDIR /src
COPY go.mod go.sum ./
# Явно указываем прокси для стабильной загрузки зависимостей
RUN go env -w GOPROXY=https://proxy.golang.org,direct
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/ppo ./cmd/api && \
    CGO_ENABLED=0 GOOS=linux go build -o /bin/gateway ./cmd/gateway && \
    CGO_ENABLED=0 GOOS=linux go build -o /bin/core ./cmd/core && \
    CGO_ENABLED=0 GOOS=linux go build -o /bin/data ./cmd/data

FROM alpine:3.19
RUN adduser -D -g '' appuser
WORKDIR /app
COPY --from=builder /bin/ppo /app/ppo
COPY --from=builder /bin/gateway /app/gateway
COPY --from=builder /bin/core /app/core
COPY --from=builder /bin/data /app/data
COPY config.yaml /app/config.yaml
COPY docs /app/docs
COPY static /app/static
RUN mkdir -p /app/logs \
 && chown -R appuser:appuser /app/logs
USER appuser
EXPOSE 8080 8090
ENTRYPOINT ["/app/ppo"]
