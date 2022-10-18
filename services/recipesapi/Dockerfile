FROM golang:1.19-alpine AS builder

RUN apk --no-cache add ca-certificates

WORKDIR /build

RUN mkdir -p ./libs/common
RUN mkdir -p ./services/recipesapi

COPY libs/common/go.mod libs/common/go.sum ./libs/common
WORKDIR /build/libs/common
RUN go mod download

WORKDIR /build
COPY services/recipesapi/go.mod services/recipesapi/go.sum ./services/recipesapi
WORKDIR /build/services/recipesapi
RUN go mod download

WORKDIR /build
COPY libs/common ./libs/common
COPY services/recipesapi ./services/recipesapi

WORKDIR /build/services/recipesapi
RUN CGO_ENABLED=0 GOOS=linux go build -o recipesapi .

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# COPY --from=builder /build/services/recipesapi/migrations /migrations
COPY --from=builder ["/build/services/recipesapi/recipesapi", "/build/.env*", "/"]

ENTRYPOINT ["/recipesapi"]

