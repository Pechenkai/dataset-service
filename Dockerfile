FROM golang:1.21-alpine AS builder
ENV GOTOOLCHAIN=auto
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/ppo ./cmd/api

FROM alpine:3.19
RUN adduser -D -g '' appuser
WORKDIR /app
COPY --from=builder /bin/ppo /app/ppo
COPY config.yaml /app/config.yaml
USER appuser
EXPOSE 8080 8090
ENTRYPOINT ["/app/ppo"]
