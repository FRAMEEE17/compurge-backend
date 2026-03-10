# Business Logic Initial Plan

## Goal

Create the first business-logic implementation for event parsing, normalization, filtering, and timestamp merging without adding HTTP handlers or frontend code.

## Scope

What is included:

- Go module setup
- domain models for events, filters, and clip ranges
- streaming CSV parser
- event normalization into internal structs
- filtering logic
- timestamp and clip-range resolution
- merge and deduplication logic
- unit tests

What is not included:

- HTTP handlers
- SQLite ingestion
- XLSX or Parquet parsing
- frontend clip processing
- timeline calibration UI

## Assumptions

- CSV is the first supported input format
- the sample export in the repo is representative enough for an initial parser
- business logic should be isolated from transport and persistence concerns

## Risks

- source schemas will vary more than the first CSV sample suggests
- half and time semantics may differ across providers
- merge behavior may need tuning once real users review clip outputs

## Plan

### Step 1

Change:

- initialize a Go module
- add internal domain packages and types

Validation:

- `go test ./...`

### Step 2

Change:

- implement streaming CSV parsing and normalization
- implement filter evaluation and time resolution

Validation:

- parser and filter unit tests

### Step 3

Change:

- implement clip-range building and merge logic
- add tests using the sample CSV and table-driven cases

Validation:

- `go test ./...`

## Completion Criteria

- core business logic is implemented without HTTP dependencies
- CSV events can be parsed and filtered into clip ranges
- merge logic is covered by tests
