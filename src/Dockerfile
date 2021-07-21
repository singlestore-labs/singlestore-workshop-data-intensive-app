FROM golang:1.15.14-buster as builder

WORKDIR /go/src/app

# cache dependencies to accelerate most rebuilds
COPY go.mod go.sum ./
RUN go mod download all

COPY . .

RUN go build -o /simulator bin/simulator/main.go

FROM debian:buster-slim

CMD ["./simulator", "--config", "config.toml"]
COPY --from=builder /simulator .
COPY config-sim.toml /config.toml