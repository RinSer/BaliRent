FROM golang:alpine AS build-env

RUN mkdir /app

ADD . /app

WORKDIR /app

RUN go build -o main ./cmd/server/main.go

FROM alpine:latest

COPY --from=build-env /app/main .

CMD [ "./main"  ]