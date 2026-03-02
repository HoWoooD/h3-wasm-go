package h3

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/HoWoooD/h3-wasm-go/internal/wasm"
)

// H3 is the main interface for working with H3 geospatial indexing.
// It wraps the WASM runtime and provides high-level Go functions.
type H3 struct {
	runtime *wasm.Runtime
	ctx     context.Context
}

// New creates a new H3 instance with an initialized WASM runtime.
func New() (*H3, error) {
	ctx := context.Background()
	rt, err := wasm.NewRuntime(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create WASM runtime: %w", err)
	}

	return &H3{
		runtime: rt,
		ctx:     ctx,
	}, nil
}

// Close releases all resources held by the H3 instance.
func (h *H3) Close() error {
	if h.runtime != nil {
		return h.runtime.Close()
	}
	return nil
}

// Helper functions for working with WASM memory

// writeFloat64 writes a float64 value to WASM memory at the given offset
func (h *H3) writeFloat64(ptr uint32, value float64) error {
	mem := h.runtime.GetMemory()
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, math.Float64bits(value))
	if !mem.Write(ptr, buf) {
		return fmt.Errorf("failed to write float64 to memory at offset %d", ptr)
	}
	return nil
}

// readFloat64 reads a float64 value from WASM memory at the given offset
func (h *H3) readFloat64(ptr uint32) (float64, error) {
	mem := h.runtime.GetMemory()
	buf, ok := mem.Read(ptr, 8)
	if !ok {
		return 0, fmt.Errorf("failed to read float64 from memory at offset %d", ptr)
	}
	bits := binary.LittleEndian.Uint64(buf)
	return math.Float64frombits(bits), nil
}

// writeUint64 writes a uint64 value to WASM memory at the given offset
func (h *H3) writeUint64(ptr uint32, value uint64) error {
	mem := h.runtime.GetMemory()
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, value)
	if !mem.Write(ptr, buf) {
		return fmt.Errorf("failed to write uint64 to memory at offset %d", ptr)
	}
	return nil
}

// readUint64 reads a uint64 value from WASM memory at the given offset
func (h *H3) readUint64(ptr uint32) (uint64, error) {
	mem := h.runtime.GetMemory()
	buf, ok := mem.Read(ptr, 8)
	if !ok {
		return 0, fmt.Errorf("failed to read uint64 from memory at offset %d", ptr)
	}
	return binary.LittleEndian.Uint64(buf), nil
}

// writeInt64 writes an int64 value to WASM memory at the given offset
func (h *H3) writeInt64(ptr uint32, value int64) error {
	mem := h.runtime.GetMemory()
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(value))
	if !mem.Write(ptr, buf) {
		return fmt.Errorf("failed to write int64 to memory at offset %d", ptr)
	}
	return nil
}

// readInt64 reads an int64 value from WASM memory at the given offset
func (h *H3) readInt64(ptr uint32) (int64, error) {
	mem := h.runtime.GetMemory()
	buf, ok := mem.Read(ptr, 8)
	if !ok {
		return 0, fmt.Errorf("failed to read int64 from memory at offset %d", ptr)
	}
	return int64(binary.LittleEndian.Uint64(buf)), nil
}

// writeInt32 writes an int32 value to WASM memory at the given offset
func (h *H3) writeInt32(ptr uint32, value int32) error {
	mem := h.runtime.GetMemory()
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(value))
	if !mem.Write(ptr, buf) {
		return fmt.Errorf("failed to write int32 to memory at offset %d", ptr)
	}
	return nil
}

// readInt32 reads an int32 value from WASM memory at the given offset
func (h *H3) readInt32(ptr uint32) (int32, error) {
	mem := h.runtime.GetMemory()
	buf, ok := mem.Read(ptr, 4)
	if !ok {
		return 0, fmt.Errorf("failed to read int32 from memory at offset %d", ptr)
	}
	return int32(binary.LittleEndian.Uint32(buf)), nil
}

// degsToRads converts degrees to radians using H3's conversion function
func (h *H3) degsToRads(degrees float64) (float64, error) {
	fn := h.runtime.GetFunction("degsToRadsWrapper")
	if fn == nil {
		return 0, fmt.Errorf("degsToRadsWrapper function not found")
	}

	results, err := fn.Call(h.ctx, math.Float64bits(degrees))
	if err != nil {
		return 0, fmt.Errorf("degsToRads failed: %w", err)
	}

	if len(results) == 0 {
		return 0, fmt.Errorf("degsToRads returned no results")
	}

	return math.Float64frombits(results[0]), nil
}

