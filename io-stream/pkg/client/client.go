package client

import (
	"bufio"
	"fmt"
	"os"
)

const (
	ModuleReadyString = "MODULE_READY\n"
)

type IoStreamWasmClient struct{}

func NewIoStreamWasmClient() *IoStreamWasmClient {
	fmt.Printf(ModuleReadyString)
	return &IoStreamWasmClient{}
}

func (i *IoStreamWasmClient) StartupArgs() []string {
	return os.Args
}

func (i *IoStreamWasmClient) SendResponse(res ...any) {
	fmt.Println(res...)
}

func (i *IoStreamWasmClient) SendError(err error) {
	fmt.Fprintln(os.Stderr, err)
}

func (i *IoStreamWasmClient) Scanner() *bufio.Scanner {
	return bufio.NewScanner(os.Stdin)
}
