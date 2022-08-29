FROM golang:1.18-alpine AS builder

RUN apk --no-cache add ca-certificates

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o shoppinglist-backend-api cmd/api/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o shoppinglist-backend-migrate cmd/dbmigrate/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o shoppinglist-backend-worker cmd/worker/main.go

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/db/migrations /db/migrations
COPY --from=builder ["/build/shoppinglist-backend-api", "/build/shoppinglist-backend-migrate", "/build/shoppinglist-backend-worker", "/build/.env*", "/"]

ENTRYPOINT ["/shoppinglist-backend-api"]

