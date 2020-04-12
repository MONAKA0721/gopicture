FROM golang:alpine

RUN apk update \
  && apk add --no-cache git curl make gcc g++ \
  && go get github.com/pilu/fresh

WORKDIR /app
COPY go.mod .
COPY go.sum .

RUN go mod download
COPY ./ .

ENV DOCKERIZE_VERSION v0.6.0
RUN apk add --no-cache openssl \
 && wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
 && tar -C /usr/local/bin -xzvf dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
 && rm dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz

CMD fresh -c my_runner.conf