// radsToDegs converts radians to degrees using H3's conversion function
func (h *H3) radsToDegs(radians float64) (float64, error) {
	fn := h.runtime.GetFunction("radsToDegsWrapper")
	if fn == nil {
		return 0, fmt.Errorf("radsToDegsWrapper function not found")
	}

	results, err := fn.Call(h.ctx, math.Float64bits(radians))
	if err != nil {
		return 0, fmt.Errorf("radsToDegs failed: %w", err)
	}

	if len(results) == 0 {
		return 0, fmt.Errorf("radsToDegs returned no results")
	}

	return math.Float64frombits(results[0]), nil
}

// LatLngToCell converts geographic coordinates to an H3 cell index at the specified resolution.
// 
// Parameters:
//   - lat: Latitude in decimal degrees (range: -90 to 90)
//   - lng: Longitude in decimal degrees (range: -180 to 180)
//   - resolution: H3 resolution level (range: 0 to 15)
//
// Returns the H3 cell index or an error if the conversion fails.
func (h *H3) LatLngToCell(lat, lng float64, resolution int) (H3Index, error) {
	// Validate input
	if resolution < MinResolution || resolution > MaxResolution {
		return H3_NULL, fmt.Errorf("resolution %d out of range [%d, %d]", resolution, MinResolution, MaxResolution)
	}

	if lat < -90 || lat > 90 {
		return H3_NULL, fmt.Errorf("latitude %f out of range [-90, 90]", lat)
	}

	if lng < -180 || lng > 180 {
		return H3_NULL, fmt.Errorf("longitude %f out of range [-180, 180]", lng)
	}

	// Convert degrees to radians
	latRad, err := h.degsToRads(lat)
	if err != nil {
		return H3_NULL, fmt.Errorf("failed to convert latitude to radians: %w", err)
	}

	lngRad, err := h.degsToRads(lng)
	if err != nil {
		return H3_NULL, fmt.Errorf("failed to convert longitude to radians: %w", err)
	}

	// Allocate memory for output H3Index
	outPtr, err := h.runtime.Allocate(8)
	if err != nil {
		return H3_NULL, fmt.Errorf("failed to allocate memory for output: %w", err)
	}
	defer h.runtime.Deallocate(outPtr)

	// Call the WASM function
	fn := h.runtime.GetFunction("latLngToCellWrapper")
	if fn == nil {
		return H3_NULL, fmt.Errorf("latLngToCellWrapper function not found")
	}

	results, err := fn.Call(h.ctx,
		math.Float64bits(latRad),
		math.Float64bits(lngRad),
		uint64(resolution),
		uint64(outPtr),
	)
	if err != nil {
		return H3_NULL, fmt.Errorf("latLngToCell call failed: %w", err)
	}

	// Check error code
	if len(results) == 0 {
		return H3_NULL, fmt.Errorf("latLngToCell returned no results")
	}

	errCode := H3Error(results[0])
	if errCode != E_SUCCESS {
		return H3_NULL, errCode
	}

	// Read the H3Index from memory
	index, err := h.readUint64(outPtr)
	if err != nil {
		return H3_NULL, fmt.Errorf("failed to read H3Index from memory: %w", err)
	}

	return H3Index(index), nil
}

// CellToLatLng converts an H3 cell index to its center coordinates.
//
// Parameters:
//   - cell: The H3 cell index to convert
//
// Returns the latitude and longitude of the cell center in decimal degrees, or an error.
func (h *H3) CellToLatLng(cell H3Index) (lat, lng float64, err error) {
	// Validate input
	if cell == H3_NULL {
		return 0, 0, fmt.Errorf("invalid H3 index: H3_NULL")
	}

	// Allocate memory for output lat/lng (2 doubles = 16 bytes)
	latPtr, err := h.runtime.Allocate(8)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to allocate memory for latitude: %w", err)
	}
	defer h.runtime.Deallocate(latPtr)

	lngPtr, err := h.runtime.Allocate(8)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to allocate memory for longitude: %w", err)
	}
	defer h.runtime.Deallocate(lngPtr)

	// Call the WASM function
	fn := h.runtime.GetFunction("cellToLatLngWrapper")
	if fn == nil {
		return 0, 0, fmt.Errorf("cellToLatLngWrapper function not found")
	}

	results, callErr := fn.Call(h.ctx,
		uint64(cell),
		uint64(latPtr),
		uint64(lngPtr),
	)
	if callErr != nil {
		return 0, 0, fmt.Errorf("cellToLatLng call failed: %w", callErr)
	}

	// Check error code
	if len(results) == 0 {
		return 0, 0, fmt.Errorf("cellToLatLng returned no results")
	}

	errCode := H3Error(results[0])
	if errCode != E_SUCCESS {
		return 0, 0, errCode
	}

	// Read lat/lng from memory (in radians)
	latRad, err := h.readFloat64(latPtr)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read latitude from memory: %w", err)
	}

	lngRad, err := h.readFloat64(lngPtr)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read longitude from memory: %w", err)
	}

	// Convert radians to degrees
	latDeg, err := h.radsToDegs(latRad)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to convert latitude to degrees: %w", err)
	}

	lngDeg, err := h.radsToDegs(lngRad)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to convert longitude to degrees: %w", err)
	}

	return latDeg, lngDeg, nil
}

