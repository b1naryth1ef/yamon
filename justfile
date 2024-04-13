build-yamon-server:
	go build -o yamon-server cmd/yamon-server/main.go

build-yamon-agent:
	go build -o yamon-agent cmd/yamon-agent/main.go

build: build-yamon-server build-yamon-agent