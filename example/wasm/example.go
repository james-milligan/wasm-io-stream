package main

import (
	"encoding/json"
	"fmt"

	"github.com/james-milligan/flagd-wasm/io-stream/pkg/client"
)

func main() {
	client := client.NewIoStreamWasmClient()
	scanner := client.Scanner()
	for scanner.Scan() {
		args := []string{}
		if err := json.Unmarshal([]byte(scanner.Text()), &args); err != nil {
			client.SendError(err)
			continue
		}
		if len(args) == 0 {
			client.SendError(fmt.Errorf("no arguments provided to wasm module"))
			continue
		}
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
