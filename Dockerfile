FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o api-security-scanner .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/api-security-scanner .
COPY --from=builder /app/config.yaml ./config.yaml
COPY --from=builder /app/config-test.yaml ./config-test.yaml

EXPOSE 8080 8081

CMD ["./api-security-scanner", "-dashboard"]