// CellToBoundary returns the boundary vertices of an H3 cell.
// Most cells are hexagons (6 vertices), but some are pentagons (5 vertices).
//
// Parameters:
//   - cell: The H3 cell index
//
// Returns the cell boundary with vertices in counter-clockwise order.
func (h *H3) CellToBoundary(cell H3Index) (*CellBoundary, error) {
	// Validate input
	if cell == H3_NULL {
		return nil, fmt.Errorf("invalid H3 index: H3_NULL")
	}

	// CellBoundary struct in C:
	// struct CellBoundary {
	//     int numVerts;                        // 4 bytes at offset 0
	//     // 4 bytes padding for alignment
	//     LatLng verts[MAX_CELL_BNDRY_VERTS];  // 10 * 16 bytes = 160 bytes at offset 8
	// }
	// Total: 168 bytes
	const maxVerts = 10
	const boundarySize = 168 // Actual size with padding

	// Allocate memory for CellBoundary structure
	boundaryPtr, err := h.runtime.Allocate(boundarySize)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate memory for boundary: %w", err)
	}
	defer h.runtime.Deallocate(boundaryPtr)

	// Call the WASM function
	fn := h.runtime.GetFunction("cellToBoundaryWrapper")
	if fn == nil {
		return nil, fmt.Errorf("cellToBoundaryWrapper function not found")
	}

	results, callErr := fn.Call(h.ctx,
		uint64(cell),
		uint64(boundaryPtr),
	)
	if callErr != nil {
		return nil, fmt.Errorf("cellToBoundary call failed: %w", callErr)
	}

	// Check error code
	if len(results) == 0 {
		return nil, fmt.Errorf("cellToBoundary returned no results")
	}

	errCode := H3Error(results[0])
	if errCode != E_SUCCESS {
		return nil, errCode
	}

	// Read numVerts
	numVerts, err := h.readInt32(boundaryPtr)
	if err != nil {
		return nil, fmt.Errorf("failed to read numVerts: %w", err)
	}

	if numVerts < 0 || numVerts > maxVerts {
		return nil, fmt.Errorf("invalid numVerts: %d", numVerts)
	}

	// Read vertices (each vertex is 2 float64s = 16 bytes)
	// Vertices start at offset 8 (after int + padding)
	vertices := make([]LatLng, numVerts)
	for i := 0; i < int(numVerts); i++ {
		offset := boundaryPtr + 8 + uint32(i*16) // 8 bytes offset (int + padding) + i * sizeof(LatLng)

		latRad, err := h.readFloat64(offset)
		if err != nil {
			return nil, fmt.Errorf("failed to read vertex %d latitude: %w", i, err)
		}

		lngRad, err := h.readFloat64(offset + 8)
		if err != nil {
			return nil, fmt.Errorf("failed to read vertex %d longitude: %w", i, err)
		}

		// Convert radians to degrees
		latDeg, err := h.radsToDegs(latRad)
		if err != nil {
			return nil, fmt.Errorf("failed to convert vertex %d latitude: %w", i, err)
		}

		lngDeg, err := h.radsToDegs(lngRad)
		if err != nil {
			return nil, fmt.Errorf("failed to convert vertex %d longitude: %w", i, err)
		}

		vertices[i] = LatLng{Lat: latDeg, Lng: lngDeg}
	}

	return &CellBoundary{Vertices: vertices}, nil
}

