package h3

import (
	"math"
	"testing"
)

func TestNew(t *testing.T) {
	h, err := New()
	if err != nil {
		t.Fatalf("Failed to create H3 instance: %v", err)
	}
	defer h.Close()

	if h == nil {
		t.Fatal("H3 instance is nil")
	}
}

func TestMemoryReadWrite(t *testing.T) {
	h, err := New()
	if err != nil {
		t.Fatalf("Failed to create H3 instance: %v", err)
	}
	defer h.Close()

	// Allocate memory for testing
	ptr, err := h.runtime.Allocate(64)
	if err != nil {
		t.Fatalf("Failed to allocate memory: %v", err)
	}
	defer h.runtime.Deallocate(ptr)

	t.Run("Float64", func(t *testing.T) {
		testValue := 3.141592653589793
		if err := h.writeFloat64(ptr, testValue); err != nil {
			t.Fatalf("Failed to write float64: %v", err)
		}

		readValue, err := h.readFloat64(ptr)
		if err != nil {
			t.Fatalf("Failed to read float64: %v", err)
		}

		if readValue != testValue {
			t.Errorf("Expected %f, got %f", testValue, readValue)
		}
	})

	t.Run("Uint64", func(t *testing.T) {
		testValue := uint64(0x123456789ABCDEF0)
		if err := h.writeUint64(ptr, testValue); err != nil {
			t.Fatalf("Failed to write uint64: %v", err)
		}

		readValue, err := h.readUint64(ptr)
		if err != nil {
			t.Fatalf("Failed to read uint64: %v", err)
		}

		if readValue != testValue {
			t.Errorf("Expected %x, got %x", testValue, readValue)
		}
	})

	t.Run("Int64", func(t *testing.T) {
		testValue := int64(-123456789)
		if err := h.writeInt64(ptr, testValue); err != nil {
			t.Fatalf("Failed to write int64: %v", err)
		}

		readValue, err := h.readInt64(ptr)
		if err != nil {
			t.Fatalf("Failed to read int64: %v", err)
		}

		if readValue != testValue {
			t.Errorf("Expected %d, got %d", testValue, readValue)
		}
	})

	t.Run("Int32", func(t *testing.T) {
		testValue := int32(-12345)
		if err := h.writeInt32(ptr, testValue); err != nil {
			t.Fatalf("Failed to write int32: %v", err)
		}

		readValue, err := h.readInt32(ptr)
		if err != nil {
			t.Fatalf("Failed to read int32: %v", err)
		}

		if readValue != testValue {
			t.Errorf("Expected %d, got %d", testValue, readValue)
		}
	})
}

func TestDegsToRads(t *testing.T) {
	h, err := New()
	if err != nil {
		t.Fatalf("Failed to create H3 instance: %v", err)
	}
	defer h.Close()

	testCases := []struct {
		degrees float64
		radians float64
	}{
		{0, 0},
		{90, math.Pi / 2},
		{180, math.Pi},
		{360, 2 * math.Pi},
		{-90, -math.Pi / 2},
	}

	for _, tc := range testCases {
		result, err := h.degsToRads(tc.degrees)
		if err != nil {
			t.Errorf("degsToRads(%f) failed: %v", tc.degrees, err)
			continue
		}

		// Allow small floating point error
		if math.Abs(result-tc.radians) > 1e-10 {
			t.Errorf("degsToRads(%f) = %f, expected %f", tc.degrees, result, tc.radians)
		}
	}
}

func TestRadsToDegs(t *testing.T) {
	h, err := New()
	if err != nil {
		t.Fatalf("Failed to create H3 instance: %v", err)
	}
	defer h.Close()

	testCases := []struct {
		radians float64
		degrees float64
	}{
		{0, 0},
		{math.Pi / 2, 90},
		{math.Pi, 180},
		{2 * math.Pi, 360},
		{-math.Pi / 2, -90},
	}

	for _, tc := range testCases {
		result, err := h.radsToDegs(tc.radians)
		if err != nil {
			t.Errorf("radsToDegs(%f) failed: %v", tc.radians, err)
			continue
		}

		// Allow small floating point error
		if math.Abs(result-tc.degrees) > 1e-10 {
			t.Errorf("radsToDegs(%f) = %f, expected %f", tc.radians, result, tc.degrees)
		}
	}
}

func TestRoundTripConversion(t *testing.T) {
	h, err := New()
	if err != nil {
		t.Fatalf("Failed to create H3 instance: %v", err)
	}
	defer h.Close()

	degrees := 45.678
	rads, err := h.degsToRads(degrees)
	if err != nil {
		t.Fatalf("degsToRads failed: %v", err)
	}

	backToDegs, err := h.radsToDegs(rads)
	if err != nil {
		t.Fatalf("radsToDegs failed: %v", err)
	}

	if math.Abs(backToDegs-degrees) > 1e-10 {
		t.Errorf("Round-trip conversion failed: %f -> %f -> %f", degrees, rads, backToDegs)
	}
}
