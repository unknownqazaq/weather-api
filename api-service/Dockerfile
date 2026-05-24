FROM golang:alpine AS builder

WORKDIR /app

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api-service ./cmd/app/main.go

# Production stage
FROM alpine:3.19
RUN apk --no-cache add ca-certificates curl

WORKDIR /root/

# Copy the binary from the build stage
COPY --from=builder /app/api-service .

EXPOSE 8080

CMD ["./api-service"]
