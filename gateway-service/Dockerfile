FROM golang:alpine AS builder

WORKDIR /app

# Download dependencies
COPY go.mod ./
# If go.sum is not present or empty, we will skip it or copy if it exists.
# We will download dependencies.
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gateway-service ./cmd/main.go

# Production stage
FROM alpine:3.19
RUN apk --no-cache add ca-certificates curl

WORKDIR /root/

# Copy the binary from the build stage
COPY --from=builder /app/gateway-service .

EXPOSE 8081

CMD ["./gateway-service"]
