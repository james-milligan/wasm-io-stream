build-example:
	cd example/wasm; GOARCH=wasm GOOS=wasip1 gotip build -o ../example.wasm example.go;
run-example: build-example
	gotip run example/host.go

build-flagd:
	cd flagd/wasm; GOARCH=wasm GOOS=wasip1 gotip build -o ../flagd.wasm flagd.go
run-flagd: build-flagd
	gotip run flagd/main.go