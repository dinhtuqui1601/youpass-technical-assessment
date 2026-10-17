# ---- Build stage ----
FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api ./cmd/api \
 && CGO_ENABLED=0 GOOS=linux go build -o /out/worker ./cmd/worker

# ---- Runtime stage ----
FROM alpine:3.21

RUN adduser -D -u 10001 app
USER app
WORKDIR /app

COPY --from=builder /out/api /app/api
COPY --from=builder /out/worker /app/worker

EXPOSE 8080

CMD ["/app/api"]
