# ==================== Go API Service ====================
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /api .
COPY etc ./etc

ENV TZ=Asia/Shanghai

EXPOSE 8080

CMD ["./api", "-f", "etc/ecommerce.api"]
