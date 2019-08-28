FROM golang:1.12 AS build-env

WORKDIR /gifer

RUN apt-get update && apt-get install curl git

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN apt-get install ffmpeg -y

ENTRYPOINT ["go", "test", "-race", "."]

