FROM golang:1.12

WORKDIR /gifer

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN apt update && apt-get install ffmpeg -y

ENTRYPOINT ["go", "test", "-race", "."]

