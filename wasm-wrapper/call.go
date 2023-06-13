package wrapper

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

func (w *WasmWrapper) Call(ctx context.Context, args ...string) ([]byte, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	b, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}
	b = append(b, []byte("\n")...)
	_, err = w.ioConfig.stdinWriter.Write(b)
	if err != nil {
		return nil, err
	}

	select {
	case err := <-w.errChan:
		return nil, err
	case data := <-w.dataChan:
		return []byte(data), nil
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("timed out after %d seconds", 5)
	}
}
