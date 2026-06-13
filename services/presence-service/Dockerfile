FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /presence-service ./cmd/server/

FROM alpine:3.21
RUN apk --no-cache add ca-certificates
COPY --from=builder /presence-service /presence-service
ENTRYPOINT ["/presence-service"]
