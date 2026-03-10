# Go Checklist

Use this checklist before accepting Go code in this repository.

## Architecture

- Handler is thin and only handles transport concerns
- Business logic lives in services, not handlers
- Consumer-side interfaces are small and necessary
- Code respects hexagonal boundaries and does not couple core logic to `net/http` or `database/sql` unnecessarily

## Correctness

- No variable shadowing that changes control flow unexpectedly
- No deep nesting where early returns would be clearer
- Function inputs are validated early
- Code returns immediately after terminal HTTP responses such as `http.Error`
- Range loops do not rely on copied values when mutation of the original element is intended
- No pointers are taken to range loop variables
- No reliance on map iteration order

## Errors

- Errors are wrapped with `%w` when propagating
- `errors.Is` or `errors.As` is used where error inspection matters
- Errors are not logged repeatedly at multiple layers without added context
- Ignored errors are clearly intentional
- Close or flush errors are handled when they affect correctness
- `rows.Err()` is checked after iterating SQL rows

## Memory

- No `os.ReadFile`, `io.ReadAll`, or full-file materialization for large inputs unless explicitly justified
- Streaming APIs such as `io.Reader` are preferred
- Slice or substring operations do not accidentally retain large backing storage
- Slice length and capacity are chosen deliberately
- Appending to a subslice cannot accidentally corrupt shared backing arrays
- No unnecessary string-byte-string conversions
- Compact structs are preferred over `map[string]any` for known shapes

## Concurrency

- Concurrency is justified for the workload rather than added by default
- `WaitGroup.Add` happens before launching goroutines
- Loop variables are rebound before use in goroutines
- No data races on shared slices, maps, or structs
- Sync primitives are not copied after first use
- Long-lived goroutines have cancellation or shutdown behavior
- Concurrency is bounded when work increases RAM, CPU, or network pressure

## HTTP and SQL

- HTTP clients have explicit timeouts
- Server timeouts are configured deliberately where applicable
- Response bodies are always closed
- Response bodies are drained when connection reuse matters
- Status codes are validated before processing response bodies
- `sql.Open` is followed by connection verification when needed
- Nullable DB fields use explicit nullable handling where appropriate

## Testing

- Core logic has unit tests
- Filters and transformations use table-driven tests
- Slow or integration tests are marked with `testing.Short()` or equivalent separation
- Tests do not rely on sleeps when channels, mocks, or helpers can make them deterministic
- `httptest`, `iotest`, and `t.Cleanup` are used where they simplify tests

## Repository-Specific

- Parsing is streaming and memory-bounded
- Upload or body size limits are enforced before expensive processing
- In-memory SQLite is isolated per request
- Concurrency limits protect the process from RAM exhaustion
- Timeline offset and calibration metadata are preserved where relevant
- Backend returns timestamps or clip ranges, not rendered media
