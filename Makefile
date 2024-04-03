.PHONY: build
build:
	@GOOS=js GOARCH=wasm go build -o static/out/zip.wasm

.PHONY: serve
serve:
	@unset GOOS &&\
	unset GOARCH &&\
	go run cmd/server.go
