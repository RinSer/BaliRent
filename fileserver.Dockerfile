FROM golang:alpine AS build-env

RUN mkdir /server

ADD . /server

WORKDIR /server

RUN go build -o main ./cmd/files/main.go

FROM alpine:latest

COPY --from=build-env /server/main .

CMD ["./main"]