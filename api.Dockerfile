FROM golang

RUN mkdir /app

ADD . /app

WORKDIR /app

RUN go build -o main ./cmd/server/main.go

CMD [ "/app/main" ]