// GridDisk returns all H3 cells within k "rings" of the origin cell.
// Ring 0 is the origin cell itself, ring 1 includes all neighbors, etc.
//
// Parameters:
//   - origin: The center H3 cell
//   - k: The number of rings (0 = just origin, 1 = origin + neighbors, etc.)
//
// Returns a slice of H3 cell indices.
func (h *H3) GridDisk(origin H3Index, k int) ([]H3Index, error) {
	// Validate input
	if origin == H3_NULL {
		return nil, fmt.Errorf("invalid origin H3 index: H3_NULL")
	}

	if k < 0 {
		return nil, fmt.Errorf("k must be non-negative, got %d", k)
	}

	// Get maximum size for the result array
	maxSizePtr, err := h.runtime.Allocate(8)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate memory for maxSize: %w", err)
	}
	defer h.runtime.Deallocate(maxSizePtr)

	maxSizeFn := h.runtime.GetFunction("maxGridDiskSizeWrapper")
	if maxSizeFn == nil {
		return nil, fmt.Errorf("maxGridDiskSizeWrapper function not found")
	}

	results, err := maxSizeFn.Call(h.ctx, uint64(k), uint64(maxSizePtr))
	if err != nil {
		return nil, fmt.Errorf("maxGridDiskSize call failed: %w", err)
	}

	if len(results) == 0 || H3Error(results[0]) != E_SUCCESS {
		return nil, fmt.Errorf("maxGridDiskSize failed")
	}

	maxSize, err := h.readInt64(maxSizePtr)
	if err != nil {
		return nil, fmt.Errorf("failed to read maxSize: %w", err)
	}

	// Allocate memory for output array (maxSize * 8 bytes)
	outPtr, err := h.runtime.Allocate(uint32(maxSize * 8))
	if err != nil {
		return nil, fmt.Errorf("failed to allocate memory for output: %w", err)
	}
	defer h.runtime.Deallocate(outPtr)

	// Call gridDisk
	fn := h.runtime.GetFunction("gridDiskWrapper")
	if fn == nil {
		return nil, fmt.Errorf("gridDiskWrapper function not found")
	}

	results, callErr := fn.Call(h.ctx,
		uint64(origin),
		uint64(k),
		uint64(outPtr),
	)
	if callErr != nil {
		return nil, fmt.Errorf("gridDisk call failed: %w", callErr)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("gridDisk returned no results")
	}

	errCode := H3Error(results[0])
	if errCode != E_SUCCESS {
		return nil, errCode
	}

	// Read the H3 indices from memory
	cells := make([]H3Index, 0, maxSize)
	for i := int64(0); i < maxSize; i++ {
		offset := outPtr + uint32(i*8)
		index, err := h.readUint64(offset)
		if err != nil {
			return nil, fmt.Errorf("failed to read index %d: %w", i, err)
		}

		// Skip null indices (holes in the output)
		if index != 0 {
			cells = append(cells, H3Index(index))
		}
	}

	return cells, nil
}

// GridDistance calculates the grid distance between two H3 cells.
// Returns the minimum number of grid steps to get from origin to destination.
//
// Parameters:
//   - origin: The starting H3 cell
//   - destination: The ending H3 cell
//
// Returns the grid distance or an error if cells are not neighbors at compatible resolutions.
func (h *H3) GridDistance(origin, destination H3Index) (int, error) {
	// Validate input
	if origin == H3_NULL || destination == H3_NULL {
		return 0, fmt.Errorf("invalid H3 index")
	}

	// Allocate memory for output distance
	distPtr, err := h.runtime.Allocate(8)
	if err != nil {
		return 0, fmt.Errorf("failed to allocate memory for distance: %w", err)
	}
	defer h.runtime.Deallocate(distPtr)

	// Call the WASM function
	fn := h.runtime.GetFunction("gridDistanceWrapper")
	if fn == nil {
		return 0, fmt.Errorf("gridDistanceWrapper function not found")
	}

	results, callErr := fn.Call(h.ctx,
		uint64(origin),
		uint64(destination),
		uint64(distPtr),
	)
	if callErr != nil {
		return 0, fmt.Errorf("gridDistance call failed: %w", callErr)
	}

	if len(results) == 0 {
		return 0, fmt.Errorf("gridDistance returned no results")
	}

	errCode := H3Error(results[0])
	if errCode != E_SUCCESS {
		return 0, errCode
	}

	// Read the distance
	distance, err := h.readInt64(distPtr)
	if err != nil {
		return 0, fmt.Errorf("failed to read distance: %w", err)
	}

	return int(distance), nil
}
