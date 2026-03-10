# Naming Conventions

## Purpose

This document defines naming rules for code, files, tests, documentation, and sample data in this repository.

Use this file together with [`agent.md`](/Users/palmer/compurge/agent.md), [`docs/golang/go-practices.md`](/Users/palmer/compurge/docs/golang/go-practices.md), and [`docs/golang/go-checklist.md`](/Users/palmer/compurge/docs/golang/go-checklist.md).

## General Rules

Always choose names that are:

- descriptive
- specific
- stable
- searchable
- consistent with surrounding code

Good names should let a new contributor infer purpose quickly.

Avoid names that are:

- vague
- overloaded
- temporary-sounding
- dependent on private context that is not visible in the file

Avoid:

- `tmp`
- `misc`
- `helper`
- `util`
- `data`
- `thing`
- `stuff`
- `foo` and `bar` in production code
- `final`
- `new`
- `updated`

Use short names only when the scope is tiny and obvious, such as `i` in a small loop.

## Go Package Names

Always:

- use short lowercase package names
- use a single word when practical
- make the package name describe its responsibility

Prefer:

- `handler`
- `service`
- `repository`
- `parser`
- `model`

Avoid:

- underscores in package names
- mixedCaps package names
- generic package names like `common`, `utils`, or `helpers`
- package names that repeat the parent directory meaninglessly

## Go File Names

Go source files should be lowercase and descriptive.

Prefer:

- `timestamp_service.go`
- `event_parser.go`
- `sqlite_repository.go`
- `filter_test.go`

Avoid:

- `utils.go`
- `helpers.go`
- `misc.go`
- `temp.go`
- `new_code.go`

If a file holds one dominant concept, name the file after that concept.
If a file exists for one adapter, include the adapter in the filename.

## Types

Use `PascalCase` for exported types and clear domain names for internal types.

Prefer:

- `TimestampRequest`
- `TimestampResponse`
- `EventFilter`
- `TimelineCalibration`

Avoid:

- `DataStruct`
- `ManagerThing`
- `ProcessorHelper`

Type names should express domain meaning, not implementation vagueness.

## Functions and Methods

Use idiomatic Go `camelCase` for unexported names and `PascalCase` for exported ones.

Prefer verb-oriented names for functions that do work:

- `parseEvents`
- `fetchRemoteFile`
- `mergeTimestamps`
- `validateRequest`

Prefer noun or noun-phrase names for constructors and accessors only when appropriate:

- `NewService`
- `SourceRef`

Avoid names that hide side effects:

- `handle`
- `process`
- `run`

Use broader verbs like these only when the surrounding context already makes the action precise.

## Variables

Variable names should reflect purpose, not just data shape.

Prefer:

- `requestID`
- `eventTime`
- `clipRanges`
- `rawRow`
- `normalizedEvent`
- `responseBody`

Avoid:

- `x`
- `y`
- `data`
- `obj`
- `res`
- `val`

Exceptions:

- short loop variables
- math-heavy local logic
- conventional receiver names when the type is obvious

## Constants

Use descriptive names for constants.

Prefer:

- `DefaultUploadLimit`
- `MaxConcurrentIngestJobs`
- `DefaultHTTPTimeout`

Avoid magic numbers embedded in code when the value has policy meaning.

Use all-caps only when it is already an established convention or interoperating with external constant styles.
Prefer Go-style mixed-case constants by default.

## Interfaces

Name interfaces by behavior, not by implementation.

Prefer:

- `EventSource`
- `CustomerStorer`
- `TimestampRepository`

Avoid:

- `IEventSource`
- `DataInterface`
- `BaseRepository`

Do not create generic interface names when the interface only exists to mirror a concrete type.

## Acronyms

Follow Go acronym conventions consistently.

Prefer:

- `ID`
- `URL`
- `HTTP`
- `JSON`
- `SQL`
- `API`

Examples:

- `requestID`
- `videoURL`
- `httpClient`
- `jsonBody`

Avoid inconsistent variants like:

- `Id`
- `Url`
- `HttpClient`

## Tests and Benchmarks

Test names should explain the behavior under test.

Prefer:

- `TestFilterEvents_FirstHalfPasses`
- `TestMergeTimestamps_DeduplicatesOverlaps`
- `TestParseCSV_RejectsMissingPeriod`

Benchmark names should explain the subject and scenario:

- `BenchmarkParseCSV_Streamed`
- `BenchmarkMergeTimestamps_LargeInput`

Avoid opaque names such as:

- `Test1`
- `TestBasic`
- `BenchmarkThing`

## Docs and Specs

Use `snake_case` for documentation filenames in this repository.

Prefer:

- `manual_timeline_calibration_mvp_spec.md`
- `event_data_schema.md`
- `go_practices.md`

If a document is a spec, include `spec` in the filename.
If a document is architecture-focused, include `architecture` or a similarly precise term.
If chronology matters, use `YYYY-MM-DD`.

Avoid:

- `final.md`
- `notes_new.md`
- `updated_spec_latest.md`

## Sample Data and Exports

Sample data files should make source and purpose visible in the filename.

Prefer:

- `wyscout_player_export_2026-03-09.csv`
- `statsbomb_match_events_sample.json`
- `local_club_training_session_tags.xlsx`

Include a date only when it matters for versioning or provenance.
Prefer ISO date format: `YYYY-MM-DD`.

Avoid ambiguous filenames like:

- `export.csv`
- `data.xlsx`
- `sample_final.csv`

## Versioning in Names

Do not use ad hoc suffixes such as:

- `final`
- `final_v2`
- `really_final`

If versioning is necessary in filenames, use a consistent scheme such as:

- `v1`
- `v2`

If the file is time-based, prefer a date over a vague version suffix.

## Repository-Specific Guidance

For this repository in particular:

- make event and timestamp terminology explicit
- distinguish raw, normalized, and resolved representations in names
- distinguish local versus remote video sources in names
- distinguish clip ranges, event times, and resolved video times in names
- include `offset` or `calibration` explicitly when time alignment is involved

Prefer:

- `rawStagingRow`
- `normalizedEvent`
- `resolvedVideoTime`
- `timelineOffset`
- `localVideoSource`
- `remoteVideoRef`

Avoid names that blur domain boundaries:

- `videoData`
- `eventData` when the code really means parsed rows, filtered events, or clip ranges

## Decision Rule

When choosing between two names, prefer the one that:

1. makes search easier
2. explains domain meaning more clearly
3. stays accurate as the code evolves
4. matches Go idioms and nearby code
