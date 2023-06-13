build-example:
	cd example/wasm; GOARCH=wasm GOOS=wasip1 gotip build -o ../example.wasm example.go;
run-example: build-example
	gotip run example/host.go