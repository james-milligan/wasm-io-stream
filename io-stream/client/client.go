package client

import (
	"bufio"
	"fmt"
	"os"
)

const (
	ModuleReadyString = "MODULE_READY\n"
)

type IoStreamClient struct{}

func NewIoStreamClient() *IoStreamClient {
	fmt.Printf(ModuleReadyString)
	return &IoStreamClient{}
}

func (i *IoStreamClient) SendResponse(res ...any) {
	fmt.Println(res...)
}

func (i *IoStreamClient) SendError(err error) {
	fmt.Fprintln(os.Stderr, err)
}

func (i *IoStreamClient) Scanner() *bufio.Scanner {
	return bufio.NewScanner(os.Stdin)
}
