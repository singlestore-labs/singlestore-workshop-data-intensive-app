FROM golang:1.15.14-buster as builder

WORKDIR /go/src/app

# cache dependencies to accelerate most rebuilds
COPY go.mod go.sum ./
RUN go mod download all

COPY . .

RUN go build -o /simulator bin/simulator/main.go
RUN go build -o /api bin/api/main.go

# simulator image
FROM debian:buster-slim as simulator
ADD https://curl.haxx.se/ca/cacert.pem /etc/pki/tls/certs/ca-bundle.crt

CMD ["./simulator", "--config", "config.toml"]
COPY --from=builder /simulator .
COPY config-sim.toml /config.toml

# api image
FROM debian:buster-slim as api
ADD https://curl.haxx.se/ca/cacert.pem /etc/pki/tls/certs/ca-bundle.crt

CMD ["./api", "--config", "config.toml"]
COPY --from=builder /api .
COPY config-api.toml /config.toml