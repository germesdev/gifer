FROM golang:1.12

LABEL "com.github.actions.name"="Build image with ffmpeg"
LABEL "com.github.actions.description"="Check if application able to run on machine with ffmpeg"
LABEL "com.github.actions.icon"="package"
LABEL "com.github.actions.color"="blue"

LABEL "repository"="https://github.com/germesdev/gifer/actions/test"
LABEL "maintainer"="Philidor Green <shindu666@gmail.com>"

RUN apt update && apt install -y ca-certificates ffmpeg && rm -rf /var/lib/apt/lists/*

COPY entrypoint.sh /entrypoint.sh

RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]


