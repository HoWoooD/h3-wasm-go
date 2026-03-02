package wasm

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// Runtime manages the H3 WASM module lifecycle.
// It handles loading, instantiation, and cleanup of the WASM runtime.
type Runtime struct {
	ctx            context.Context
	runtime        wazero.Runtime
	compiledModule wazero.CompiledModule
	module         api.Module
	mu             sync.Mutex
}

// getWasmBinary loads the H3 WASM binary from the embedded location or file system
func getWasmBinary() ([]byte, error) {
	// Try to find h3.wasm in common locations
	locations := []string{
		"testdata/h3.wasm",
		"../testdata/h3.wasm",
		"../../testdata/h3.wasm",
		"internal/wasm/h3.wasm",
	}

	for _, loc := range locations {
		if data, err := os.ReadFile(loc); err == nil {
			return data, nil
		}
	}

	// Try relative to working directory
	wd, _ := os.Getwd()
	wasmPath := filepath.Join(wd, "testdata", "h3.wasm")
	if data, err := os.ReadFile(wasmPath); err == nil {
		return data, nil
	}

	return nil, fmt.Errorf("h3.wasm not found in any known location")
}

// NewRuntime creates and initializes a new H3 WASM runtime.
// The runtime is ready to use immediately after creation.
func NewRuntime(ctx context.Context) (*Runtime, error) {
	r := &Runtime{
		ctx: ctx,
	}

	// Load WASM binary
	h3WasmBinary, err := getWasmBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to load WASM binary: %w", err)
	}

	// Create wazero runtime with compiler (faster execution)
	r.runtime = wazero.NewRuntime(ctx)

	// Instantiate WASI (required for C-compiled WASM modules)
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, r.runtime); err != nil {
		r.runtime.Close(ctx)
		return nil, fmt.Errorf("failed to instantiate WASI: %w", err)
	}

	// Compile the WASM module (compilation is expensive, done once)
	compiledModule, err := r.runtime.CompileModule(ctx, h3WasmBinary)
	if err != nil {
		r.runtime.Close(ctx)
		return nil, fmt.Errorf("failed to compile WASM module: %w", err)
	}
	r.compiledModule = compiledModule

	// Instantiate the module (creates memory and function instances)
	module, err := r.runtime.InstantiateModule(ctx, compiledModule, wazero.NewModuleConfig())
	if err != nil {
		r.runtime.Close(ctx)
		return nil, fmt.Errorf("failed to instantiate WASM module: %w", err)
	}
	r.module = module

	return r, nil
}

// GetFunction retrieves an exported function from the WASM module by name.
// Returns nil if the function doesn't exist.
func (r *Runtime) GetFunction(name string) api.Function {
	return r.module.ExportedFunction(name)
}

// GetMemory returns the WASM module's linear memory for reading/writing data.
func (r *Runtime) GetMemory() api.Memory {
	return r.module.Memory()
}

// Allocate allocates memory in the WASM linear memory space.
// Returns the pointer (offset) to the allocated memory.
func (r *Runtime) Allocate(size uint32) (uint32, error) {
	allocFn := r.GetFunction("allocate")
	if allocFn == nil {
		return 0, fmt.Errorf("allocate function not found")
	}

	results, err := allocFn.Call(r.ctx, uint64(size))
	if err != nil {
		return 0, fmt.Errorf("allocate failed: %w", err)
	}

	if len(results) == 0 {
		return 0, fmt.Errorf("allocate returned no results")
	}

	ptr := uint32(results[0])
	if ptr == 0 {
		return 0, fmt.Errorf("allocate returned null pointer")
	}

	return ptr, nil
}

// Deallocate frees memory in the WASM linear memory space.
func (r *Runtime) Deallocate(ptr uint32) error {
	if ptr == 0 {
		return nil // No-op for null pointer
	}

	deallocFn := r.GetFunction("deallocate")
	if deallocFn == nil {
		return fmt.Errorf("deallocate function not found")
	}

	_, err := deallocFn.Call(r.ctx, uint64(ptr))
	if err != nil {
		return fmt.Errorf("deallocate failed: %w", err)
	}

	return nil
}

// Close cleans up the runtime and releases all resources.
// After calling Close, the runtime cannot be used anymore.
func (r *Runtime) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.runtime != nil {
		return r.runtime.Close(r.ctx)
	}
	return nil
}
