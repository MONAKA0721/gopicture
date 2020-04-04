FROM golang:alpine as builder

RUN apk update \
  && apk add --no-cache git curl make gcc g++ \
  && go get github.com/oxequa/realize

WORKDIR /app
COPY go.mod .
COPY go.sum .

RUN go mod download
COPY ./ .

RUN GOOS=linux GOARCH=amd64 go build app.go

FROM alpine

COPY --from=builder /app/ /app

WORKDIR /app
CMD /app/app
