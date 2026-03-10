# Frontend Implementation Plan

## Goal

Implement a frontend that integrates with the existing backend and supports:

1. local video plus event-data workflow
2. manual timeline calibration
3. backend-driven clip planning
4. sequential frontend-side clip extraction
5. export of separate clips or ZIP output

## Scope

This plan covers:

1. frontend application structure
2. state and API integration
3. timeline calibration flow
4. clip job orchestration
5. OPFS-backed local processing

This plan does not cover:

1. visual design systems
2. animation polish
3. cloud auth
4. multi-user features

## Guiding Principles

1. Keep backend as the source of truth for event filtering and clip-range planning
2. Keep frontend as the source of truth for local video access and local clip processing
3. Treat local full-match video as the primary path
4. Process clips sequentially to protect RAM and browser stability
5. Use OPFS for outputs and temp persistence instead of holding large data in memory

## Recommended Stack

If the frontend is Vite plus React:

1. React for UI state and composition
2. TypeScript for contract safety
3. a small state layer only if needed
4. FFmpeg.wasm for extraction
5. OPFS for browser-local storage
6. standard `fetch` for backend calls

## Delivery Phases

## Phase 1: App Skeleton and Shared Types

### Objective

Create the minimal frontend structure needed to support the workflow.

### Tasks

1. create frontend app shell
2. add typed API models for:
   - timestamp request
   - timestamp response
   - clip range
   - clip job
   - calibration state
3. create folder structure for:
   - api
   - features
   - state
   - processing
   - storage
4. add environment config for backend base URL
5. define API model for preview events

### Output

1. buildable frontend shell
2. typed contracts aligned with backend

### Validation

1. app starts locally
2. type-check passes
3. sample backend payload can be represented without `any`

## Phase 2: Source Setup and Session State

### Objective

Allow the user to choose video and event data, and maintain project session state.

### Tasks

1. implement project session store
2. support local video file selection
3. support remote video URL input as optional best-effort mode
4. support CSV and XLSX file selection
5. persist in-memory session state across view changes
6. show ready state when required inputs exist

### Output

1. working source-setup area
2. session state model wired to file inputs

### Validation

1. local video file can be selected and previewed
2. event file metadata is stored correctly
3. session state updates are deterministic

## Phase 3: Backend Client and Request Builder

### Objective

Create the frontend path that sends filters and files to the backend.

### Tasks

1. implement multipart client for `POST /timestamps`
2. implement JSON client for `POST /timestamps/json`
3. implement preview client for a lightweight event-preview path
4. build request serializer from session state
5. support current backend filter set:
   - period
   - outcome
   - minute bounds
   - spatial filters
   - analytics filters
   - clip timing controls
   - global and half-specific offsets
6. normalize backend errors into frontend-safe messages

### Output

1. reusable API client
2. request-builder utility
3. preview-event client for calibration

### Validation

1. can reproduce successful Postman requests from the frontend
2. request payload matches backend docs
3. response parses into typed models
4. preview-event response is sufficient to build calibration event list

## Phase 4: Timeline Calibration

### Objective

Implement the calibration flow that creates timeline offsets.

### Tasks

1. build video preview panel
2. build event preview list for sync selection from preview-event data, not from clip-range responses
3. implement:
   - set current video time as match start
   - set current video time as selected event time
4. calculate draft offset
5. support nudge controls
6. support confirmation state
7. support:
   - global offset
   - first-half offset
   - second-half offset

### Output

1. working calibration module
2. offsets stored in session state

### Validation

1. user can create and edit offsets without backend dependency
2. confirmed offsets are included in backend requests
3. half-specific offsets override global offset in request payload
4. calibration can start before full clip-planning request is made

## Phase 5: Filter Builder and Clip Planning

### Objective

Let the user define football filters and get clip ranges from the backend.

### Tasks

1. implement filter-state editing
2. map filter state to request payload
3. submit request to backend
4. store backend `clipRanges`
5. present clip plan metadata for confirmation

### Output

1. functioning clip-planning step
2. backend response stored as clip plan

### Validation

1. requests with analytics and spatial filters succeed
2. clip ranges returned by backend appear correctly in frontend state
3. user can rerun planning with new filters without resetting source inputs

## Phase 6: Clip Job Model and Sequential Processor

### Objective

Turn backend clip ranges into executable processing jobs.

### Tasks

