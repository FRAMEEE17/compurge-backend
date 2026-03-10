# Go Practices

## Purpose

This document defines the Go implementation practices that AI coding assistants must follow in this repository.

Use this file as the Go-specific companion to [`agent.md`](/Users/palmer/compurge/agent.md).
Use the example catalog under [`go_practices/src`](/Users/palmer/compurge/go_practices/src) as supporting reference material, not as a rule file.

This rule set has been distilled from the nested examples under `go_practices/src`, including subdirectories for project organization, data types, errors, concurrency, standard library usage, testing, and optimizations.

## Core Rules

- Prefer simple, idiomatic Go over abstraction-heavy designs
- Use early returns to avoid deep nesting
- Wrap errors with context using `%w`
- Keep handlers thin and move business logic into services
- Avoid unnecessary interfaces; define interfaces at the consumer side
- Pass `context.Context` through request-scoped and cancelable operations
- Write table-driven tests for business logic and filters
- Minimize allocations and avoid unnecessary copies
- Prefer explicit code over magic or reflection-heavy patterns

## Project Organization

- Avoid variable shadowing in non-trivial flows
- Avoid deep nesting; split complex branches into helper functions
- Do not rely on `init()` for important runtime behavior
- Do not introduce utility packages without a clear shared need
- Use embedding sparingly and only when it improves clarity
- Use generics only when they reduce duplication without hiding logic
- Avoid `any` when a concrete type is known
- Collapse repeated branch-local assignments into one post-branch error check when it improves clarity
- Prefer plain config structs for simple construction; use functional options only when configuration is genuinely optional or likely to grow
- Do not embed types only to imitate inheritance or expose implementation details accidentally
- Be careful embedding types that contain locks or externally visible methods

Desired:

- assign branch-specific values, then perform one shared `if err != nil` check after the branch

Avoid:

- redeclaring `client, err := ...` in both branches when the outer values are needed later

## Interfaces

- Define interfaces where they are consumed, not where they are implemented
- Do not create interfaces for a single implementation unless it improves testing or architectural boundaries
- Keep interfaces small and behavior-focused
- Prefer concrete structs in internal code paths unless abstraction is necessary
- Do not publish fat storage interfaces from provider packages
- Restrict behavior with the minimum method set the consumer actually needs

Desired:

- `type customerStorer interface { StoreCustomer(Customer) error }`

Avoid:

- large repository interfaces that mix reads, writes, listing, and unrelated queries when a consumer only needs one method

## Data Types and Memory

- Be explicit about slice length and capacity when it affects allocation behavior
- Avoid retaining large backing arrays accidentally through slicing or substring operations
- Copy slices only when ownership or mutation safety requires it
- Initialize maps before writes
- Be careful with integer overflow and floating-point comparisons
- Normalize source data into compact structs rather than wide generic maps
- When keeping only a small prefix of a large slice or string, clone or copy the needed subset so the large backing storage can be released
- Avoid returning pointers to short-lived local values when a plain value return is sufficient
- Prefer `len(s) == 0` over `s != nil` when checking whether a slice is empty
- Distinguish `nil` versus empty slices only when external behavior depends on it, such as JSON encoding
- When building a result slice from a known input size, preallocate length or capacity deliberately
- Be aware that appending to a subslice may mutate the original backing array unless capacity is constrained or data is copied
- Deleting entries from a map does not guarantee that memory is returned to the runtime immediately
- Avoid relying on `reflect.DeepEqual` as a default equality strategy for domain logic

Examples of accidental retention to avoid:

- `msg[:5]` kept after `msg` was a 1 MB buffer
- `log[:36]` kept when only a UUID-sized prefix is needed
- substrings or sub-slices stored long-term without copying

## Control Flow

- Do not rely on map iteration order
- Be careful with `defer` inside loops
- Avoid subtle range-loop behavior with copied values or pointer capture
- Break complex loops into smaller functions when the flow becomes hard to follow
- Return immediately after `http.Error` or other terminal HTTP response writes
- Be explicit with labeled `break` or `return` when leaving nested loops or `select` blocks
- Remember that `range` evaluates its operand once; mutating the ranged collection may not affect the loop the way you expect
- When iterating over slices of structs, updating the range variable does not mutate the original element
- Do not take the address of a range loop variable when storing pointers; use an index or a loop-local copy

Avoid:

- writing an error response and then continuing to write a success body or status

## Strings and Conversions

