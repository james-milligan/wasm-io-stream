# wasm-io-stream

Implementation for bidirectional communication with WASM modules via `stdin`, `stdout` and `stderr`.  

Notes:
- you cannot log within the wasm module, as it will be interpreted as the modules response

## Usage

### Host implementation:
```go
package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"

	"github.com/james-milligan/flagd-wasm/io-stream/pkg/host"
)

//go:embed example.wasm
var exampleWasm []byte

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
    // create a new IoStreamHost instance passing the embeded wasm module, and any required startup arguments
	wasmWrapper, err := host.NewIoStreamHost(ctx, exampleWasm, "startup", "args")
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
```

Wasm module implementation:
```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/james-milligan/flagd-wasm/io-stream/pkg/client"
)

func main() {
    // create a new IoStreamWasmClient
	client := client.NewIoStreamWasmClient()
    // create a new *bufio.Scanner from the client
	scanner := client.Scanner()
	for scanner.Scan() {
		args := []string{}
		if err := json.Unmarshal([]byte(scanner.Text()), &args); err != nil {
            // if the message cannot be parsed into []string return an error using the client.SendError method
			client.SendError(err)
			continue
		}
		if len(args) == 0 {
            // if the arguments slice is empty, return an error using the client.SendError method
			client.SendError(fmt.Errorf("no arguments provided to wasm module"))
			continue
		}
        // create a new response string and return it to the host using the client.SendResponse method
		res := ""
		for x, arg := range args {
			res = fmt.Sprintf("%s(arg%d: %s) ", res, x, arg)
		}
		client.SendResponse(res)
	}

	if err := scanner.Err(); err != nil {
		client.SendError(err)
	}
}
```

## Run the example:

The example can be found in `/example`, containing a simple wasm module and host application.

Run example:
```bash
make run-example
```

Expected output:
```
(arg0: hello) (arg1: world) (arg2: foo) (arg3: bar) 
```