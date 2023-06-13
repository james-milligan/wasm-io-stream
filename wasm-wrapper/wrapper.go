package wrapper

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/tetratelabs/wazero/sys"
)

const (
	ModuleReadyString = "MODULE_READY\n"
)

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

func NewWasmWrapper(ctx context.Context, module []byte) (*WasmWrapper, error) {
	ready := make(chan struct{})
	errChan := make(chan error)
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()
	stderrReader, stderrWriter := io.Pipe()
	wasmWrapper := &WasmWrapper{
		ioConfig: &ioConfig{
			stdinReader:  stdinReader,
			stdinWriter:  stdinWriter,
			stdoutReader: stdoutReader,
			stdoutWriter: stdoutWriter,
			stderrReader: stderrReader,
			stderrWriter: stderrWriter,
		},
		mu:       &sync.Mutex{},
		errChan:  errChan,
		dataChan: make(chan string, 1),
	}
	go wasmWrapper.init(ctx, module, ready, errChan)

	select {
	case <-ready:
		return wasmWrapper, nil
	case err := <-errChan:
		return nil, err
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("timed out after %d seconds", 5)
	}
}

func (w *WasmWrapper) init(ctx context.Context, module []byte, ready chan struct{}, errChan chan error) {
	r := wazero.NewRuntime(ctx)
	go func() {
		<-ctx.Done()
		w.mu.Lock()
		defer w.mu.Unlock()
		r.Close(ctx)
		w.ioConfig.stdinReader.Close()
		w.ioConfig.stdinWriter.Close()
		w.ioConfig.stdoutReader.Close()
		w.ioConfig.stdoutWriter.Close()
		w.ioConfig.stderrReader.Close()
		w.ioConfig.stderrWriter.Close()
	}()

	// Instantiate WASI, which implements system I/O such as console output.
	wasi_snapshot_preview1.MustInstantiate(ctx, r)

	// Read from stderrReader in a separate goroutine
	go func() {
		reader := bufio.NewReader(w.ioConfig.stderrReader)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				} else {
					w.errChan <- fmt.Errorf("Error reading from stderr: %w", err)
				}
			}
			w.errChan <- fmt.Errorf(line)
		}
	}()

	// Read from stdoutReader in a separate goroutine
	go func() {
		reader := bufio.NewReader(w.ioConfig.stdoutReader)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				} else {
					w.errChan <- fmt.Errorf("Error reading from stdout: %w", err)
				}
			}
			if line == ModuleReadyString {
				close(ready)
				continue
			}
			w.dataChan <- line
		}
	}()

	// InstantiateModule runs the "_start" function, WASI's "main".
	_, err := r.InstantiateWithConfig(ctx, module, wazero.NewModuleConfig().WithArgs("wasi",
		"2").WithStdout(w.ioConfig.stdoutWriter).WithStdin(w.ioConfig.stdinReader).WithStderr(w.ioConfig.stderrWriter))
	if err != nil {
		// Note: Most compilers do not exit the module after running "_start",
		// unless there was an error. This allows you to call exported functions.
		if exitErr, ok := err.(*sys.ExitError); ok && exitErr.ExitCode() != 0 {
			errChan <- fmt.Errorf("exit_code: %d\n", exitErr.ExitCode())
			return
		} else if !ok {
			errChan <- err
			return
		}
	}
}