- Be explicit about rune versus byte behavior
- Use `strings.Builder` or equivalent patterns for repeated string concatenation
- Avoid repeated string-byte-string conversions without a reason
- Trim and normalize external string input before matching headers or enum-like values
- When slicing strings for long-term storage, clone or copy the substring if it would otherwise retain a much larger source string
- Use byte-oriented helpers such as `bytes.TrimSpace` when data is already in `[]byte`
- Prefer `TrimPrefix` and `TrimSuffix` over broader trim functions when removing exact prefixes or suffixes

## Functions and Methods

- Avoid named result parameters unless they materially improve readability
- Avoid side effects on named return values
- Validate function inputs early
- Use pointer or value receivers intentionally; do not mix without reason
- Be explicit about deferred call behavior and argument evaluation
- Do not use value receivers on structs containing mutexes or other sync primitives when mutation or locking is intended
- Accept `io.Reader` or other narrow abstractions instead of filenames or concrete resources when the function only needs a stream
- Be careful returning typed nil pointers as `error`; return plain `nil` when there is no error
- Remember that deferred function arguments are evaluated at defer time, while deferred closures observe later variable updates

## Error Handling

- Do not use `panic` for normal application errors
- Wrap errors with context
- Use `errors.Is` and `errors.As` instead of brittle direct comparisons when appropriate
- Do not handle the same error multiple times at multiple layers unless each layer adds meaning
- Never ignore returned errors unless it is a deliberate and documented choice
- Check close or flush errors where they matter
- Prefer one layer to add context and a higher layer to log, rather than logging and returning the same error repeatedly
- After iterating `sql.Rows`, always check `rows.Err()`
- When a deferred close can affect correctness, propagate its error instead of silently discarding it
- Ignoring an error is only acceptable when the code comment explains why best-effort behavior is intended
- Prefer `%w` over `%v` when the wrapped error must remain inspectable

Desired:

- `return fmt.Errorf("validate target coordinates: %w", err)`

Avoid:

- logging inside a low-level validator and then returning the same error to be logged again upstream

## Concurrency

- Do not add concurrency unless it improves latency or throughput for this workload
- Respect workload type; CPU-bound and IO-bound work need different strategies
- Pass contexts through goroutines and outbound calls
- Do not capture loop variables incorrectly in goroutines
- Avoid data races on slices, maps, and shared structs
- Do not copy `sync.Mutex`, `sync.RWMutex`, `sync.WaitGroup`, or similar sync primitives after first use
- Prefer `errgroup` or clear coordination patterns for multi-step concurrent work
- Use bounded concurrency when work can amplify RAM or network pressure
- Use loop-local copies or function parameters when launching goroutines inside loops
- Do not store contexts in structs for later reuse unless there is a clear lifecycle reason
- Use request-scoped timeouts for outbound or bounded work, not `context.Background()` deep inside the domain without reason
- Avoid goroutines without a shutdown story; long-lived watchers need cancellation or explicit close semantics
- Call `WaitGroup.Add` before starting the goroutine, not inside it
- Do not read or format mutex-protected structs in ways that reacquire the same lock while already locked
- Copy shared slices before appending from multiple goroutines
- If a channel is closed, set it to `nil` in a `select` loop when you need to disable that case cleanly
- Prefer `sync.Cond` only when simpler channel-based coordination is not expressive enough
- Use atomic operations only for truly atomic state; they do not replace broader synchronization needs

## Standard Library Usage

- Use explicit HTTP clients with timeouts; do not rely blindly on defaults
- Always close response bodies and other resources
- Validate HTTP status codes before treating a response as success
- Use `database/sql` carefully and close rows consistently
- Be explicit with time units and durations
- Avoid `time.After` in hot loops when timer reuse is more appropriate
- Configure both client-side and server-side HTTP timeouts deliberately
- Drain or discard HTTP response bodies when connection reuse matters
- In SQL code, distinguish nullable values explicitly instead of assuming zero values mean null
- Prefer prepared statements only when they are reused enough to justify them
- Call `db.Ping()` or equivalent health verification after `sql.Open` when connection validation matters
- Be careful with JSON encoding of embedded types such as `time.Time`; embedding can flatten or distort the output shape
- Do not default to `map[string]any` for JSON payloads when a stable struct shape is known
- Truncate monotonic clock data from `time.Time` before equality-sensitive serialization checks when needed

## Testing

