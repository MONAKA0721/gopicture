FROM golang:alpine

RUN apk update \
  && apk add --no-cache git curl make gcc g++ \
  && go get github.com/pilu/fresh

WORKDIR /app
COPY go.mod .
COPY go.sum .

RUN go mod download
COPY ./ .

CMD fresh -c my_runner.conf
