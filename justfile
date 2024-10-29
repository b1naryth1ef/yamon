build-yamon-server:
	go build -o yamon-server cmd/yamon-server/main.go

build-yamon-agent:
	go build -o yamon-agent cmd/yamon-agent/main.go

build-yamon-debug:
	go build -o yamon-debug cmd/yamon-debug/main.go

build-docker:
	docker build -t yamon:local-dev .

build: build-yamon-server build-yamon-agent build-yamon-debug