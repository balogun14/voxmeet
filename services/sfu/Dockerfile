FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY pkgs/go.mod pkgs/go.sum ./pkgs/
RUN cd pkgs && go mod download
COPY services/sfu/go.mod services/sfu/go.sum ./services/sfu/
RUN cd services/sfu && go mod download
COPY pkgs/ ./pkgs/
COPY services/sfu/ ./services/sfu/
RUN cd services/sfu && CGO_ENABLED=0 GOOS=linux go build -o /sfu ./cmd/server/

FROM alpine:3.21
RUN apk --no-cache add ca-certificates
COPY --from=builder /sfu /sfu
ENTRYPOINT ["/sfu"]