1. define clip job model with status lifecycle
2. build job queue from `clipRanges`
3. implement sequential job runner
4. enforce one active clip-processing job at a time
5. support cancel and retry

### Output

1. deterministic job queue
2. stable sequential processing pipeline

### Validation

1. jobs move through `queued -> processing -> completed/failed`
2. only one job processes at a time
3. failures remain localized to affected jobs

## Phase 7: OPFS Storage Layer

### Objective

Add durable browser-local storage for outputs and temp artifacts.

### Tasks

1. create OPFS wrapper utilities
2. support write, read, delete, and list operations
3. create naming strategy for generated clips
4. store completed clip outputs in OPFS
5. create cleanup logic for temp artifacts
6. keep exported and completed clips separate from disposable temp files

### Output

1. storage adapter for clip outputs
2. cleanup path for old temp files

### Validation

1. completed outputs persist in OPFS during session
2. temp files are removed after each job
3. storage paths remain deterministic and collision-free

## Phase 8: FFmpeg.wasm Processing Integration

### Objective

Integrate the actual extraction engine.

### Tasks

1. initialize FFmpeg.wasm lazily
2. use an FFmpeg.wasm version and mount strategy that can read large local file objects without first copying the full source into MEMFS
3. prepare per-job input handling
4. run extraction commands for current clip only
5. move finished output to OPFS
6. clear FFmpeg temp files after each job
7. avoid retaining batch outputs in FFmpeg memory

### Output

1. end-to-end clip extraction path

### Validation

1. a single clip can be extracted from local source video
2. multiple clips can be processed sequentially without memory growth spikes
3. output files appear in OPFS after each completed job
4. large local source files are not copied wholesale into WASM memory

## Phase 9: Export Modes

### Objective

Enable the user to retrieve generated outputs.

### Tasks

1. implement per-clip download
2. implement streaming ZIP export of all completed clips from OPFS
3. if supported, evaluate direct directory-save export as an alternative bulk-export path
4. optionally prepare concat mode as deferred feature flag

### Output

1. useful export path for analyst workflow

### Validation

1. user can download one clip
2. user can download a ZIP of all successful clips without loading the full archive into RAM
3. ZIP export works even if some jobs failed

## Phase 10: Reliability and Guardrails

### Objective

Add production-minded protections around the browser workflow.

### Tasks

1. add capability checks for:
   - local file presence
   - backend reachability
   - optional remote video compatibility
2. add processing-estimate warnings for large clip batches
3. add storage-usage warnings where possible
4. add safe handling for user cancellation
5. add recovery path for failed jobs
6. add `beforeunload` protection while processing is active

### Output

1. safer browser processing flow

### Validation

1. failure modes are surfaced clearly
2. a single failure does not corrupt the whole session
3. user can retry failed jobs without rebuilding everything
4. accidental refresh or tab-close attempts warn while processing is active

## Data Contracts

The frontend must align with backend request fields:

1. `period`
2. `eventType`
3. `outcomeType`
4. `minMinute`
5. `maxMinute`
6. `minX`
7. `maxX`
8. `minY`
9. `maxY`
10. `minEndX`
11. `maxEndX`
12. `minEndY`
13. `maxEndY`
14. `minXT`
15. `maxXT`
16. `minProgPass`
17. `maxProgPass`
18. `minProgCarry`
19. `maxProgCarry`
20. `preRollSeconds`
21. `postRollSeconds`
22. `mergeGapSeconds`
23. `timelineOffset`
24. `firstHalfOffset`
25. `secondHalfOffset`

## Technical Constraints

1. The frontend must not replicate backend filtering logic.
2. The frontend must not upload local full-match video to the backend for clipping.
3. The frontend must not process clip jobs in parallel for MVP.
4. The frontend must not keep the full source plus all outputs in memory.
5. The frontend must treat remote video as optional and unreliable.

## Suggested Execution Order

If time is limited, implement in this exact order:

1. app shell and typed models
2. source setup and session state
3. backend client and request builder
4. preview-event path for calibration
5. timeline calibration
6. filter builder and clip planning
7. clip job queue
8. OPFS layer
9. FFmpeg integration
10. ZIP export

## Definition of Done

This frontend plan is complete when:

1. a user can select local video and event data
2. a user can calibrate timeline offsets
3. a user can call the backend and receive clip ranges
4. the frontend can process each clip sequentially
5. outputs are stored in OPFS
6. the user can export clips individually or as a ZIP
