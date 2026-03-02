package main

import (
	"math/rand"
	"testing"

	"github.com/HoWoooD/h3-wasm-go/h3"
	"github.com/golang/geo/s2"
	"github.com/mmcloughlin/geohash"
)

var coords [][4]float64
var h3Instance *h3.H3
var h3Indices []h3.H3Index

// Подготовим тестовые данные (origin + destination)
func init() {
	n := 10000
	coords = make([][4]float64, n)
	for i := 0; i < n; i++ {
		// Москва +- небольшой разброс
		coords[i] = [4]float64{
			55.75 + rand.Float64()*0.1,
			37.61 + rand.Float64()*0.1,
			55.75 + rand.Float64()*0.1,
			37.61 + rand.Float64()*0.1,
		}
	}

	// Инициализируем H3 один раз для всех бенчмарков
	var err error
	h3Instance, err = h3.New()
	if err != nil {
		panic(err)
	}

	// Предварительно создаём H3 индексы для бенчмарков
	h3Indices = make([]h3.H3Index, n)
	for i := 0; i < n; i++ {
		c := coords[i]
		idx, err := h3Instance.LatLngToCell(c[0], c[1], 9)
		if err != nil {
			panic(err)
		}
		h3Indices[i] = idx
	}
}

func BenchmarkGeohash(b *testing.B) {
	b.ReportAllocs()
	precision := 7

	for i := 0; i < b.N; i++ {
		c := coords[i%len(coords)]

		origin := geohash.Encode(c[0], c[1])[:precision]
		dest := geohash.Encode(c[2], c[3])[:precision]

		_ = origin + ":" + dest
	}
}

func BenchmarkS2(b *testing.B) {
	b.ReportAllocs()
	level := 15 // примерно сопоставимо с geohash ~7

	for i := 0; i < b.N; i++ {
		c := coords[i%len(coords)]

		origin := s2.CellIDFromLatLng(
			s2.LatLngFromDegrees(c[0], c[1]),
		).Parent(level)

		dest := s2.CellIDFromLatLng(
			s2.LatLngFromDegrees(c[2], c[3]),
		).Parent(level)

		_ = origin.String() + ":" + dest.String()
	}
}

// ============================================================================
// H3 Benchmarks
// ============================================================================
//
// Результаты на Apple M3 Max (darwin/arm64):
//
// BenchmarkH3_LatLngToCell-14       186414  6396 ns/op   58369 B/op  20 allocs/op
// BenchmarkH3_CellToLatLng-14       152006  8327 ns/op   81681 B/op  28 allocs/op
// BenchmarkH3_RoundTrip-14           83190 15016 ns/op  140051 B/op  48 allocs/op
// BenchmarkH3_CellToBoundary-14      68841 17677 ns/op  175100 B/op  62 allocs/op
// BenchmarkH3_GridDisk_K1-14        168274  6909 ns/op   70097 B/op  25 allocs/op
// BenchmarkH3_GridDisk_K2-14        169635  7265 ns/op   70193 B/op  25 allocs/op
// BenchmarkH3_GridDisk_K3-14        157735  7824 ns/op   70353 B/op  25 allocs/op
// BenchmarkH3_GridDistance-14       353406  3541 ns/op   35024 B/op  12 allocs/op
// BenchmarkH3_MultiResolution-14    169488  6551 ns/op   58369 B/op  20 allocs/op
//
// Сравнение с другими библиотеками (координаты → индекс + конкатенация):
//
// BenchmarkGeohash-14    22002048    55 ns/op      32 B/op   2 allocs/op
// BenchmarkS2-14          5085116   222 ns/op     224 B/op   5 allocs/op
// BenchmarkH3-14            89342 13130 ns/op  116960 B/op  45 allocs/op
//
// Выводы:
// - H3 через WASM медленнее нативных библиотек (Geohash, S2)
// - Основные затраты: конвертация degrees↔radians и WASM boundary overhead
// - GridDistance самая быстрая операция (~3.5 µs)
// - CellToBoundary самая медленная (~17 µs) из-за чтения массива вершин
// - Производительность приемлема для большинства use-cases (>100K ops/sec)
//
// ============================================================================

