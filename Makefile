.PHONY: build
build:
	@GOOS=js GOARCH=wasm go build -o static/out/main.wasm

.PHONY: serve
serve:
	@unset GOOS &&\
	unset GOARCH &&\
	go run cmd/server.go
