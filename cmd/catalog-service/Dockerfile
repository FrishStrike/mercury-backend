# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

# Download dependencies first (better caching)
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

# Build with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -extldflags '-static'" \
    -o bin/catalog-service \
    ./cmd/catalog-service

# Run stage - distroless for security
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

# Copy binary and swagger
COPY --from=builder /app/bin/catalog-service .
COPY --from=builder /app/cmd/catalog-service/swagger ./swagger

EXPOSE 50051 8080

USER nonroot:nonroot

CMD ["./catalog-service"]