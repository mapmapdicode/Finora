FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/wealthos-backend ./cmd/server

FROM alpine:3.21
WORKDIR /app
COPY --from=builder /bin/wealthos-backend /usr/local/bin/wealthos-backend
COPY --from=builder /app/internal/db/migrations /app/internal/db/migrations
EXPOSE 8080
ENTRYPOINT ["wealthos-backend"]

