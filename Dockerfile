FROM golang:1.24-alpine3.20 AS builder

WORKDIR /root

COPY . ./

RUN go build -o bin/barong-users-manager cmd/main.go

FROM alpine:3.20

ARG UID=1000
ARG GID=1000

RUN addgroup -g ${GID} -S app && adduser -u ${UID} -S -G app app

WORKDIR /home/app

COPY --from=builder /root/bin/barong-users-manager ./

RUN chown app:app ./barong-users-manager && chmod +x ./barong-users-manager

USER app

CMD ["./barong-users-manager"]
