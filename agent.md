# agent.md

## Purpose

This repository is the backend API and data orchestrator for an open source football (soccer) video highlight generator.

Always treat this codebase as a timestamp generation service.

Its job is to:

1. Accept API requests with:
   - a video URL
   - event data in CSV, XLSX, or Parquet
   - filters such as player, stat, half, period, or event type
2. Parse and normalize event data
3. Apply filtering logic
4. Return timestamps or timestamp ranges to the frontend

Never treat this repository as a video processing system.

Detailed event-data schema notes, source mappings, and provider-specific quirks belong in [`docs/event-data-schema.md`](/Users/palmer/compurge/docs/event-data-schema.md), not in this file.
Detailed Go implementation conventions and anti-patterns belong in [`docs/golang/go-practices.md`](/Users/palmer/compurge/docs/golang/go-practices.md). Follow that document when generating or modifying Go code in this repository.
Detailed naming rules belong in [`docs/golang/naming-conventions.md`](/Users/palmer/compurge/docs/golang/naming-conventions.md). Follow that document when naming Go packages, files, types, functions, variables, constants, tests, and docs.
Current HTTP API behavior and examples belong in [`docs/api/timestamps_api.md`](/Users/palmer/compurge/docs/api/timestamps_api.md), not in this file.
Use [`docs/golang/go-checklist.md`](/Users/palmer/compurge/docs/golang/go-checklist.md) as a final self-check before accepting Go changes.
The example tree under [`go_practices/src`](/Users/palmer/compurge/go_practices/src) is supporting reference material only. Do not treat those example files as repository code to copy blindly; apply the distilled rules from [`docs/golang/go-practices.md`](/Users/palmer/compurge/docs/golang/go-practices.md) instead.

## Core Architectural Rules

### API Boundary

Always keep this repository focused on:

- HTTP APIs
- request validation
- data ingestion
- data normalization
- filtering
- orchestration
- timestamp generation

Never add:

- server-side FFmpeg
- video transcoding
- clip generation
- media muxing
- rendering pipelines

The frontend handles video processing via client-side FFmpeg. The backend only returns timestamps and related metadata.

### Project Structure

Follow standard Go project layout:

- `/cmd/api`: application bootstrap and wiring
- `/internal/handler`: HTTP handlers and transport logic
- `/internal/service`: business logic and orchestration
- `/internal/repository`: storage and query logic
- `/internal/parser`: streaming parsers and normalization
- `/internal/model`: domain models and DTOs

Always keep handlers thin.
Never put business logic in handlers.
Never couple parsers directly to HTTP request structs.

### Naming Rules

Use clear, stable, descriptive names across code and documentation.

Always:

- choose names that explain purpose, not just type
- prefer names that are searchable and unambiguous
- use one naming pattern consistently for the same kind of thing
- name files and symbols so a new contributor can infer intent without opening every file

Never:

- use vague names such as `tmp`, `misc`, `helper`, `util`, `data`, `thing`, `final`, or `new` unless the scope is truly tiny and local
- create multiple files whose names differ only by vague suffixes
- encode transient status in filenames such as `final_v2_really_final`

Follow repository naming rules for:

- Go packages and source files
- types, functions, variables, and constants
- tests and benchmarks
- docs, specs, and sample data files

### Implementation Planning

For non-trivial features or refactors, create or update a short implementation plan before major code changes.

Always:

- break substantial work into small verifiable steps
- implement incrementally
- validate each phase with tests or explicit checks

Do not treat a large multi-part change as one undifferentiated code edit when the work can be phased safely.

### Hexagonal Architecture

Follow a hexagonal architecture style inside the standard Go layout.

Treat the domain and application logic as the core.
Treat HTTP, SQLite, remote file access, and parsers as adapters around that core.

Use these rules:

