FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY pkgs/go.mod pkgs/go.sum ./pkgs/
RUN cd pkgs && go mod download
COPY services/chat-service/go.mod services/chat-service/go.sum ./services/chat-service/
RUN cd services/chat-service && go mod download
COPY pkgs/ ./pkgs/
COPY services/chat-service/ ./services/chat-service/
RUN cd services/chat-service && CGO_ENABLED=0 GOOS=linux go build -o /chat-service ./cmd/server/

FROM alpine:3.21
RUN apk --no-cache add ca-certificates
COPY --from=builder /chat-service /chat-service
ENTRYPOINT ["/chat-service"]
