package h3

import "fmt"

// H3Index is a unique identifier for an H3 cell (hexagon or pentagon).
// It's a 64-bit unsigned integer encoding the cell's location and resolution.
type H3Index uint64

// H3_NULL represents an invalid H3 index, analogous to NaN in floating point.
const H3_NULL H3Index = 0

// Resolution constants define the level of detail for H3 cells.
// Resolution 0 is the coarsest (largest hexagons), 15 is the finest (smallest hexagons).
const (
	MinResolution = 0
	MaxResolution = 15
)

// LatLng represents a geographic coordinate in decimal degrees.
// Internally, H3 uses radians, but this API uses degrees for convenience.
type LatLng struct {
	Lat float64 // Latitude in decimal degrees
	Lng float64 // Longitude in decimal degrees
}

// CellBoundary represents the vertices of an H3 cell boundary.
// Most cells are hexagons (6 vertices), but 12 per resolution are pentagons (5 vertices).
type CellBoundary struct {
	Vertices []LatLng
}

// H3Error represents error codes returned by H3 functions.
type H3Error uint32

// H3 error codes matching the C library error codes
const (
	E_SUCCESS         H3Error = 0  // Success (no error)
	E_FAILED          H3Error = 1  // The operation failed but a more specific error is not available
	E_DOMAIN          H3Error = 2  // Argument was outside of acceptable range
	E_LATLNG_DOMAIN   H3Error = 3  // Latitude or longitude arguments were outside of acceptable range
	E_RES_DOMAIN      H3Error = 4  // Resolution argument was outside of acceptable range
	E_CELL_INVALID    H3Error = 5  // H3Index cell argument was not valid
	E_DIR_EDGE_INVALID H3Error = 6 // H3Index directed edge argument was not valid
	E_UNDIR_EDGE_INVALID H3Error = 7 // H3Index undirected edge argument was not valid
	E_VERTEX_INVALID  H3Error = 8  // H3Index vertex argument was not valid
	E_PENTAGON        H3Error = 9  // Pentagon distortion was encountered
	E_DUPLICATE_INPUT H3Error = 10 // Duplicate input was encountered
	E_NOT_NEIGHBORS   H3Error = 11 // H3Index cell arguments were not neighbors
	E_RES_MISMATCH    H3Error = 12 // H3Index cell arguments had incompatible resolutions
	E_MEMORY_ALLOC    H3Error = 13 // Necessary memory allocation failed
	E_MEMORY_BOUNDS   H3Error = 14 // Bounds of provided memory were not large enough
	E_OPTION_INVALID  H3Error = 15 // Mode or flags argument was not valid
	E_INDEX_INVALID   H3Error = 16 // H3Index argument was not valid
	E_BASE_CELL_DOMAIN H3Error = 17 // Base cell number was outside of acceptable range
	E_DIGIT_DOMAIN    H3Error = 18 // Child digits invalid
	E_DELETED_DIGIT   H3Error = 19 // Deleted subsequence indicates invalid index
)

// Error messages for H3 errors
var h3ErrorMessages = map[H3Error]string{
	E_SUCCESS:         "Success",
	E_FAILED:          "Operation failed",
	E_DOMAIN:          "Argument outside acceptable range",
	E_LATLNG_DOMAIN:   "Latitude or longitude outside acceptable range",
	E_RES_DOMAIN:      "Resolution outside acceptable range",
	E_CELL_INVALID:    "Invalid H3 cell",
	E_DIR_EDGE_INVALID: "Invalid directed edge",
	E_UNDIR_EDGE_INVALID: "Invalid undirected edge",
	E_VERTEX_INVALID:  "Invalid vertex",
	E_PENTAGON:        "Pentagon distortion encountered",
	E_DUPLICATE_INPUT: "Duplicate input encountered",
	E_NOT_NEIGHBORS:   "Cells are not neighbors",
	E_RES_MISMATCH:    "Cell resolutions do not match",
	E_MEMORY_ALLOC:    "Memory allocation failed",
	E_MEMORY_BOUNDS:   "Memory bounds exceeded",
	E_OPTION_INVALID:  "Invalid mode or flags",
	E_INDEX_INVALID:   "Invalid H3 index",
	E_BASE_CELL_DOMAIN: "Base cell outside acceptable range",
	E_DIGIT_DOMAIN:    "Invalid child digits",
	E_DELETED_DIGIT:   "Deleted subsequence in index",
}

// Error implements the error interface for H3Error.
func (e H3Error) Error() string {
	if msg, ok := h3ErrorMessages[e]; ok {
		return fmt.Sprintf("H3 error: %s (code %d)", msg, e)
	}
	return fmt.Sprintf("H3 error: unknown error code %d", e)
}

// IsValid returns true if the H3Index is not null.
func (h H3Index) IsValid() bool {
	return h != H3_NULL
}

// String returns the hexadecimal string representation of the H3Index.
func (h H3Index) String() string {
	return fmt.Sprintf("%016x", uint64(h))
}