- define inbound ports for application use cases when needed
- define outbound ports for persistence, remote fetch, or external integrations
- keep adapter code in transport, repository, or parser packages
- keep business rules independent from `net/http`, SQLite-specific details, and file format libraries

Always make the service layer depend on abstractions owned by the core, not concrete adapter implementations.
Never let domain logic import handler packages.
Never let domain logic depend directly on `database/sql`, `chi`, or HTTP request/response types.

Desired:

```go
type EventSource interface {
	Open(ctx context.Context, ref SourceRef) (io.ReadCloser, error)
}

type TimestampService struct {
	source EventSource
}
```

Avoid:

```go
type TimestampService struct {
	db  *sql.DB
	req *http.Request
}
```

Prefer dependency injection through constructors.
Keep adapters replaceable so the same service logic can be tested without HTTP servers, SQLite, or network calls.

### Request Pipeline

Design flows in this order:

1. Validate request early
2. Open remote or uploaded data as a stream
3. Parse incrementally
4. Normalize to a minimal internal schema
5. Apply filters
6. Merge or deduplicate timestamps if needed
7. Return compact JSON

Prefer single-pass or bounded-memory algorithms.

### Video Workflow Constraints

Treat video as an external media source handled primarily by the frontend runtime.
This backend may accept a video reference, but it must not assume that video access is uniform or reliable across environments.

Use these rules:

- support local full-match video as a first-class workflow
- treat remote full-match video as a best-effort workflow with strict caveats
- return clip timestamps or ranges, not rendered media
- treat timeline calibration metadata as first-class input and output
- treat first-half and second-half alignment as potentially different calibration problems

#### Local Full-Match Video

Assume local full-match video is the primary supported path for large files.
Local files avoid CORS, remote auth expiry, and range-request variability.

Design API responses so the frontend can use a local full-match file with:

- raw event timestamps
- resolved clip ranges
- calibration or timeline offset metadata

Do not assume local full-match files start exactly at match `00:00`.
Always allow timeline offset or calibration data to adjust event time to video time.
Do not assume one global offset is always sufficient for the whole match.
If the source video includes halftime, pre-roll, missing lead-in, or a discontinuity between halves, support separate calibration for first half and second half.

Treat `period` as a timeline-critical field, not just a display label.
Second-half events may require different resolution logic than first-half events, especially when the source video includes halftime or trimmed segments.

#### Remote Full-Match Video

Treat remote full-match video as reference-only unless proven fetchable in the client.
Do not assume a remote video URL will work in the browser.

Remote full-match access may fail because of:

- CORS restrictions
- missing `Accept-Ranges: bytes`
- expired signed URLs
- auth headers the browser cannot send
- origin rate limits
- unstable network conditions

Do not design backend behavior that depends on the frontend being able to fetch arbitrary remote video URLs.
If a remote URL is accepted, treat it as a hint or source reference, not a guarantee of successful client-side clipping.

#### Where Chunking Happens

Chunking belongs in the frontend video-processing pipeline, not in the backend event-processing pipeline.

The backend should:

- return normalized timestamps or clip ranges
- optionally return pre-roll and post-roll-expanded ranges
- optionally return sorted processing jobs for the frontend

The frontend should:

- process only selected highlight ranges
- fetch or read only the byte ranges or time ranges needed for one clip at a time
- run sequential extraction for large full-match sources

Do not design the frontend to load an entire full-match video into FFmpeg.wasm memory before clipping.
For full-match inputs, prefer sequential clip extraction over whole-file processing.

#### Where OPFS Is Used

Use OPFS as disk-backed temporary storage for frontend-generated clip artifacts.
OPFS is the correct place to stage intermediate clips, concat lists, and output files when browser-side video processing is required.

Use OPFS for:

- storing extracted clip files between processing steps
- storing concat input lists or manifests
- storing zipped clip bundles before download
- reducing pressure on browser RAM during multi-clip jobs

Do not use RAM-only storage for multi-clip workflows when OPFS is available.
Do not assume OPFS is available in every browser without feature detection.

