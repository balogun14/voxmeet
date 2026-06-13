FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY pkgs/go.mod pkgs/go.sum ./pkgs/
RUN cd pkgs && go mod download
COPY services/room-service/go.mod services/room-service/go.sum ./services/room-service/
RUN cd services/room-service && go mod download
COPY pkgs/ ./pkgs/
COPY services/room-service/ ./services/room-service/
RUN cd services/room-service && CGO_ENABLED=0 GOOS=linux go build -o /room-service ./cmd/server/

FROM alpine:3.21
RUN apk --no-cache add ca-certificates
COPY --from=builder /room-service /room-service
ENTRYPOINT ["/room-service"]
