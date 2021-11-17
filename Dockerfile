FROM golang:1.17-alpine AS builder

RUN apk --no-cache add ca-certificates

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o benzinpriser main.go

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/benzinpriser /

ENTRYPOINT ["/benzinpriser"]