#### Where FFmpeg.wasm Should Hold Memory

FFmpeg.wasm should hold only the working set for the current clip or current processing step.

Allowed in FFmpeg.wasm memory:

- one active clip window
- transient decode or encode buffers
- small command manifests required for the current operation

Do not keep these in FFmpeg.wasm memory longer than necessary:

- the entire full-match video
- many extracted clips at once
- a full compilation plus all source clips simultaneously
- large temporary artifacts that can be moved to OPFS

The correct frontend pattern for large sources is:

1. read or fetch one clip range
2. process one clip
3. move the result to OPFS
4. clear temporary FFmpeg.wasm filesystem state
5. continue with the next clip

Prefer `zip of clips` or separate clip download as the default product output.
Treat full concat into one compilation file as an optional, higher-risk workflow.

## Golang Standards

### General

Always write idiomatic Go.
Prefer the standard library plus small focused packages.
Use `go-chi/chi` with `net/http`.
Avoid heavy frameworks and unnecessary abstractions.

Go is the primary language. Python is allowed only for isolated data-processing scripts when absolutely necessary.

### Control Flow

Always use early returns.
Avoid deep nesting.

Desired:

```go
if err := req.Validate(); err != nil {
	return nil, fmt.Errorf("validate request: %w", err)
}
```

Avoid:

```go
if err := req.Validate(); err == nil {
	// nested flow
}
```

### JSON and Types

Always use explicit JSON struct tags with `camelCase`.

Desired:

```go
type TimestampResponse struct {
	Timestamps []float64 `json:"timestamps"`
	Source     string    `json:"source"`
}
```

Prefer clear concrete structs over `map[string]any` when schema is known.

### Errors

Always wrap errors with context.

Desired:

```go
return fmt.Errorf("parse parquet row: %w", err)
```

Never leak internal paths, SQL details, or stack traces in HTTP responses.
Log internal context safely, but return sanitized client-facing errors.

## Memory & Space Complexity (CRITICAL)

### Prime Directive

Always optimize for the lowest practical space complexity.
Memory footprint is a primary design constraint, not a secondary optimization.

Do not confuse stream-based file reading with bounded total memory usage.
If streamed rows are inserted into in-memory SQLite, total RAM usage still grows with dataset size.
Assume effective space complexity becomes `O(N)` once rows are stored in SQLite.

### Streaming Only

Never load entire CSV, XLSX, or Parquet files into memory unless there is no viable alternative and the reason is documented.

Always prefer:

- `io.Reader`
- `bufio.Reader`
- streaming decoders
- row-by-row processing
- chunked processing

Never default to:

- `os.ReadFile`
- `io.ReadAll`
- `ReadAll()` style CSV parsing
- storing whole remote responses in `[]byte`

Streaming input is necessary but not sufficient.
You must also cap request size before data reaches SQLite, or valid large files can still exhaust server RAM.

Desired:

```go
r := csv.NewReader(bufio.NewReader(src))
for {
	record, err := r.Read()
	if err == io.EOF {
		break
	}
	if err != nil {
		return fmt.Errorf("read csv record: %w", err)
	}

	if err := consumeRecord(record); err != nil {
		return fmt.Errorf("consume csv record: %w", err)
	}
}
```

Avoid:

```go
data, err := os.ReadFile(path)
if err != nil {
	return err
}

records, err := csv.NewReader(bytes.NewReader(data)).ReadAll()
if err != nil {
	return err
}
```

### Allocation Discipline

Always minimize allocations.

- Reuse buffers where safe
- Avoid duplicate in-memory representations of the same dataset
- Avoid unnecessary `string` and `[]byte` conversions
- Keep internal event structs compact
- Prefer bounded-memory merge and filter logic

Never materialize entire datasets if filtering can occur during parsing.

### Mandatory Request Size Limits

