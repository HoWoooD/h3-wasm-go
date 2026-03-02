# wazero-h3

Go bindings for [Uber's H3](https://github.com/uber/h3) geospatial indexing library via WebAssembly using [wazero](https://wazero.io/).

H3 is a hexagonal hierarchical geospatial indexing system that allows you to index geographic locations into hexagonal cells at various resolutions (0-15), enabling efficient spatial queries and analysis.

## Features

- ✅ **Pure Go**: No CGo required - runs entirely in Go using WebAssembly
- ✅ **Cross-platform**: Works on any platform supported by Go (no C compiler needed)
- ✅ **Fast**: Uses wazero's compiler mode for near-native performance
- ✅ **Type-safe**: Idiomatic Go API with proper types and error handling
- ✅ **Core H3 Functions**: Implements the most commonly used H3 operations

## Installation

```bash
go get github.com/algysbuldakov/wazero-h3
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/algysbuldakov/wazero-h3/h3"
)

func main() {
    // Create H3 instance
    h, err := h3.New()
    if err != nil {
        log.Fatal(err)
    }
    defer h.Close()
    
    // Convert coordinates to H3 index
    lat, lng := 37.7749, -122.4194 // San Francisco
    index, err := h.LatLngToCell(lat, lng, 9)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("H3 Index: %s\n", index.String())
    // Output: H3 Index: 089283082803ffff
    
    // Get cell center coordinates
    centerLat, centerLng, err := h.CellToLatLng(index)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Center: (%.6f, %.6f)\n", centerLat, centerLng)
    
    // Get neighboring cells
    neighbors, err := h.GridDisk(index, 1)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Neighbors: %d cells\n", len(neighbors))
}
```

## API Documentation

### Core Types

- **`H3Index`** - 64-bit unsigned integer representing an H3 cell
- **`LatLng`** - Geographic coordinates in decimal degrees
- **`CellBoundary`** - Cell boundary with vertices
- **`H3Error`** - Error codes from H3 operations

### Available Functions

#### `New() (*H3, error)`
Creates a new H3 instance. Must call `Close()` when done.

#### `LatLngToCell(lat, lng float64, resolution int) (H3Index, error)`
Converts geographic coordinates to an H3 cell index at the specified resolution (0-15).

**Parameters:**
- `lat`: Latitude in decimal degrees (-90 to 90)
- `lng`: Longitude in decimal degrees (-180 to 180)
- `resolution`: H3 resolution level (0 = coarsest, 15 = finest)

**Returns:** H3 cell index

#### `CellToLatLng(cell H3Index) (lat, lng float64, error)`
Converts an H3 cell index to its center coordinates.

**Returns:** Latitude and longitude in decimal degrees

#### `CellToBoundary(cell H3Index) (*CellBoundary, error)`
Returns the boundary vertices of an H3 cell.

**Returns:** CellBoundary with 5 (pentagon) or 6 (hexagon) vertices in counter-clockwise order

#### `GridDisk(origin H3Index, k int) ([]H3Index, error)`
Returns all H3 cells within k "rings" of the origin cell.

**Parameters:**
- `origin`: The center H3 cell
- `k`: Number of rings (0 = just origin, 1 = origin + neighbors, etc.)

**Returns:** Slice of H3 cell indices

#### `GridDistance(origin, destination H3Index) (int, error)`
Calculates the grid distance between two H3 cells.

**Returns:** Minimum number of grid steps between cells

### Error Handling

All functions return H3-specific errors that implement the `error` interface:

```go
index, err := h.LatLngToCell(lat, lng, resolution)
if err != nil {
    switch err {
    case h3.E_LATLNG_DOMAIN:
        // Invalid coordinates
    case h3.E_RES_DOMAIN:
        // Invalid resolution
    default:
        // Other H3 error
    }
}
```

## Examples

See [examples/basic/main.go](examples/basic/main.go) for a comprehensive demonstration including:
- Coordinate to H3 index conversion
- H3 index to coordinates conversion
- Cell boundary vertices
- Grid disk (neighbors)
- Grid distance calculation
- Multi-resolution indexing

Run the example:
```bash
go run examples/basic/main.go
```

## Building

The H3 WebAssembly module is pre-compiled and embedded in the library. If you need to rebuild it:

### Prerequisites

- Go 1.24+
- CMake
- WASI SDK (automatically installed by build script)

### Build Steps

```bash
# Clone repository with H3 source
git clone https://github.com/uber/h3.git h3-source

# Run build script
./build-wasm.sh
```

This will:
1. Configure H3 with CMake for wasm32-wasi target
2. Build the H3 static library
3. Compile wrapper functions
4. Link everything into `testdata/h3.wasm`

See [BUILD.md](BUILD.md) for detailed build instructions.

## Testing

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./h3

# Run specific test
go test -v ./h3 -run TestLatLngToCell
```

## Architecture

This library uses WebAssembly as a bridge between Go and the H3 C library:

```
┌─────────────┐
│   Go Code   │
│  (Your App) │
└─────┬───────┘
      │
      │ Go API (h3 package)
      ▼
┌─────────────────────┐
│  WASM Runtime       │
│  (wazero)           │
│  ┌───────────────┐  │
│  │   h3.wasm     │  │
│  │ (H3 C Library)│  │
│  └───────────────┘  │
└─────────────────────┘
```

- **Go API Layer** (`h3` package): Type-safe, idiomatic Go interface
- **WASM Runtime** (`internal/wasm` package): Manages wazero runtime and module lifecycle
- **H3 WebAssembly Module** (`testdata/h3.wasm`): H3 C library compiled to WASM

### Why WebAssembly?

- **No CGo**: Easier to build, better cross-compilation
- **Portable**: Works on any platform Go supports
- **Safe**: WASM sandbox provides memory safety
- **Fast**: wazero's compiler provides near-native performance

## Performance

WebAssembly performance with wazero's compiler mode is competitive for most use cases:

- **LatLngToCell**: ~6.4 µs per operation (~156K ops/sec)
- **CellToLatLng**: ~8.3 µs per operation (~120K ops/sec)
- **GridDisk (k=1)**: ~6.9 µs per operation (~145K ops/sec)
- **GridDistance**: ~3.5 µs per operation (~282K ops/sec)

While WASM is slower than native geospatial libraries (Geohash: 55ns, S2: 222ns), it provides:
- Zero CGo dependencies (easier deployment)
- Cross-platform compatibility
- Full H3 feature set (not just basic indexing)

See [BENCHMARKS.md](BENCHMARKS.md) for detailed performance analysis and optimization recommendations.

**Best practices for production:**
1. Reuse `H3` instance across your application
2. Batch operations when possible
3. Cache frequently used indices
4. Pre-compute H3 indices for static data

## Limitations

This library currently implements the core H3 functions. Additional H3 functions can be added by:
1. Adding wrapper functions in `h3_wrapper.c`
2. Exporting them in the build script
3. Implementing Go wrappers in `h3/h3.go`

Not yet implemented:
- Edge functions
- Vertex functions
- Hierarchy functions
- Directed edge operations

Pull requests welcome!

## License

This project is licensed under the Apache License 2.0 - see [LICENSE](LICENSE) for details.

The H3 library itself is also Apache 2.0 licensed. See [h3-source/LICENSE](h3-source/LICENSE).

## Credits

- [Uber H3](https://github.com/uber/h3) - The amazing geospatial indexing library
- [wazero](https://wazero.io/) - The zero-dependency WebAssembly runtime for Go

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Links

- [H3 Documentation](https://h3geo.org/)
- [H3 Specification](https://h3geo.org/docs/)
- [wazero Documentation](https://wazero.io/)
- [WebAssembly](https://webassembly.org/)
