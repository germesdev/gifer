FROM golang:1.12-alpine AS build-env

RUN apk update && apk add --no-cache curl git

WORKDIR /gifer

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN go build -ldflags="-s -w" -o gifer

FROM ubuntu
WORKDIR /app

# RUN apk update && apk add ca-certificates ffmpeg && rm -rf /var/cache/apk/*
RUN apt update && apt install -y ca-certificates ffmpeg && rm -rf /var/lib/apt/lists/*

COPY --from=build-env /gifer/gifer /bin/

ENTRYPOINT ["gifer"]
