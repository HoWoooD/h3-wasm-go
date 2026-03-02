package h3

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// Global WASM runtime shared across all tests
var (
	testRuntime     wazero.Runtime
	testModule      wazero.CompiledModule
	testModuleOnce  sync.Once
	testRuntimeInit error
)

// initTestRuntime initializes the WASM runtime once for all tests
func initTestRuntime(t *testing.T) (wazero.Runtime, wazero.CompiledModule) {
	t.Helper()

	testModuleOnce.Do(func() {
		ctx := context.Background()

		// Create runtime
		testRuntime = wazero.NewRuntime(ctx)

		// Instantiate WASI (required for C-compiled WASM)
		if _, err := wasi_snapshot_preview1.Instantiate(ctx, testRuntime); err != nil {
			testRuntimeInit = err
			return
		}

		// Read WASM binary
		wasmPath := "../testdata/h3.wasm"
		wasmBytes, err := os.ReadFile(wasmPath)
		if err != nil {
			testRuntimeInit = err
			return
		}

		// Compile WASM module (compiled once, reused for all test instances)
		testModule, err = testRuntime.CompileModule(ctx, wasmBytes)
		if err != nil {
			testRuntimeInit = err
			return
		}
	})

	if testRuntimeInit != nil {
		t.Fatalf("Failed to initialize test runtime: %v", testRuntimeInit)
	}

	return testRuntime, testModule
}

// cleanupTestRuntime should be called in TestMain to cleanup resources
func cleanupTestRuntime() {
	if testRuntime != nil {
		ctx := context.Background()
		testRuntime.Close(ctx)
	}
}

// newTestModule creates a new module instance for a test
func newTestModule(t *testing.T) api.Module {
	t.Helper()

	runtime, compiledModule := initTestRuntime(t)
	ctx := context.Background()

	// Instantiate the module
	mod, err := runtime.InstantiateModule(ctx, compiledModule, wazero.NewModuleConfig())
	if err != nil {
		t.Fatalf("Failed to instantiate module: %v", err)
	}

	return mod
}

// TestMain handles setup and teardown for all tests
func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()

	// Cleanup
	cleanupTestRuntime()

	os.Exit(code)
}

// TestWASMModuleLoads verifies that the WASM module can be loaded
func TestWASMModuleLoads(t *testing.T) {
	mod := newTestModule(t)
	defer mod.Close(context.Background())

	// Verify key functions are exported
	requiredFunctions := []string{
		"latLngToCellWrapper",
		"cellToLatLngWrapper",
		"cellToBoundaryWrapper",
		"gridDiskWrapper",
		"gridDistanceWrapper",
		"maxGridDiskSizeWrapper",
		"degsToRadsWrapper",
		"radsToDegsWrapper",
		"allocate",
		"deallocate",
	}

	for _, fnName := range requiredFunctions {
		fn := mod.ExportedFunction(fnName)
		if fn == nil {
			t.Errorf("Expected function %s to be exported", fnName)
		}
	}
}
