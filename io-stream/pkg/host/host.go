package host

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/james-milligan/wasm-io-stream/io-stream/pkg/client"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/tetratelabs/wazero/sys"
)

const (
	ModuleReadyString = "MODULE_READY\n"
)

type IoStreamClient struct {
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

func NewIoStreamClient(ctx context.Context, module []byte, args ...string) (*IoStreamClient, error) {
	ready := make(chan struct{})
	errChan := make(chan error, 1)

	wasmWrapper := &IoStreamClient{
		mu:       &sync.Mutex{},
		errChan:  errChan,
		dataChan: make(chan string, 1),
	}

	go wasmWrapper.init(ctx, module, ready, args)

	select {
	case <-ready:
		return wasmWrapper, nil
	case err := <-errChan:
		return nil, err
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("init timed out after %d seconds", 5)
	}
}

func (i *IoStreamClient) init(ctx context.Context, module []byte, ready chan struct{}, args []string) {
	r := wazero.NewRuntime(ctx)

	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()
	stderrReader, stderrWriter := io.Pipe()
	i.ioConfig = &ioConfig{
		stdinReader:  stdinReader,
		stdinWriter:  stdinWriter,
		stdoutReader: stdoutReader,
		stdoutWriter: stdoutWriter,
		stderrReader: stderrReader,
		stderrWriter: stderrWriter,
	}

	go func() {
		<-ctx.Done()
		i.mu.Lock()
		defer i.mu.Unlock()
		r.Close(ctx)
		i.ioConfig.stdinReader.Close()
		i.ioConfig.stdinWriter.Close()
		i.ioConfig.stdoutReader.Close()
		i.ioConfig.stdoutWriter.Close()
		i.ioConfig.stderrReader.Close()
		i.ioConfig.stderrWriter.Close()
	}()

	// Instantiate WASI, which implements system I/O such as console output.
	wasi_snapshot_preview1.MustInstantiate(ctx, r)

	// Read from stderrReader in a separate goroutine
	go func() {
		reader := bufio.NewReader(i.ioConfig.stderrReader)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				} else {
					i.errChan <- fmt.Errorf("Error reading from stderr: %w", err)
				}
			}
			i.errChan <- fmt.Errorf(line[:len(line)-1]) // trim newline
		}
	}()

	// Read from stdoutReader in a separate goroutine
	go func() {
		reader := bufio.NewReader(i.ioConfig.stdoutReader)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				} else {
					i.errChan <- fmt.Errorf("Error reading from stdout: %w", err)
				}
			}
			if line == client.ModuleReadyString {
				close(ready)
				continue
			}
			i.dataChan <- line[:len(line)-1] // trim newline
		}
	}()

	// InstantiateModule runs the "_start" function, WASI's "main".
	_, err := r.InstantiateWithConfig(ctx, module, wazero.NewModuleConfig().WithArgs(args...).WithStdout(i.ioConfig.stdoutWriter).WithStdin(i.ioConfig.stdinReader).WithStderr(i.ioConfig.stderrWriter))
	if err != nil {
		// Note: Most compilers do not exit the module after running "_start",
		// unless there was an error. This allows you to call exported functions.
		if exitErr, ok := err.(*sys.ExitError); ok && exitErr.ExitCode() != 0 {
			i.errChan <- fmt.Errorf("exit_code: %d\n", exitErr.ExitCode())
			return
		} else if !ok {
			i.errChan <- fmt.Errorf("InstantiateWithConfig: %w", err)
			return
		}
	}
}
