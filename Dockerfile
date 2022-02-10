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
FROM gcr.io/distroless/base-debian11

WORKDIR /

COPY --from=build /cryptopricey /cryptopricey

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/cryptopricey"]
