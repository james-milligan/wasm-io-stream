build:
	cd example/wasm; GOARCH=wasm GOOS=wasip1 gotip build -o ../example.wasm example.go;
run-example: build
	gotip run example/host.go