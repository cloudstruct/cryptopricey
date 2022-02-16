# syntax=docker/dockerfile:1

## Build
FROM golang:1.17-alpine as build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download \
  && mkdir /app/assets

COPY *.go ./
COPY assets/* /app/assets/

RUN go build -o /cryptopricey

## Deploy
# hadolint ignore=DL3006
FROM alpine:3

WORKDIR /

COPY --from=build /cryptopricey /cryptopricey

RUN chmod 0755 /cryptopricey

EXPOSE 8080

CMD ["/cryptopricey"]
