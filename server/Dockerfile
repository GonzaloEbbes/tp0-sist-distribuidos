FROM golang:1.17 AS builder

WORKDIR /src

COPY go.mod go.sum ./
COPY vendor ./vendor
COPY server ./server

RUN GOOS=linux go test ./server/common
RUN GOOS=linux go build -mod=vendor -o /server github.com/7574-sistemas-distribuidos/docker-compose-init/server

FROM busybox

COPY --from=builder /server /server

ENTRYPOINT ["/bin/sh"]
