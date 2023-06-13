package test

import (
	"encoding/json"
	"fmt"

	"github.com/james-milligan/flagd-wasm/io-stream/pkg/client"
)

func main() {
	client := client.NewIoStreamClient()
	scanner := client.Scanner()
	for scanner.Scan() {
		args := []string{}
		if err := json.Unmarshal([]byte(scanner.Text()), &args); err != nil {
			client.SendError(err)
			continue
		}
		if len(args) != 2 {
			client.SendError(fmt.Errorf("unexpected number of args, got %d, want 2", len(args)))
			continue
		}
		client.SendResponse("arg0: ", args[0], "arg1:", args[1])
	}

	if err := scanner.Err(); err != nil {
		client.SendError(err)
	}

}
