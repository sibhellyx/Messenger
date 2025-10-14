
# Build stage
FROM golang:1.24-alpine AS builder
RUN apk add --no-cache git ca-certificates tzdata
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/app
# Runtime stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -S app && adduser -S app -G app
RUN mkdir -p /app/configs && chown -R app:app /app
USER app
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/configs/ ./configs/
EXPOSE 8080
CMD ["./main"]