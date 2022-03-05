FROM golang:1.17-alpine AS builder

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go test /app/internal/handlers

RUN CGO_ENABLED=0 GOOS=linux go build -o server cmd/server/main.go

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/server /app/server

ENTRYPOINT ["/app/server"]