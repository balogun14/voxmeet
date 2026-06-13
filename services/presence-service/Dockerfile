FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY pkgs/go.mod pkgs/go.sum ./pkgs/
RUN cd pkgs && go mod download
COPY services/presence-service/go.mod services/presence-service/go.sum ./services/presence-service/
RUN cd services/presence-service && go mod download
COPY pkgs/ ./pkgs/
COPY services/presence-service/ ./services/presence-service/
RUN CGO_ENABLED=0 GOOS=linux go build -o /presence-service ./services/presence-service/cmd/server/

FROM alpine:3.21
RUN apk --no-cache add ca-certificates
COPY --from=builder /presence-service /presence-service
ENTRYPOINT ["/presence-service"]
