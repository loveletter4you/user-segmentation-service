FROM golang:1.21-alpine3.18 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /main ./cmd/app/app.go

FROM scratch
COPY --from=builder /app/config /config
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder main /bin/main


ENTRYPOINT ["/bin/main"]