Always enforce upload and body size limits at the HTTP boundary.
Reject oversized requests before parsing or inserting data into SQLite.

Treat size limits as a safety control, not a convenience feature.

Desired:

```go
const maxUploadSize = 10 << 20 // 10 MB

r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
if err := r.ParseMultipartForm(maxUploadSize); err != nil {
	http.Error(w, "file too large", http.StatusBadRequest)
	return
}
```

Avoid:

```go
if err := r.ParseMultipartForm(128 << 20); err != nil {
	return err
}
```

Do not accept unbounded multipart uploads.
Do not defer size enforcement to parser code or SQLite insertion code.
When limits are exceeded, return `400 Bad Request` or `413 Request Entity Too Large` with a safe message.

## Resilience

### External Calls

Any external HTTP call or remote file fetch must:

1. Use `context.Context`
2. Use `cenkalti/backoff/v4`
3. Respect timeouts
4. Validate status codes
5. Close response bodies

Desired:

```go
op := func() error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return backoff.Permanent(fmt.Errorf("build request: %w", err))
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("perform request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		return backoff.Permanent(fmt.Errorf("unexpected status: %d", resp.StatusCode))
	}
	if resp.StatusCode >= 500 {
		return fmt.Errorf("server status: %d", resp.StatusCode)
	}

	return parseStream(resp.Body)
}

if err := backoff.Retry(op, backoff.NewExponentialBackOff()); err != nil {
	return fmt.Errorf("fetch remote data: %w", err)
}
```

Never implement ad hoc retry loops when `backoff` should be used.
Never retry malformed input or deterministic parse failures.

### Database

Use in-memory SQLite only, but isolate it per request.

- Driver: `mattn/go-sqlite3`

Never write application data to physical disk.
Never introduce persistent local SQLite files.

Never share one in-memory SQLite namespace across unrelated requests.
Never use a single global DSN like `file::memory:?cache=shared` for all users.

Every request that needs SQLite must get its own unique in-memory database name.
Close the connection when the request ends so that request-scoped data is released from RAM.

Desired:

```go
dbName := fmt.Sprintf("file:req_%s?mode=memory&cache=shared", requestID)
db, err := sql.Open("sqlite3", dbName)
if err != nil {
	return fmt.Errorf("open request db: %w", err)
}
defer db.Close()
```

Avoid:

```go
db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
if err != nil {
	return err
}
```

Assume a shared in-memory DSN can cause cross-request data leakage under concurrency.
Data isolation is mandatory, not optional.

### Concurrency Guardrails

Always protect the server from too many concurrent in-memory workloads.
In-memory SQLite is a vertical-scaling bottleneck because each active request consumes RAM.

Enforce a bounded concurrency limit for parse-and-query workloads.
If the system can safely process 50 requests concurrently, queue or reject the rest.

Desired:

```go
var ingestSem = make(chan struct{}, 50)

func withIngestSlot(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case ingestSem <- struct{}{}:
			defer func() { <-ingestSem }()
			next.ServeHTTP(w, r)
		default:
			http.Error(w, "server busy", http.StatusTooManyRequests)
		}
	})
}
```

Do not allow unbounded concurrent requests to each allocate independent in-memory SQLite databases.
Do not assume horizontal traffic spikes can be absorbed without explicit admission control.

### Staging Table ELT Pattern

When input schemas are unknown or inconsistent, use a small-scale ELT pattern:

1. Read the header row first
2. Create a raw staging table dynamically
3. Declare all staging columns as `TEXT`
4. Insert streamed rows into the staging table
5. Transform and cast with SQL into the normalized query shape

Always load raw staging data as text first when column types are not yet trusted.
This prevents early insert failures caused by inconsistent source typing.

Desired:

```go
_, err := db.Exec(`
	CREATE TABLE raw_staging (
		col1 TEXT,
		col2 TEXT,
		col3 TEXT
	)
