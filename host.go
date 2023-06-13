package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	wasmWrapper "github.com/james-milligan/flagd-wasm/wasm-wrapper"
)

//go:embed evaluator.wasm
var evaluatorWasm []byte

type WasmWrapper struct {
	ioConfig *ioConfig
	mu       *sync.Mutex
	errChan  chan error
	dataChan chan string
}

type ioConfig struct {
	stdinReader  *io.PipeReader
	stdinWriter  *io.PipeWriter
	stdoutReader *io.PipeReader
	stdoutWriter *io.PipeWriter
	stderrReader *io.PipeReader
	stderrWriter *io.PipeWriter
}

func main() {
	// Choose the context to use for function calls.
	ctx, close := context.WithCancel(context.Background())
	wasmWrapper, err := wasmWrapper.NewWasmWrapper(ctx, evaluatorWasm)
	if err != nil {
		log.Panic(err)
	}
	res, err := wasmWrapper.Call(ctx, "hello", "world")
	if err != nil {
		log.Panic(err)
	}
	response := []string{}
	err = json.Unmarshal(res, &response)
	fmt.Println(err, response)

	close()
	time.Sleep(1 * time.Second)
	res, err = wasmWrapper.Call(ctx, "hello", "world")
	if err != nil {
		log.Panic(err)
	}
	response = []string{}
	err = json.Unmarshal(res, &response)
	fmt.Println(err, response)
}
