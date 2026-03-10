# Frontend Spec

## Purpose

This document defines the frontend scope for the football highlight workflow.
It focuses on functional requirements, system behavior, state, API usage, and implementation constraints.
It does not define visual design.

The frontend is responsible for:

1. Collecting local video and event data from the user
2. Helping the user calibrate match time against video time
3. Sending filters and calibration metadata to the backend
4. Receiving clip ranges from the backend
5. Preparing clip jobs for frontend-side video extraction
6. Managing local processing, temporary storage, and export

The frontend is not responsible for:

1. Recomputing football event filtering logic already implemented in the backend
2. Trusting arbitrary remote video URLs to work in the browser
3. Assuming frame-accurate output by default
4. Loading an entire full-match video into memory at once

## Primary User

The primary user is a desktop-first analyst, scout, or small-club staff member who:

1. Has a local full-match video file or a usable remote video source
2. Has event data in CSV or XLSX form
3. Needs clips quickly
4. Accepts approximate cuts when speed is more important than exact frame precision

## User Needs

The frontend must satisfy these user needs:

1. Import a local full-match video without requiring upload to a backend
2. Import event data quickly and inspect what was loaded
3. Calibrate video time to event-data time without understanding timecode internals
4. Apply football-specific filters such as event type, outcome, xT, progressive pass, progressive carry, and spatial filters
5. Generate a clip list quickly from the backend
6. Preview what will be clipped before processing
7. Export either separate clips or a ZIP of clips
8. Optionally merge clips into a single compilation later
9. Keep local video processing stable on real machines without browser crashes

## Supported Input Modes

### Video Input

The frontend must support these video-source modes:

1. Local full-match file
   - Primary supported path
   - Best reliability
   - No CORS dependency
   - Best fit for desktop-first MVP
2. Remote full-match URL
   - Secondary, best-effort path
   - Must be treated as unreliable unless CORS and range support are confirmed

### Event Data Input

The frontend must support:

1. CSV file upload
2. XLSX file upload
3. Optional future direct JSON input for internal tools only

## Core Frontend Modules

The frontend should be split into these logical modules:

1. Project session module
   - owns current video source, event file, filters, offsets, clip jobs, and export state
2. Event data import module
   - validates file type
   - shows import status
   - sends file or JSON payload to backend
3. Timeline calibration module
   - manages draft and confirmed offsets
   - supports both global and half-specific offsets
4. Filter builder module
   - owns user-selected filter values
   - serializes them into backend request format
5. Backend client module
   - calls `/timestamps` or `/timestamps/json`
   - parses responses and errors
6. Clip job module
   - converts backend clip ranges into frontend extraction jobs
7. Video processing module
   - orchestrates FFmpeg.wasm work sequentially
   - uses OPFS for temporary and output files
8. Export module
   - downloads separate clips, ZIP archives, or optional merged output
9. Event preview module
   - provides lightweight event rows for calibration and preview
   - must not duplicate backend filtering logic

## Required Screens or Functional Areas

The frontend must provide these functional areas:

1. Source setup
   - choose local video or remote video mode
   - upload event data
   - show detected file names and ready state
2. Calibration
   - preview video
   - set `00:00` anchor or selected-event anchor
   - adjust offsets
   - confirm sync
3. Filtering
   - choose period, event type, outcome, time, analytics, and spatial filters
4. Clip planning
   - request clip ranges from backend
   - display returned clip ranges before processing
5. Processing
   - create clip jobs
   - process sequentially
   - show progress and failures per clip
6. Export
   - save individual clips
   - save ZIP of clips
   - optional merged compilation

## Functional Requirements

### 1. Project Session Management

The frontend must maintain one active project session with:

1. video source metadata
2. event data file metadata
3. current filter state
4. draft offsets
5. confirmed offsets
6. backend clip ranges
7. generated clip job list
8. processing progress
9. export artifacts

The session state must survive ordinary UI transitions.
Persistent storage across reloads is optional for MVP.

### 2. Video Source Handling

#### Local Full-Match Video

The frontend must treat local full-match video as the primary supported path.

Requirements:

1. Accept large local video files without uploading them to the backend
2. Use browser-native file references where possible
3. Avoid reading the whole file into JS memory at once
4. Use local playback for calibration and preview

#### Remote Full-Match Video

The frontend may support remote video as a best-effort path only.

Requirements:

1. Validate URL format before use
2. Detect whether remote playback is possible
3. Treat remote processing as unsupported if CORS or range support is missing
4. Explain that remote video compatibility depends on the source server

The frontend must not promise that all remote video URLs will work.

