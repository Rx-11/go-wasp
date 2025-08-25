package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Rx-11/go-wasp/constants"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

func ExecuteWASM(wasmBytes []byte, input map[string]any) (map[string]any, error) {
	inJSON, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshal input: %w", err)
	}
	stdin := bytes.NewReader(inJSON)
	var stdout, stderr bytes.Buffer

	ctx, cancel := context.WithTimeout(context.Background(), constants.Timeout)
	defer cancel()

	rt := wazero.NewRuntime(ctx)
	defer rt.Close(ctx)

	if _, err := wasi_snapshot_preview1.Instantiate(ctx, rt); err != nil {
		return nil, fmt.Errorf("instantiate WASI: %w", err)
	}

	compiled, err := rt.CompileModule(ctx, wasmBytes)
	if err != nil {
		return nil, fmt.Errorf("compile wasm: %w", err)
	}

	cfg := wazero.NewModuleConfig().
		WithStdin(stdin).
		WithStdout(&stdout).
		WithStderr(&stderr)

	mod, err := rt.InstantiateModule(ctx, compiled, cfg)
	if err != nil {
		return nil, fmt.Errorf("instantiate module: %w; stderr=%s", err, stderr.String())
	}
	defer mod.Close(ctx)

	start := mod.ExportedFunction("_start")
	if start == nil {
		return nil, errors.New("module has no _start export (WASI entrypoint required)")
	}

	if _, err := start.Call(ctx); err != nil {
		return nil, fmt.Errorf("invoke _start: %w; stderr=%s", err, stderr.String())
	}

	var out map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		return nil, fmt.Errorf("invalid JSON on stdout: %w; raw=%q", err, stdout.String())
	}

	return out, nil
}
