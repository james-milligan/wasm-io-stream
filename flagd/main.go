package main

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/james-milligan/wasm-io-stream/io-stream/pkg/host"
)

//go:embed flagd.wasm
var flagdWasm []byte

const flags = "{\n  \"flags\": {\n    \"myBoolFlag\": {\n      \"state\": \"ENABLED\",\n      \"variants\": {\n        \"on\": true,\n        \"off\": false\n      },\n      \"defaultVariant\": \"on\"\n    },\n    \"myStringFlag\": {\n      \"state\": \"ENABLED\",\n      \"variants\": {\n        \"key1\": \"val1\",\n        \"key2\": \"val2\"\n      },\n      \"defaultVariant\": \"key1\"\n    },\n    \"myFloatFlag\": {\n      \"state\": \"ENABLED\",\n      \"variants\": {\n        \"one\": 1.23,\n        \"two\": 2.34\n      },\n      \"defaultVariant\": \"one\"\n    },\n    \"myIntFlag\": {\n      \"state\": \"ENABLED\",\n      \"variants\": {\n        \"one\": 1,\n        \"two\": 2\n      },\n      \"defaultVariant\": \"one\"\n    },\n    \"myObjectFlag\": {\n      \"state\": \"ENABLED\",\n      \"variants\": {\n        \"object1\": {\n          \"key\": \"val\"\n        },\n        \"object2\": {\n          \"key\": true\n        }\n      },\n      \"defaultVariant\": \"object1\"\n    },\n    \"isColorYellow\": {\n      \"state\": \"ENABLED\",\n      \"variants\": {\n        \"on\": true,\n        \"off\": false\n      },\n      \"defaultVariant\": \"off\",\n      \"targeting\": {\n        \"if\": [\n          {\n            \"==\": [\n              {\n                \"var\": [\n                  \"color\"\n                ]\n              },\n              \"yellow\"\n            ]\n          },\n          \"on\",\n          \"off\"\n        ]\n      }\n    },\n    \"fibAlgo\": {\n      \"variants\": {\n        \"recursive\": \"recursive\",\n        \"memo\": \"memo\",\n        \"loop\": \"loop\",\n        \"binet\": \"binet\"\n      },\n      \"defaultVariant\": \"recursive\",\n      \"state\": \"ENABLED\",\n      \"targeting\": {\n        \"if\": [\n          {\n            \"$ref\": \"emailWithFaas\"\n          }, \"binet\", null\n        ]\n      }\n    },\n    \"headerColor\": {\n      \"variants\": {\n        \"red\": \"#FF0000\",\n        \"blue\": \"#0000FF\",\n        \"green\": \"#00FF00\",\n        \"yellow\": \"#FFFF00\"\n      },\n      \"defaultVariant\": \"red\",\n      \"state\": \"ENABLED\",\n      \"targeting\": {\n        \"if\": [\n          {\n            \"$ref\": \"emailWithFaas\"\n          },\n          {\n            \"fractionalEvaluation\": [\n              \"email\",\n              [\n                \"red\",\n                25\n              ],\n              [\n                \"blue\",\n                25\n              ],\n              [\n                \"green\",\n                25\n              ],\n              [\n                \"yellow\",\n                25\n              ]\n            ]\n          }, null\n        ]\n      }\n    }\n  },\n  \"$evaluators\": {\n    \"emailWithFaas\": {\n          \"in\": [\"@faas.com\", {\n            \"var\": [\"email\"]\n          }]\n    }\n  }\n}\n"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// create a new IoStreamHost instance passing the embedded wasm module, passing the flag configuration as a startup argument
	wasmWrapper, err := host.NewIoStreamClient(ctx, flagdWasm, flags)
	if err != nil {
		fmt.Println(err)
	}

	// call the flagd wasm module via stdin, all strings are json marshalled into a single stringified []string argument
	// RESOLVE_STRING RESOLVE_BOOLEAN RESOLVE_INT RESOLVE_FLOAT RESOLVE_OBJECT
	res, err := wasmWrapper.Call(ctx, "RESOLVE_BOOLEAN", "isColorYellow", `{"color":"yellow"}`)
	if err != nil {
		fmt.Println(err)
	}
	// print the response
	fmt.Println(string(res))
}
