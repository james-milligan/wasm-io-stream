package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"time"

	"github.com/james-milligan/flagd-wasm/io-stream/pkg/host"
)

//go:embed example.wasm
var exampleWasm []byte

func main() {
	// Choose the context to use for function calls.
	ctx, _ := context.WithCancel(context.Background())
	wasmWrapper, err := host.NewIoStreamHost(ctx, exampleWasm, "flags", "test")
	if err != nil {
		log.Panic(err)
	}
	res, err := wasmWrapper.Call(ctx, "LOG", "world")
	if err != nil {
		log.Panic(err)
	}
	fmt.Println(string(res))

	time.Sleep(1 * time.Second)
	// res, err = wasmWrapper.Call(ctx, "ERROR", "world")
	// if err != nil {
	// 	fmt.Println("error:", err)
	// }
	// fmt.Println(res)
}
