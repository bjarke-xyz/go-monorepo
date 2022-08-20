FROM golang:1.19-alpine AS builder

RUN apk --no-cache add ca-certificates

WORKDIR /build

RUN mkdir -p ./libs/common
RUN mkdir -p ./services/fuelpricesapi

COPY libs/common/go.mod libs/common/go.sum ./libs/common
WORKDIR /build/libs/common
RUN go mod download

WORKDIR /build
COPY services/fuelpricesapi/go.mod services/fuelpricesapi/go.sum ./services/fuelpricesapi
WORKDIR /build/services/fuelpricesapi
RUN go mod download

WORKDIR /build
COPY libs/common ./libs/common
COPY services/fuelpricesapi ./services/fuelpricesapi

WORKDIR /build/services/fuelpricesapi
RUN CGO_ENABLED=0 GOOS=linux go build -o fuelpricesapi .

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/services/fuelpricesapi/migrations /migrations
COPY --from=builder ["/build/services/fuelpricesapi/fuelpricesapi", "/build/.env*", "/"]

ENTRYPOINT ["/fuelpricesapi"]