- Prefer table-driven tests for parsing, filtering, and transformation logic
- Separate fast unit tests from slower integration tests
- Avoid sleep-based tests when deterministic coordination is possible
- Make time-dependent code testable by injecting clocks or time sources when needed
- Benchmark only where performance matters and interpret results carefully
- When using table-driven tests with subtests and parallel execution, rebind the loop variable inside the loop
- Use `testing.Short()` or build tags to separate long-running or integration-style tests
- Prefer synchronization primitives, channels, `httptest`, and `iotest` helpers over sleeps and ad hoc timing
- Use `t.Cleanup` for scoped teardown when it keeps setup and cleanup local to the test
- Small test helper functions may call `t.Fatal` directly when that reduces repetitive error plumbing
- Be cautious interpreting benchmark results; compiler optimizations, timer scope, and observer effects can mislead

## Repository-Specific Emphasis

For this repository, pay special attention to these practice areas:

- low-memory streaming parsers
- error wrapping and safe HTTP error translation
- request-scoped SQLite isolation
- bounded concurrency
- context-aware HTTP fetches with backoff
- table-driven tests for filters and timestamp logic
- avoiding accidental memory retention in slices and strings
- returning after terminal HTTP responses
- checking `rows.Err()` and close semantics in SQLite code
- avoiding copied mutexes or shared-state races in request coordination
- avoiding pointer returns or copies that increase heap pressure without benefit
- preferring compact structs and predictable memory layouts only when profiling shows it matters
- keeping optimization work subordinate to correctness, simplicity, and measured bottlenecks

## Reference Mapping

The following folders in [`go_practices/src`](/Users/palmer/compurge/go_practices/src) are especially relevant to this repository:

- `02-code-project-organization`
- `03-data-types`
- `07-error-management`
- `08-concurrency-foundations`
- `09-concurrency-practice`
- `10-standard-lib`
- `11-testing`
- `12-optimizations`

Use those examples to clarify edge cases and anti-patterns, but follow the rules in this file when generating code.

## Revisit Only When Needed

These example areas usually do not need to be re-read for routine backend work, but they are worth revisiting when implementation details require them:

- [`go_practices/src/11-testing/87-time-api/`](/Users/palmer/compurge/go_practices/src/11-testing/87-time-api) for deterministic time-dependent tests
- [`go_practices/src/11-testing/89-benchmark/`](/Users/palmer/compurge/go_practices/src/11-testing/89-benchmark) for benchmark design and interpretation
- [`go_practices/src/12-optimizations/91-cpu-caches/`](/Users/palmer/compurge/go_practices/src/12-optimizations/91-cpu-caches) for cache behavior tradeoffs
- [`go_practices/src/12-optimizations/92-false-sharing/`](/Users/palmer/compurge/go_practices/src/12-optimizations/92-false-sharing) for concurrent counter and struct-layout issues
- [`go_practices/src/12-optimizations/93-instruction-level-parallelism/`](/Users/palmer/compurge/go_practices/src/12-optimizations/93-instruction-level-parallelism) for tight-loop optimization details
- [`go_practices/src/12-optimizations/94-data-alignment/`](/Users/palmer/compurge/go_practices/src/12-optimizations/94-data-alignment) for struct layout and padding concerns
- [`go_practices/src/12-optimizations/95-stack-heap/`](/Users/palmer/compurge/go_practices/src/12-optimizations/95-stack-heap) for escape-analysis and heap-pressure questions
- [`go_practices/src/12-optimizations/96-reduce-allocations/`](/Users/palmer/compurge/go_practices/src/12-optimizations/96-reduce-allocations) for allocation tuning after profiling
- [`go_practices/src/10-standard-lib/77-json-handling/`](/Users/palmer/compurge/go_practices/src/10-standard-lib/77-json-handling) for JSON shape, embedded types, and time serialization issues
- [`go_practices/src/10-standard-lib/78-sql/`](/Users/palmer/compurge/go_practices/src/10-standard-lib/78-sql) for SQL scanning, null handling, and query iteration details
- [`go_practices/src/09-concurrency-practice/72-cond/`](/Users/palmer/compurge/go_practices/src/09-concurrency-practice/72-cond) for advanced coordination cases
- [`go_practices/src/09-concurrency-practice/73-errgroup/`](/Users/palmer/compurge/go_practices/src/09-concurrency-practice/73-errgroup) for fan-out work with shared cancellation and error aggregation
- [`go_practices/src/02-code-project-organization/11-functional-options/`](/Users/palmer/compurge/go_practices/src/02-code-project-organization/11-functional-options) for constructor and configuration API design
