FROM golang:1.18-alpine AS builder

RUN apk --no-cache add ca-certificates

WORKDIR /build

COPY services/shoppinglist-backend/go.mod services/shoppinglist-backend/go.sum ./
RUN go mod download

COPY services/shoppinglist-backend ./services/shoppinglist-backend

WORKDIR /build/services/shoppinglist-backend
RUN CGO_ENABLED=0 GOOS=linux go build -o shoppinglist-backend-api cmd/api/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o shoppinglist-backend-migrate cmd/dbmigrate/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o shoppinglist-backend-worker cmd/worker/main.go

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/services/shoppinglist-backend/db/migrations /db/migrations
COPY --from=builder ["/build/services/shoppinglist-backend/shoppinglist-backend-api", "/build/services/shoppinglist-backend/shoppinglist-backend-migrate", "/build/services/shoppinglist-backend/shoppinglist-backend-worker", "/build/services/shoppinglist-backend/.env*", "/"]

ENTRYPOINT ["/shoppinglist-backend-api"]