### 3. Event Data Upload and Backend Requesting

The frontend must send event data to the backend in one of two ways:

1. `POST /timestamps`
   - multipart upload
   - preferred for file upload flows
2. `POST /timestamps/json`
   - JSON body
   - preferred for advanced or internal flows

For MVP, the frontend should default to `POST /timestamps`.

The frontend also needs a lightweight event-preview source for calibration.
Because the calibration flow needs event rows before full clip planning, the frontend must support one of these two patterns:

1. preferred pattern: a dedicated backend preview endpoint such as `POST /parse-preview`
   - accepts the same event file
   - returns a small parsed preview such as the first 50 normalized events
   - keeps parsing logic centralized in the backend
2. fallback pattern: local head-only parsing in the browser
   - only for preview purposes
   - limited to a small number of rows
   - must not become a second filtering engine in the frontend

For this project, the preferred direction is a backend preview endpoint.

The frontend must allow users to:

1. upload CSV or XLSX
2. resubmit requests with different filters
3. reuse the same event file across multiple requests

### 4. Timeline Calibration

The frontend must support:

1. global `timelineOffset`
2. `firstHalfOffset`
3. `secondHalfOffset`

The frontend must expose calibration in user language, but internally map to these backend fields.

Rules:

1. If the user sets only one offset, send `timelineOffset`
2. If the user calibrates each half separately, send `firstHalfOffset` and `secondHalfOffset`
3. If half-specific offsets are set, they must override global offset in outgoing requests

The frontend must support:

1. draft offset changes before confirmation
2. event preview jumps for sync verification
3. small nudge adjustments
4. recalibration after clips are generated
5. calibration based on lightweight preview events returned from a preview source

Calibration must not depend on the frontend first generating clip ranges from the backend.
The frontend must be able to show a small event preview before full planning.

### 5. Filter Support

The frontend must support all current backend filters.

#### Basic Filters

1. `period`
2. `eventType`
3. `outcomeType`
4. `minMinute`
5. `maxMinute`

#### Spatial Filters

1. `minX`
2. `maxX`
3. `minY`
4. `maxY`
5. `minEndX`
6. `maxEndX`
7. `minEndY`
8. `maxEndY`

#### Analytics Filters

1. `minXT`
2. `maxXT`
3. `minProgPass`
4. `maxProgPass`
5. `minProgCarry`
6. `maxProgCarry`

#### Clip Timing Controls

1. `preRollSeconds`
2. `postRollSeconds`
3. `mergeGapSeconds`

The frontend must only send fields that the user has set or that have defined defaults.

### 6. Backend Response Handling

The frontend must consume:

1. `requestId`
2. `clipRanges[]`

Each clip range must be treated as a planned extraction job.

The frontend must display or internally store:

1. `eventTime`
2. `resolvedEventTime`
3. `resolvedStartTime`
4. `resolvedEndTime`
5. `sourceEventType`
6. `sourceOutcomeType`

### 7. Clip Job Generation

The frontend must convert each returned clip range into a clip job with:

1. clip identifier
2. source video reference
3. start time
4. end time
5. duration
6. event metadata
7. processing status
8. output file target

The frontend must support:

1. queued
2. processing
3. completed
4. failed
5. canceled

### 8. Frontend Video Processing

The frontend must process clips sequentially, not in parallel, for MVP.

Reasons:

1. lower memory pressure
2. lower risk of browser crashes
3. easier progress accounting
4. more predictable CPU and disk behavior

The frontend must use:

1. FFmpeg.wasm for extraction
2. OPFS as disk-backed staging and output storage
3. a file-mount strategy that avoids copying large full-match files into FFmpeg memory

The frontend must target an FFmpeg.wasm integration that supports mounting browser file objects instead of forcing large full-match files through in-memory `writeFile` flows.
Do not design the pipeline around copying a multi-gigabyte local video file into MEMFS.

The frontend must not:

1. hold the full-match source plus many generated clips in JS memory at once
2. process many clip jobs simultaneously
3. keep temp files in FFmpeg virtual FS longer than needed

### 9. Chunking and Memory Rules

Chunking must happen at the clip-job level.

That means:

1. backend returns a list of highlight ranges
2. frontend turns each range into one extraction job
3. frontend processes one job at a time
4. after each job:
   - move output to OPFS
   - clear FFmpeg temp files
   - release temporary references

For remote sources, chunking may also involve range-based retrieval, but that is secondary and best-effort.

#### Where OPFS Must Be Used

OPFS must be used for:

1. generated clip outputs
2. temporary clip outputs before export
3. optional concat input lists and intermediate artifacts