`)
if err != nil {
	return fmt.Errorf("create raw staging: %w", err)
}
```

Then transform with SQL:

```sql
SELECT
	col1 AS match_minute,
	CAST(col2 AS INTEGER) AS player_id
FROM raw_staging
WHERE col3 = 'Pass';
```

Do not try to guess final numeric or temporal types during the first streaming insert if the source format is not stable.
Do not couple raw ingestion directly to the final domain schema when auto-reformatting is required.

### Event Time and Half Semantics

Treat `minute`, `second`, and `period` as timeline-critical fields.

Always:

- normalize raw event time into a consistent match-time representation
- preserve the original `period`
- resolve first-half and second-half events deliberately

Do not assume second-half timestamps can be mapped with naive arithmetic alone.
If the source video includes halftime, pre-roll, broadcast padding, or trims, allow first-half and second-half offsets to be calibrated independently.

### Quality Filters

Treat `type` and `outcomeType` as first-class filter inputs.

Always allow workflows such as:

- only successful passes
- only unsuccessful carries
- only specific event categories within one half

Do not treat `outcomeType` as optional noise when the source includes it.
Analyst workflows often depend on separating successful and unsuccessful actions.

### Spatial and Zone Filters

Treat `x`, `y`, `endX`, and `endY` as tactical filter fields, not just metadata.

Always preserve these fields when the source provides them.
Design the normalized model so spatial queries can support:

- start-zone filters
- end-zone filters
- final-third entry filters
- penalty-box entry filters
- directional pass or carry filters

Do not discard coordinate fields during normalization if the source includes them.
Spatial filtering is a core professional workflow, not an edge case.

### Logging

Always log with useful context.
Never log secrets, auth tokens, or sensitive filesystem paths.

## Testing

Always add or update tests when behavior changes.

Required coverage:

- data auto-reformat logic
- timestamp merging logic
- filter logic
- malformed input handling
- upload size limit handling
- request-scoped SQLite isolation
- concurrency limiting behavior
- staging-table normalization from raw `TEXT` columns
- first-half and second-half time resolution behavior
- `outcomeType` filter behavior
- spatial filter behavior using `x`, `y`, `endX`, and `endY`

Use table-driven tests for filter behavior.

Desired:

```go
func TestFilterEvents(t *testing.T) {
	tests := []struct {
		name     string
		input    []Event
		filter   Filter
		expected []float64
	}{
		{
			name: "player chances created first half",
			input: sampleEvents(),
			filter: Filter{
				Player: "Player A",
				Type:   "Chances Created",
				Half:   1,
			},
			expected: []float64{120.5, 848.2},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := FilterEvents(tc.input, tc.filter)
			if diff := cmp.Diff(tc.expected, got); diff != "" {
				t.Fatalf("unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}
```

Prefer in-memory fixtures and streamed test inputs.
Do not rely on physical disk persistence for SQLite tests unless unavoidable.

Add tests that prove:

- oversized uploads are rejected before SQLite insertion
- two concurrent requests never share the same in-memory database state
- raw staging tables can ingest mixed-type source columns without insert failure
- concurrency caps return a safe overload response instead of exhausting RAM
- half-specific calibration or offset logic does not collapse first half and second half into one naive timeline
- successful versus unsuccessful action filters produce distinct results
- spatial filters correctly isolate tactical zones

## Non-Negotiable Rules

Always preserve these constraints:

- Backend only; no server-side video processing
- Go first; `chi` plus `net/http`
- In-memory SQLite only, isolated per request
- Streaming parsers and bounded memory use
- Hard request size limits before parsing
- Bounded concurrency for in-memory workloads
- ELT staging tables with raw `TEXT` columns when schema is uncertain
- Wrapped errors with safe HTTP responses
- Thin handlers, business logic in services
- Tests for normalization, filtering, and timestamp merging

If a requested change conflicts with these rules, follow these rules.
