package wasm

import (
	"context"
	"testing"
)

func TestNewRuntime(t *testing.T) {
	ctx := context.Background()
	rt, err := NewRuntime(ctx)
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	defer rt.Close()

	// Verify runtime is not nil
	if rt == nil {
		t.Fatal("Runtime is nil")
	}

	// Verify module is loaded
	if rt.module == nil {
		t.Fatal("Module is nil")
	}
}

func TestGetFunction(t *testing.T) {
	ctx := context.Background()
	rt, err := NewRuntime(ctx)
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	defer rt.Close()

	// Test getting an exported function
	fn := rt.GetFunction("latLngToCellWrapper")
	if fn == nil {
		t.Error("Expected latLngToCellWrapper function to exist")
	}

	// Test getting a non-existent function
	fn = rt.GetFunction("nonExistentFunction")
	if fn != nil {
		t.Error("Expected nonExistentFunction to not exist")
	}
}

func TestMemoryAllocation(t *testing.T) {
	ctx := context.Background()
	rt, err := NewRuntime(ctx)
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	defer rt.Close()

	// Allocate memory
	ptr, err := rt.Allocate(64)
	if err != nil {
		t.Fatalf("Failed to allocate memory: %v", err)
	}

	if ptr == 0 {
		t.Fatal("Allocated pointer is null")
	}

	// Deallocate memory
	err = rt.Deallocate(ptr)
	if err != nil {
		t.Fatalf("Failed to deallocate memory: %v", err)
	}

	// Test deallocating null pointer (should not error)
	err = rt.Deallocate(0)
	if err != nil {
		t.Errorf("Deallocating null pointer should not error: %v", err)
	}
}

func TestGetMemory(t *testing.T) {
	ctx := context.Background()
	rt, err := NewRuntime(ctx)
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	defer rt.Close()

	mem := rt.GetMemory()
	if mem == nil {
		t.Fatal("Memory is nil")
	}

	// Memory should have non-zero size
	if mem.Size() == 0 {
		t.Error("Memory size is zero")
	}
}
