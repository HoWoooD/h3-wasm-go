#!/bin/bash
set -e

# Path to WASI SDK
export WASI_SDK_PATH="$(pwd)/wasi-sdk"
export CC="$WASI_SDK_PATH/bin/clang"
export CXX="$WASI_SDK_PATH/bin/clang++"
export AR="$WASI_SDK_PATH/bin/llvm-ar"
export RANLIB="$WASI_SDK_PATH/bin/llvm-ranlib"

echo "Building H3 for WebAssembly..."
echo "WASI SDK: $WASI_SDK_PATH"

# Create build directory
mkdir -p build-wasm
cd build-wasm

# Configure with CMake for WASM
cmake ../h3-source \
    -DCMAKE_TOOLCHAIN_FILE="$WASI_SDK_PATH/share/cmake/wasi-sdk.cmake" \
    -DCMAKE_BUILD_TYPE=Release \
    -DBUILD_SHARED_LIBS=OFF \
    -DBUILD_TESTING=OFF \
    -DBUILD_BENCHMARKS=OFF \
    -DBUILD_FILTERS=OFF \
    -DBUILD_GENERATORS=OFF \
    -DENABLE_COVERAGE=OFF \
    -DENABLE_FORMAT=OFF \
    -DENABLE_LINTING=OFF

# Build
cmake --build . --config Release

echo ""
echo "Build complete!"
echo "Looking for static library..."
find . -name "*.a" -type f

# Create a simple wrapper C file that exports the functions we need
cd ..
cat > build-wasm/h3_wrapper.c << 'EOF'
#include "h3-source/src/h3lib/include/h3api.h"

// Export functions for WASM
__attribute__((used)) H3Error latLngToCellExport(const LatLng *g, int res, H3Index *out) {
    return latLngToCell(g, res, out);
}

__attribute__((used)) H3Error cellToLatLngExport(H3Index cell, LatLng *g) {
    return cellToLatLng(cell, g);
}

__attribute__((used)) H3Error cellToBoundaryExport(H3Index cell, CellBoundary *bndry) {
    return cellToBoundary(cell, bndry);
}

__attribute__((used)) H3Error gridDiskExport(H3Index origin, int k, H3Index *out) {
    return gridDisk(origin, k, out);
}

__attribute__((used)) H3Error gridDistanceExport(H3Index origin, H3Index h3, int64_t *distance) {
    return gridDistance(origin, h3, distance);
}

__attribute__((used)) H3Error maxGridDiskSizeExport(int k, int64_t *out) {
    return maxGridDiskSize(k, out);
}

__attribute__((used)) double degsToRadsExport(double degrees) {
    return degsToRads(degrees);
}

__attribute__((used)) double radsToDegsExport(double radians) {
    return radsToDegs(radians);
}
EOF

echo ""
echo "Compiling wrapper and linking to WASM..."

# Compile wrapper and link with H3 library
$CC \
    -o testdata/h3.wasm \
    wasm-src/h3_wrapper.c \
    -I build-wasm/src/h3lib/include \
    -L build-wasm/lib \
    -lh3 \
    -O3 \
    -nostartfiles \
    -Wl,--no-entry \
    -Wl,--export=latLngToCellWrapper \
    -Wl,--export=cellToLatLngWrapper \
    -Wl,--export=cellToBoundaryWrapper \
    -Wl,--export=gridDiskWrapper \
    -Wl,--export=gridDistanceWrapper \
    -Wl,--export=maxGridDiskSizeWrapper \
    -Wl,--export=degsToRadsWrapper \
    -Wl,--export=radsToDegsWrapper \
    -Wl,--export=allocate \
    -Wl,--export=deallocate

echo ""
echo "✅ Success! WebAssembly module created: testdata/h3.wasm"
ls -lh testdata/h3.wasm
