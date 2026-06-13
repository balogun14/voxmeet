FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY pkgs/go.mod pkgs/go.sum ./pkgs/
RUN cd pkgs && go mod download
COPY services/api-gateway/go.mod services/api-gateway/go.sum ./services/api-gateway/
RUN cd services/api-gateway && go mod download
COPY pkgs/ ./pkgs/
COPY services/api-gateway/ ./services/api-gateway/
RUN CGO_ENABLED=0 GOOS=linux go build -o /api-gateway ./services/api-gateway/cmd/server/

FROM alpine:3.21
RUN apk --no-cache add ca-certificates
COPY --from=builder /api-gateway /api-gateway
EXPOSE 8080
ENTRYPOINT ["/api-gateway"]
