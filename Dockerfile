FROM golang:1.24.7-alpine3.22 AS builder

ENV GOOS=linux
ENV CGO_ENABLED=0

WORKDIR /app

RUN apk update
RUN apk add git

COPY go.mod go.mod
RUN go mod download

COPY . .

RUN go build -buildvcs -o /app/api-server ./cmd/api-server

FROM alpine:3.22

RUN mkdir /app

RUN adduser -S -D -H -h /app app
RUN chown app /app

USER app

COPY --from=builder /app/api-server /app/
