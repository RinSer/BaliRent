FROM golang

RUN mkdir /server

ADD . /server

WORKDIR /server

RUN go build -o main ./cmd/files/main.go

CMD [ "/server/main" ]