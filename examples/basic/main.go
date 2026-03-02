package main

import (
	"fmt"
	"log"

	"github.com/HoWoooD/h3-wasm-go/h3"
)

func main() {
	// Create H3 instance
	h, err := h3.New()
	if err != nil {
		log.Fatalf("Failed to create H3 instance: %v", err)
	}
	defer h.Close()

	fmt.Println("🌍 H3 Geospatial Indexing Demo\n")

	// Example 1: Convert coordinates to H3 index
	fmt.Println("Example 1: Coordinates to H3 Index")
	fmt.Println("=====================================")
	
	lat, lng := 37.7749, -122.4194 // San Francisco
	resolution := 9

	index, err := h.LatLngToCell(lat, lng, resolution)
	if err != nil {
		log.Fatalf("LatLngToCell failed: %v", err)
	}

	fmt.Printf("Location: San Francisco (%.4f, %.4f)\n", lat, lng)
	fmt.Printf("Resolution: %d\n", resolution)
	fmt.Printf("H3 Index: %s\n\n", index.String())

	// Example 2: Convert H3 index back to coordinates
	fmt.Println("Example 2: H3 Index to Coordinates")
	fmt.Println("====================================")

	centerLat, centerLng, err := h.CellToLatLng(index)
	if err != nil {
		log.Fatalf("CellToLatLng failed: %v", err)
	}

	fmt.Printf("H3 Index: %s\n", index.String())
	fmt.Printf("Cell Center: (%.6f, %.6f)\n\n", centerLat, centerLng)

	// Example 3: Get cell boundary
	fmt.Println("Example 3: Cell Boundary Vertices")
	fmt.Println("===================================")

	boundary, err := h.CellToBoundary(index)
	if err != nil {
		log.Fatalf("CellToBoundary failed: %v", err)
	}

	fmt.Printf("Number of vertices: %d\n", len(boundary.Vertices))
	for i, vertex := range boundary.Vertices {
		fmt.Printf("  Vertex %d: (%.6f, %.6f)\n", i+1, vertex.Lat, vertex.Lng)
	}
	fmt.Println()

	// Example 4: Get neighboring cells
	fmt.Println("Example 4: Grid Disk (Neighbors)")
	fmt.Println("==================================")

	k := 1 // One ring of neighbors
	neighbors, err := h.GridDisk(index, k)
	if err != nil {
		log.Fatalf("GridDisk failed: %v", err)
	}

	fmt.Printf("Origin: %s\n", index.String())
	fmt.Printf("Neighbors within %d ring(s): %d cells\n", k, len(neighbors))
	for i, neighbor := range neighbors {
		if neighbor == index {
			fmt.Printf("  %d. %s (origin)\n", i+1, neighbor.String())
		} else {
			fmt.Printf("  %d. %s\n", i+1, neighbor.String())
		}
	}
	fmt.Println()

	// Example 5: Calculate distance between cells
	fmt.Println("Example 5: Grid Distance")
	fmt.Println("=========================")

	// Get a neighbor
	var neighbor h3.H3Index
	for _, cell := range neighbors {
		if cell != index {
			neighbor = cell
			break
		}
	}

	distance, err := h.GridDistance(index, neighbor)
	if err != nil {
		log.Fatalf("GridDistance failed: %v", err)
	}

	fmt.Printf("From: %s\n", index.String())
	fmt.Printf("To:   %s\n", neighbor.String())
	fmt.Printf("Grid Distance: %d\n\n", distance)

	// Example 6: Multi-resolution demonstration
	fmt.Println("Example 6: Multi-Resolution Indexing")
	fmt.Println("======================================")

	resolutions := []int{0, 5, 10, 15}
	for _, res := range resolutions {
		idx, err := h.LatLngToCell(lat, lng, res)
		if err != nil {
			log.Printf("Resolution %d failed: %v", res, err)
			continue
		}
		fmt.Printf("Resolution %2d: %s\n", res, idx.String())
	}

	fmt.Println("\n✅ Demo completed successfully!")
}
