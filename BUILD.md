# Build Instructions

## Prerequisites

### WASM Toolchain

This project uses WASI SDK for compiling H3 C library to WebAssembly.

**WASI SDK is already installed in the project:**
- Location: `./wasi-sdk/`
- Version: wasi-sdk-24.0
- Clang version: 18.1.2

**To use it:**
```bash
export WASI_SDK_PATH=$(pwd)/wasi-sdk
```

### H3 Source Code

H3 source code is cloned in `./h3-source/` directory.

## Building H3 WebAssembly Module

The project includes a build script that compiles H3 C library to WebAssembly.

### Quick Build

```bash
./build-wasm.sh
```

This will:
1. Configure H3 with CMake for wasm32-wasi target
2. Build static library (libh3.a)
3. Compile wrapper functions
4. Link everything into `testdata/h3.wasm`

### Manual Build

If you need to rebuild manually:

```bash
export WASI_SDK_PATH=$(pwd)/wasi-sdk

# 1. Build H3 static library
mkdir -p build-wasm && cd build-wasm
cmake ../h3-source \
    -DCMAKE_TOOLCHAIN_FILE="$WASI_SDK_PATH/share/cmake/wasi-sdk.cmake" \
    -DCMAKE_BUILD_TYPE=Release \
    -DBUILD_SHARED_LIBS=OFF \
    -DBUILD_TESTING=OFF
cmake --build . --config Release
cd ..

# 2. Compile wrapper to WASM
$WASI_SDK_PATH/bin/clang \
  -o testdata/h3.wasm \
  h3_wrapper.c \
  -I build-wasm/src/h3lib/include \
  -L build-wasm/lib \
  -lh3 \
  -O3 \
  -Wl,--export=latLngToCellWrapper \
  -Wl,--export=cellToLatLngWrapper \
  -Wl,--export=cellToBoundaryWrapper \
  -Wl,--export=gridDiskWrapper \
  -Wl,--export=gridDistanceWrapper \
  -Wl,--export=maxGridDiskSizeWrapper \
  -Wl,--export=degsToRadsWrapper \
  -Wl,--export=radsToDegsWrapper \
  -Wl,--export=allocate \
  -Wl,--export=deallocate \
  -Wl,--no-entry
```

### Exported Functions

The WASM module exports the following H3 functions:
- `latLngToCellWrapper` - Convert lat/lng to H3 cell index
- `cellToLatLngWrapper` - Convert H3 cell to lat/lng
- `cellToBoundaryWrapper` - Get cell boundary vertices
- `gridDiskWrapper` - Get cells within k rings
- `gridDistanceWrapper` - Calculate distance between cells
- `maxGridDiskSizeWrapper` - Get max size for gridDisk result
- `degsToRadsWrapper` / `radsToDegsWrapper` - Unit conversion
- `allocate` / `deallocate` - Memory management

## Testing

```bash
go test ./...
```
