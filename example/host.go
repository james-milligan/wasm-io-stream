package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"

	"github.com/james-milligan/wasm-io-stream/io-stream/pkg/host"
)

//go:embed example.wasm
var exampleWasm []byte

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// create a new IoStreamHost instance passing the embeded wasm module, and any required startup arguments
	wasmWrapper, err := host.NewIoStreamClient(ctx, exampleWasm, "startup", "args")
	if err != nil {
		log.Panic(err)
	}
	// call the wasm module via stdin, all strings are json marshalled into a single stringified []string argument
	res, err := wasmWrapper.Call(ctx, "hello", "world", "foo", "bar")
	if err != nil {
		log.Panic(err)
	}
	// print the response
	fmt.Println(string(res))
}