OPFS should be treated as the durable local work area during a session.
The frontend should prefer OPFS over RAM-backed temporary accumulation whenever outputs are larger than trivial preview artifacts.

#### Where FFmpeg.wasm Should Hold Memory

FFmpeg.wasm should hold only:

1. the current working input slice or active extraction input
2. the current in-progress output clip
3. small command-side temporary artifacts

FFmpeg.wasm should not hold:

1. the entire full-match video file in memory
2. all clips from a batch at once
3. long-lived outputs that should already be moved to OPFS

### 10. Export Modes

The frontend must support these export modes:

1. download separate clips
2. download ZIP of clips

The frontend may later support:

1. merged compilation output

For MVP, ZIP of clips is the preferred bulk export mode only if it is implemented as a streaming ZIP pipeline.

Rules:

1. do not build ZIP exports by loading all clips into RAM first
2. do not use ZIP implementations that accumulate the entire archive in memory before download
3. build ZIP exports by streaming clip data from OPFS into the archive
4. prefer direct per-file export or directory-save flows when browser capabilities allow them

If the environment supports a directory-save flow, saving clips directly to a user-selected directory is an acceptable or preferable alternative to ZIP.

Reasons:

1. lower failure risk than concat
2. often closer to real analyst workflow
3. easier to keep bounded in memory if implemented as a streaming export

### 11. Error Handling

The frontend must handle these failure categories:

1. invalid event file
2. backend request validation failure
3. remote video incompatibility
4. calibration not confirmed
5. FFmpeg processing failure on an individual clip
6. OPFS write failure or quota issue
7. user cancellation

The frontend must treat clip-processing failures as per-job failures where possible, not session-fatal failures.

### 12. Browser and Device Strategy

The frontend must treat desktop browser support as the primary target.

MVP assumptions:

1. desktop and laptop first
2. mobile is best-effort only
3. low-end devices may be blocked or warned for heavy processing

The frontend must not promise equal performance across all devices.

### 13. Navigation and Session Safety

The frontend must protect users from accidental loss of long-running local processing work.

Requirements:

1. if clip processing is active, register a `beforeunload` guard
2. warn before refresh, tab close, or navigation away while processing is active
3. remove the warning when processing is idle
4. keep completed outputs already written to OPFS even if the page is later closed

## State Model

The frontend state should include:

1. `videoSource`
2. `eventDataFile`
3. `sourceMode`
4. `calibration`
5. `filters`
6. `backendRequest`
7. `backendResponse`
8. `clipJobs`
9. `processing`
10. `exports`
11. `errors`

### Minimum Session Shape

```ts
type ProjectSession = {
  sourceMode: "local" | "remote";
  videoSource: {
    fileName?: string;
    url?: string;
    mimeType?: string;
  } | null;
  eventDataFile: {
    fileName: string;
    kind: "csv" | "xlsx";
  } | null;
  calibration: {
    timelineOffset?: number;
    firstHalfOffset?: number;
    secondHalfOffset?: number;
    confirmed: boolean;
  };
  filters: {
    period?: string;
    eventType?: string;
    outcomeType?: string;
    minMinute?: number;
    maxMinute?: number;
    minX?: number;
    maxX?: number;
    minY?: number;
    maxY?: number;
    minEndX?: number;
    maxEndX?: number;
    minEndY?: number;
    maxEndY?: number;
    minXT?: number;
    maxXT?: number;
    minProgPass?: number;
    maxProgPass?: number;
    minProgCarry?: number;
    maxProgCarry?: number;
    preRollSeconds?: number;
    postRollSeconds?: number;
    mergeGapSeconds?: number;
  };
  clipJobs: ClipJob[];
};
```

## API Integration Rules

The frontend must follow these rules:

1. Use multipart requests for ordinary file-upload flows
2. Use JSON endpoint only when the frontend already has structured payloads or preprocessed data
3. Reuse backend filtering logic instead of duplicating filter evaluation in the frontend
4. Treat backend clip ranges as source of truth for planned extraction

## Non-Goals

The frontend MVP does not need:

1. automatic event-data parsing in the browser
2. automatic multi-provider schema mapping
3. cloud rendering
4. multi-user collaboration
5. full project persistence backend
6. server-side clip generation
7. frame-accurate rendering by default

## Acceptance Criteria

The frontend is MVP-ready when a user can:

1. load a local full-match video
2. upload CSV or XLSX event data
3. calibrate match time to video time
4. apply football-specific filters
5. request clip ranges from the backend
6. process returned clip ranges sequentially
7. export separate clips or a ZIP of clips
