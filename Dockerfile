FROM golang:1.24-alpine

RUN mkdir -p /usr/src/
WORKDIR /usr/src/

COPY go.mod go.sum /usr/src/yamon/

WORKDIR /usr/src/yamon
RUN go mod download

COPY . /usr/src/yamon/

ENV GOCACHE=/var/cache/go-build
RUN --mount=type=cache,target="/var/cache/go-build" go build -v -o /bin/yamon-agent cmd/yamon-agent/main.go
RUN --mount=type=cache,target="/var/cache/go-build" go build -v -o /bin/yamon-server cmd/yamon-server/main.go
RUN --mount=type=cache,target="/var/cache/go-build" go build -v -o /bin/yamon-debug cmd/yamon-debug/main.go

WORKDIR /opt
ENTRYPOINT ["/bin/yamon-agent"]
