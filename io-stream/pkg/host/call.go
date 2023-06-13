package host

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

func (i *IoStreamClient) Call(ctx context.Context, args ...string) ([]byte, error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	b, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}
	b = append(b, []byte("\n")...)
	_, err = i.ioConfig.stdinWriter.Write(b)
	if err != nil {
		return nil, err
	}

	select {
	case err := <-i.errChan:
		return nil, err
	case data := <-i.dataChan:
		return []byte(data), nil
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("function call timed out after %d seconds", 5)
	}
}