// BenchmarkH3_LatLngToCell измеряет производительность конвертации координат в H3 индекс
func BenchmarkH3_LatLngToCell(b *testing.B) {
	b.ReportAllocs()
	resolution := 9 // средний уровень детализации

	for i := 0; i < b.N; i++ {
		c := coords[i%len(coords)]
		_, err := h3Instance.LatLngToCell(c[0], c[1], resolution)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkH3_CellToLatLng измеряет производительность конвертации H3 индекса в координаты
func BenchmarkH3_CellToLatLng(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		idx := h3Indices[i%len(h3Indices)]
		_, _, err := h3Instance.CellToLatLng(idx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkH3_RoundTrip измеряет производительность конвертации туда-обратно
func BenchmarkH3_RoundTrip(b *testing.B) {
	b.ReportAllocs()
	resolution := 9

	for i := 0; i < b.N; i++ {
		c := coords[i%len(coords)]
		
		// Координаты -> H3
		idx, err := h3Instance.LatLngToCell(c[0], c[1], resolution)
		if err != nil {
			b.Fatal(err)
		}

		// H3 -> Координаты
		_, _, err = h3Instance.CellToLatLng(idx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkH3_CellToBoundary измеряет получение границ ячейки
func BenchmarkH3_CellToBoundary(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		idx := h3Indices[i%len(h3Indices)]
		_, err := h3Instance.CellToBoundary(idx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkH3_GridDisk_K1 измеряет получение соседей (1 кольцо = 7 ячеек)
func BenchmarkH3_GridDisk_K1(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		idx := h3Indices[i%len(h3Indices)]
		_, err := h3Instance.GridDisk(idx, 1)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkH3_GridDisk_K2 измеряет получение соседей (2 кольца = 19 ячеек)
func BenchmarkH3_GridDisk_K2(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		idx := h3Indices[i%len(h3Indices)]
		_, err := h3Instance.GridDisk(idx, 2)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkH3_GridDisk_K3 измеряет получение соседей (3 кольца = 37 ячеек)
func BenchmarkH3_GridDisk_K3(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		idx := h3Indices[i%len(h3Indices)]
		_, err := h3Instance.GridDisk(idx, 3)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkH3_GridDistance измеряет вычисление расстояния между ячейками
func BenchmarkH3_GridDistance(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		origin := h3Indices[i%len(h3Indices)]
		dest := h3Indices[(i+1)%len(h3Indices)]
		
		// Не все пары могут иметь расстояние (разные резолюции или слишком далеко)
		_, _ = h3Instance.GridDistance(origin, dest)
	}
}

// BenchmarkH3_MultiResolution измеряет индексацию на разных уровнях
func BenchmarkH3_MultiResolution(b *testing.B) {
	b.ReportAllocs()
	resolutions := []int{5, 9, 12}

	for i := 0; i < b.N; i++ {
		c := coords[i%len(coords)]
		res := resolutions[i%len(resolutions)]
		
		_, err := h3Instance.LatLngToCell(c[0], c[1], res)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkH3_vs_Geohash сравнивает H3 с Geohash
func BenchmarkH3(b *testing.B) {
	b.ReportAllocs()
	resolution := 9

	for i := 0; i < b.N; i++ {
		c := coords[i%len(coords)]

		origin, err := h3Instance.LatLngToCell(c[0], c[1], resolution)
		if err != nil {
			b.Fatal(err)
		}

		dest, err := h3Instance.LatLngToCell(c[2], c[3], resolution)
		if err != nil {
			b.Fatal(err)
		}

		_ = origin.String() + ":" + dest.String()
	}
}
