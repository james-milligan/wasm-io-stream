package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/james-milligan/wasm-io-stream/io-stream/pkg/client"
	"github.com/open-feature/flagd/core/pkg/eval"
	"github.com/open-feature/flagd/core/pkg/logger"
	"github.com/open-feature/flagd/core/pkg/store"
	"github.com/open-feature/flagd/core/pkg/sync"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	CmdResolveBoolean    = "RESOLVE_BOOLEAN"
	CmdResolveString     = "RESOLVE_STRING"
	CmdResolveInt        = "RESOLVE_INT"
	CmdResolveFloat      = "RESOLVE_FLOAT"
	CmdResolveObject     = "RESOLVE_OBJECT"
	CmdFlagConfiguration = "FLAG_CONFIGURATION"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// create a new IoStreamWasmClient
	client := client.NewIoStreamWasmClient()
	args := client.StartupArgs()

	eval := eval.NewJSONEvaluator(logger.NewLogger(nil, false), store.NewFlags())
	_, _, err := eval.SetState(sync.DataSync{
		FlagData: args[0],
		Source:   "wasm",
		Type:     sync.ALL,
	})
	if err != nil {
		client.SendError(err)
	}

	// create a new *bufio.Scanner from the client
	scanner := client.Scanner()
	for scanner.Scan() {
		args := []string{}
		if err := json.Unmarshal([]byte(scanner.Text()), &args); err != nil {
			// if the message cannot be parsed into []string return an error using the client.SendError method
			client.SendError(err)
			continue
		}
		if len(args) == 2 && args[0] == CmdFlagConfiguration {
			_, _, err = eval.SetState(sync.DataSync{
				FlagData: args[1],
				Source:   "wasm",
				Type:     sync.ALL,
			})
			if err != nil {
				client.SendError(err)
			} else {
				client.SendResponse("flag configuration set successfully")
			}
			continue
		}
		if len(args) != 3 {
			// if the arguments slice is empty, return an error using the client.SendError method
			client.SendError(fmt.Errorf("wrong number of arguments provided to wasm module, expected 2, got %d", len(args)))
			continue
		}
		evalCtx, err := stringToStruct(args[2])
		if err != nil {
			client.SendError(err)
			continue
		}
		switch args[0] {
		case CmdResolveBoolean:
			value, variant, reason, err := eval.ResolveBooleanValue(ctx, "", args[1], evalCtx)
			handleEvaluationResponse(client, value, variant, reason, err)
		case CmdResolveString:
			value, variant, reason, err := eval.ResolveStringValue(ctx, "", args[1], evalCtx)
			handleEvaluationResponse(client, value, variant, reason, err)
		case CmdResolveInt:
			value, variant, reason, err := eval.ResolveIntValue(ctx, "", args[1], evalCtx)
			handleEvaluationResponse(client, value, variant, reason, err)
		case CmdResolveFloat:
			value, variant, reason, err := eval.ResolveFloatValue(ctx, "", args[1], evalCtx)
			handleEvaluationResponse(client, value, variant, reason, err)
		case CmdResolveObject:
			value, variant, reason, err := eval.ResolveObjectValue(ctx, "", args[1], evalCtx)
			handleEvaluationResponse(client, value, variant, reason, err)
		default:
			client.SendError(fmt.Errorf("unrecognized command string at position 0: %s", args[0]))
		}
	}

	if err := scanner.Err(); err != nil {
		client.SendError(err)
	}
}

func handleEvaluationResponse(client *client.IoStreamWasmClient, value interface{}, variant string, reason string, err error) {
	if err != nil {
		client.SendError(fmt.Errorf("unable to evaluate flag: %w", err))
		return
	}
	res := map[string]interface{}{
		"value":   value,
		"variant": variant,
		"reason":  reason,
	}
	b, _ := json.Marshal(res)
	client.SendResponse(string(b))
}

func stringToStruct(in string) (*structpb.Struct, error) {
	object := map[string]interface{}{}
	if err := json.Unmarshal([]byte(in), &object); err != nil {
		return nil, err
	}
	return structpb.NewStruct(object)
}
