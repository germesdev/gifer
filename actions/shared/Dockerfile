FROM golang:1.12

LABEL "com.github.actions.name"="Image with golang and ffmpeg binary"
LABEL "com.github.actions.description"=""
LABEL "com.github.actions.icon"="package"
LABEL "com.github.actions.color"="blue"

LABEL "repository"="https://github.com/germesdev/gifer/actions/shared"
LABEL "maintainer"="Philidor Green <shindu666@gmail.com>"

RUN apt update && apt install -y ca-certificates ffmpeg && rm -rf /var/lib/apt/lists/*

COPY entrypoint.sh /entrypoint.sh

RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]



