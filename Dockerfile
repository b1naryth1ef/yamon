FROM golang:1.23-alpine

RUN mkdir -p /usr/src/
WORKDIR /usr/src/

COPY go.mod go.sum /usr/src/yamon/

WORKDIR /usr/src/yamon
RUN go mod download

COPY . /usr/src/yamon/

RUN go build -v -o /bin/yamon-agent cmd/yamon-agent/main.go
RUN go build -v -o /bin/yamon-server cmd/yamon-server/main.go
RUN go build -v -o /bin/yamon-debug cmd/yamon-debug/main.go

WORKDIR /opt
ENTRYPOINT ["/bin/yamon-agent"]
