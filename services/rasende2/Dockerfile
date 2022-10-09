FROM golang:1.19-alpine AS builder

RUN apk --no-cache add ca-certificates

WORKDIR /build

RUN mkdir -p ./libs/common
RUN mkdir -p ./services/rasende2

COPY libs/common/go.mod libs/common/go.sum ./libs/common
WORKDIR /build/libs/common
RUN go mod download

WORKDIR /build
COPY services/rasende2/go.mod services/rasende2/go.sum ./services/rasende2
WORKDIR /build/services/rasende2
RUN go mod download

WORKDIR /build
COPY libs/common ./libs/common
COPY services/rasende2 ./services/rasende2

WORKDIR /build/services/rasende2
RUN CGO_ENABLED=0 GOOS=linux go build -o rasende2 .

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/services/rasende2/migrations /migrations
COPY --from=builder /build/services/rasende2/rss.json /rss.json
COPY --from=builder ["/build/services/rasende2/rasende2", "/build/.env*", "/"]

ENTRYPOINT ["/rasende2"]

