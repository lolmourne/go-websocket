go-build:
	@go build -v -o go-websocket-app *.go

go-run:
	@./go-websocket-app

run:
	make go-build
	make go-run