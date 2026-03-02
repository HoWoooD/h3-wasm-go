# Benchmarks

Результаты производительности для H3 WASM библиотеки.

## Окружение

- CPU: Apple M3 Max
- OS: macOS (darwin/arm64)  
- Go: 1.24.4
- wazero: v1.11.0

## H3 Operations Performance

```
BenchmarkH3_LatLngToCell-14       186414   6396 ns/op   58369 B/op  20 allocs/op
BenchmarkH3_CellToLatLng-14       152006   8327 ns/op   81681 B/op  28 allocs/op
BenchmarkH3_RoundTrip-14           83190  15016 ns/op  140051 B/op  48 allocs/op
BenchmarkH3_CellToBoundary-14      68841  17677 ns/op  175100 B/op  62 allocs/op
BenchmarkH3_GridDisk_K1-14        168274   6909 ns/op   70097 B/op  25 allocs/op
BenchmarkH3_GridDisk_K2-14        169635   7265 ns/op   70193 B/op  25 allocs/op
BenchmarkH3_GridDisk_K3-14        157735   7824 ns/op   70353 B/op  25 allocs/op
BenchmarkH3_GridDistance-14       353406   3541 ns/op   35024 B/op  12 allocs/op
BenchmarkH3_MultiResolution-14    169488   6551 ns/op   58369 B/op  20 allocs/op
```

### Производительность по операциям

| Операция | Время (ns/op) | Throughput (ops/sec) |
|----------|---------------|----------------------|
| LatLngToCell | 6,396 | ~156,000 |
| CellToLatLng | 8,327 | ~120,000 |
| Round-trip | 15,016 | ~66,600 |
| CellToBoundary | 17,677 | ~56,500 |
| GridDisk (k=1) | 6,909 | ~144,700 |
| GridDisk (k=2) | 7,265 | ~137,600 |
| GridDisk (k=3) | 7,824 | ~127,800 |
| GridDistance | 3,541 | ~282,400 |

### Выводы

- **Самая быстрая**: GridDistance (~3.5 µs)
- **Самая медленная**: CellToBoundary (~17.7 µs)
- **Базовые операции**: LatLngToCell/CellToLatLng ~6-8 µs
- **GridDisk**: Линейная сложность с k (~7-8 µs независимо от размера)

## Сравнение с другими библиотеками

```
BenchmarkGeohash-14    22002048    55 ns/op      32 B/op   2 allocs/op
BenchmarkS2-14          5085116   222 ns/op     224 B/op   5 allocs/op  
BenchmarkH3-14            89342 13130 ns/op  116960 B/op  45 allocs/op
```

### Относительная производительность

| Библиотека | Время (ns/op) | Относительно H3 | Аллокации |
|------------|---------------|-----------------|-----------|
| Geohash | 55 | **238x быстрее** | 2 |
| S2 | 222 | **59x быстрее** | 5 |
| H3 (WASM) | 13,130 | 1x (baseline) | 45 |

### Анализ

**Почему H3 через WASM медленнее?**

1. **WASM overhead**: Вызовы функций через WASM boundary (~50-100ns каждый)
2. **Memory copies**: Конвертация данных между Go и WASM памятью
3. **Координатные преобразования**: degrees ↔ radians (дополнительные WASM вызовы)
4. **Аллокации**: Каждая операция требует выделения памяти в WASM

**Когда использовать H3 через WASM?**

✅ **Хорошо подходит:**
- Анализ геоданных с batch обработкой
- Серверные приложения с умеренной нагрузкой (<1M ops/sec)
- Кросс-платформенные решения без CGo
- Когда нужны специфичные функции H3 (hexagons, grid algorithms)

❌ **Не подходит:**
- High-frequency trading или real-time системы
- Микросекундные latency требования
- Простая геохэш индексация (используйте geohash напрямую)

## Оптимизация

### Рекомендации для production:

1. **Reuse H3 instance**: Создавайте один экземпляр на всё приложение
2. **Batch operations**: Группируйте операции вместо отдельных вызовов
3. **Cache results**: Кэшируйте часто используемые индексы
4. **Pre-compute**: Предварительно вычисляйте H3 индексы для статических данных

### Потенциальные улучшения:

- [ ] Batch API для множественных конвертаций за один вызов
- [ ] Кэширование скомпилированного WASM модуля
- [ ] Прямое использование radians в Go API (опциональное)
- [ ] Memory pooling для уменьшения аллокаций

## Запуск бенчмарков

```bash
# Все H3 бенчмарки
go test -bench=BenchmarkH3 -benchmem

# Сравнение библиотек
go test -bench="Benchmark(Geohash|S2|H3)$" -benchmem

# Конкретная операция
go test -bench=BenchmarkH3_LatLngToCell -benchmem

# С профилированием CPU
go test -bench=BenchmarkH3_LatLngToCell -cpuprofile=cpu.prof

# С профилированием памяти  
go test -bench=BenchmarkH3_LatLngToCell -memprofile=mem.prof
```
