package h3

import (
	"math"
	"testing"
)

// TestLatLngToCellBasic tests basic conversion from coordinates to H3 index
func TestLatLngToCellBasic(t *testing.T) {
	h, err := New()
	if err != nil {
		t.Fatalf("Failed to create H3 instance: %v", err)
	}
	defer h.Close()

	// San Francisco coordinates
	lat, lng := 37.7749, -122.4194

	testCases := []struct {
		resolution int
		expectErr  bool
	}{
		{0, false},
		{5, false},
		{10, false},
		{15, false},
		{-1, true},  // Invalid: resolution too low
		{16, true},  // Invalid: resolution too high
	}

	for _, tc := range testCases {
		index, err := h.LatLngToCell(lat, lng, tc.resolution)
		
		if tc.expectErr {
			if err == nil {
				t.Errorf("Expected error for resolution %d, got none", tc.resolution)
			}
			continue
		}

		if err != nil {
			t.Errorf("LatLngToCell(res=%d) failed: %v", tc.resolution, err)
			continue
		}

		if index == H3_NULL {
			t.Errorf("LatLngToCell(res=%d) returned null index", tc.resolution)
		}
	}
}

// TestCellToLatLngBasic tests conversion from H3 index back to coordinates
func TestCellToLatLngBasic(t *testing.T) {
	h, err := New()
	if err != nil {
		t.Fatalf("Failed to create H3 instance: %v", err)
	}
	defer h.Close()

	// San Francisco
	origLat, origLng := 37.7749, -122.4194
	
	index, err := h.LatLngToCell(origLat, origLng, 9)
	if err != nil {
		t.Fatalf("LatLngToCell failed: %v", err)
	}

	lat, lng, err := h.CellToLatLng(index)
	if err != nil {
		t.Fatalf("CellToLatLng failed: %v", err)
	}

	// Check that we're close to the original coordinates
	// (won't be exact due to hexagon center)
	const tolerance = 0.01 // About 1km at this latitude
	if math.Abs(lat-origLat) > tolerance {
		t.Errorf("Latitude mismatch: got %f, expected ~%f", lat, origLat)
	}
	if math.Abs(lng-origLng) > tolerance {
		t.Errorf("Longitude mismatch: got %f, expected ~%f", lng, origLng)
	}
}

// TestCellToBoundaryBasic tests getting cell boundary vertices
func TestCellToBoundaryBasic(t *testing.T) {
	h, err := New()
	if err != nil {
		t.Fatalf("Failed to create H3 instance: %v", err)
	}
	defer h.Close()

	index, err := h.LatLngToCell(37.7749, -122.4194, 9)
	if err != nil {
		t.Fatalf("LatLngToCell failed: %v", err)
	}

	boundary, err := h.CellToBoundary(index)
	if err != nil {
		t.Fatalf("CellToBoundary failed: %v", err)
	}

	// Most cells are hexagons (6 vertices), some are pentagons (5 vertices)
	if len(boundary.Vertices) != 6 && len(boundary.Vertices) != 5 {
		t.Errorf("Expected 5 or 6 vertices, got %d", len(boundary.Vertices))
	}

	// Verify all vertices have valid coordinates
	for i, v := range boundary.Vertices {
		if v.Lat < -90 || v.Lat > 90 {
			t.Errorf("Vertex %d has invalid latitude: %f", i, v.Lat)
		}
		if v.Lng < -180 || v.Lng > 180 {
			t.Errorf("Vertex %d has invalid longitude: %f", i, v.Lng)
		}
	}
}

// TestGridDiskBasic tests getting neighboring cells
func TestGridDiskBasic(t *testing.T) {
	h, err := New()
	if err != nil {
		t.Fatalf("Failed to create H3 instance: %v", err)
	}
	defer h.Close()

	origin, err := h.LatLngToCell(37.7749, -122.4194, 9)
	if err != nil {
		t.Fatalf("LatLngToCell failed: %v", err)
	}

	testCases := []struct {
		k            int
		expectedMin  int
		expectedMax  int
	}{
		{0, 1, 1},     // k=0: just the origin
		{1, 7, 7},     // k=1: origin + 6 neighbors
		{2, 19, 19},   // k=2: origin + rings 1 and 2
	}

	for _, tc := range testCases {
		cells, err := h.GridDisk(origin, tc.k)
		if err != nil {
			t.Errorf("GridDisk(k=%d) failed: %v", tc.k, err)
			continue
		}

		if len(cells) < tc.expectedMin || len(cells) > tc.expectedMax {
			t.Errorf("GridDisk(k=%d) returned %d cells, expected %d-%d",
				tc.k, len(cells), tc.expectedMin, tc.expectedMax)
		}

		// Verify origin is in the result for k >= 0
		found := false
		for _, cell := range cells {
			if cell == origin {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GridDisk(k=%d) did not include origin cell", tc.k)
		}
	}
}

// TestGridDistanceBasic tests distance calculation between cells
func TestGridDistanceBasic(t *testing.T) {
	h, err := New()
	if err != nil {
		t.Fatalf("Failed to create H3 instance: %v", err)
	}
	defer h.Close()

	origin, err := h.LatLngToCell(37.7749, -122.4194, 9)
	if err != nil {
		t.Fatalf("LatLngToCell failed: %v", err)
	}

	// Test distance to self
	dist, err := h.GridDistance(origin, origin)
	if err != nil {
		t.Fatalf("GridDistance to self failed: %v", err)
	}
	if dist != 0 {
		t.Errorf("Distance to self should be 0, got %d", dist)
	}

	// Get a neighbor and test distance
	neighbors, err := h.GridDisk(origin, 1)
	if err != nil {
		t.Fatalf("GridDisk failed: %v", err)
	}

	// Find a neighbor (not the origin itself)
	var neighbor H3Index
	for _, cell := range neighbors {
		if cell != origin {
			neighbor = cell
			break
		}
	}

	if neighbor == H3_NULL {
		t.Fatal("Could not find a neighbor cell")
	}

	dist, err = h.GridDistance(origin, neighbor)
	if err != nil {
		t.Fatalf("GridDistance to neighbor failed: %v", err)
	}

	if dist != 1 {
		t.Errorf("Distance to immediate neighbor should be 1, got %d", dist)
	}
}

// TestRoundTrip tests converting coordinates to H3 and back
func TestRoundTrip(t *testing.T) {
	h, err := New()
	if err != nil {
		t.Fatalf("Failed to create H3 instance: %v", err)
	}
	defer h.Close()

	testLocations := []struct {
		name string
		lat  float64
		lng  float64
	}{
		{"San Francisco", 37.7749, -122.4194},
		{"New York", 40.7128, -74.0060},
		{"Tokyo", 35.6762, 139.6503},
		{"London", 51.5074, -0.1278},
		{"Sydney", -33.8688, 151.2093},
	}

	for _, loc := range testLocations {
		t.Run(loc.name, func(t *testing.T) {
			index, err := h.LatLngToCell(loc.lat, loc.lng, 9)
			if err != nil {
				t.Fatalf("LatLngToCell failed: %v", err)
			}

			lat, lng, err := h.CellToLatLng(index)
			if err != nil {
				t.Fatalf("CellToLatLng failed: %v", err)
			}

			const tolerance = 0.01
			if math.Abs(lat-loc.lat) > tolerance || math.Abs(lng-loc.lng) > tolerance {
				t.Errorf("Round trip failed: (%f, %f) -> %s -> (%f, %f)",
					loc.lat, loc.lng, index.String(), lat, lng)
			}
		})
	}
}
