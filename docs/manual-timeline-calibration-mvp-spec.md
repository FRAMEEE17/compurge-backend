# Manual Timeline Calibration MVP Spec

## Feature

Manual Timeline Calibration for local football video plus event data sync.

## Objective

Let a user align local match video with imported event data in under 30 seconds so generated highlight timestamps map correctly to the video.

## Primary User

Independent analyst, scout, or small-club staff using a desktop or laptop with:

- a local full-match video file
- event data export from another tool

## Problem

Event timestamps are only useful if the event-data clock matches the video clock. In real workflows, local video often has:

- intro before kickoff
- missing opening seconds
- trimmed halves
- different start point than the exported event data

Without calibration, clips are wrong even if filtering is correct.

## User Value

- No manual clip hunting in editing software
- Faster first-pass comp creation
- Works with local files the user already owns
- No upload required

## Scope

This MVP supports:

- one local video file
- one event data file
- one global time offset per project or session
- manual anchor setting by the user
- instant preview of synced events
- small timing nudges before confirmation

This MVP does not support:

- full automatic sync
- drift correction across long videos
- multi-cam sync
- multiple video files per match
- frame-accurate validation
- separate half offsets

## User Story

As an analyst, I want to set the match start on my local video so that event timestamps line up and I can generate clips quickly.

## Success Criteria

- User can import video and event data and produce a usable sync without reading documentation
- Median time to first confirmed sync is under 30 seconds
- User can preview at least 3 events before confirming
- User can fix small timing errors with simple controls
- Confirmed sync is applied to all downstream clip generation

## Core Workflow

1. Import
- User selects local video file
- User uploads CSV, XLSX, or Parquet event data
- System parses event file and shows a basic summary

2. Sync Prompt
- System asks whether video starts exactly at kickoff
- User chooses:
  - starts at kickoff
  - has intro before kickoff
  - not sure

3. Calibration
- User scrubs the video to kickoff or another recognizable event
- User clicks one of:
  - `Set this moment as 00:00`
  - `Set this moment as selected event time`

4. Preview
- System calculates offset
- System previews several synced events using the current offset
- User reviews whether alignment feels correct

5. Adjust
- User nudges timing with simple buttons or arrow keys
- System updates previews immediately

6. Confirm
- User clicks `Confirm sync`
- System stores offset in project or session state
- All clip timestamps are resolved using this offset

## Functional Requirements

### Import

- Accept local video file
- Accept event data file
- Parse and validate event schema
- Extract at minimum:
  - period
  - minute
  - second
  - event type
  - outcome if available

### Calibration

- Allow user to set a video time as:
  - data `00:00`
  - a selected event timestamp
- Compute offset as:
  - `offset = videoTime - dataTime`
- Support negative and positive offsets

### Preview

- Show current offset after anchor set
- Show 3 to 5 event preview points
- Let the user jump the player to each preview event
- Reflect updated time mapping immediately after adjustments

### Adjustment

- Provide timing controls:
  - `-5s`
  - `-1s`
  - `-0.2s`
  - `+0.2s`
  - `+1s`
  - `+5s`
- Provide keyboard shortcuts:
  - left and right arrow = `-0.2s / +0.2s`
  - shift plus left and right arrow = `-1s / +1s`

### Persistence

- Store confirmed offset in local project or session state
- Reuse offset for highlight generation until the user changes it

## UI Requirements

### Screen Structure

Three-panel layout:

- Left: video player
- Right: event list and sync preview
- Bottom or side control bar: calibration controls

### Required UI Elements

- video player with current video time
- imported event summary
- selected event timestamp
- current calculated sync offset
- anchor actions:
  - `Set this moment as 00:00`
  - `Set this moment as selected event`
- preview event list
- timing nudge controls
- `Confirm sync` button
- sync status:
  - `Not set`
  - `Draft`
  - `Confirmed`

### Copy Rules

Use user-facing language.

Prefer:

- `Set match start`
- `Check alignment`
- `Adjust timing`

Avoid exposing technical language like `offset` as the primary label. It can appear as secondary detail.

### Default Preview Behavior

After the anchor is set, auto-suggest preview checks around:

- an early event
- a mid-early event
- a later event in the same half

Example:

- `00:30`
- `01:45`
- `05:00`

This helps the user detect obvious misalignment quickly.

### Event List Behavior

- Show first events in time order
- Prioritize recognizable events if available:
  - kickoff
  - first pass
  - shot
  - foul
- Clicking an event arms it as the selected anchor target

## State Model

Project or session state should track:

- source video
- source event file
- parsed events
- current draft offset
- confirmed offset
- selected anchor event
- sync confidence state

## Validation and Edge Cases

### Must Handle

- video starts before kickoff
- video starts after kickoff
- no obvious kickoff visual
- event file starts at a non-zero first event
- halftime or long dead time exists in the file
- event rows are valid but sparse

### Out of Scope for MVP

- multiple offsets for first half and second half
- clock drift correction
- automatic whistle detection
- sync from audio analysis
- scene detection

### Error States

- video cannot be decoded locally
- event data cannot be parsed
- no usable timestamp fields found
- user confirms without setting anchor

Each error must return a clear action:

- re-upload file
- choose another column mapping
- select an event manually
- retry with another anchor point

## Performance Requirements

- local parsing summary should appear quickly after import
- setting anchor should update draft sync immediately
- preview jumps should feel near-instant on desktop
- no server upload is required for calibration

## Analytics to Capture

Track:

- time from import to confirmed sync
- whether the user used `Set 00:00` or event-based anchor
- number of nudge adjustments
- whether sync was later edited
- failure points in import or confirmation

## Acceptance Criteria

- User can import one local video and one event file
- User can set one anchor point and generate a draft sync
- User can preview at least 3 mapped events
- User can nudge timing in sub-second and whole-second increments
- User can confirm sync and use it for later clip generation
- Flow works without server-side video processing

## Nice to Have After MVP

- auto-suggest anchor candidates
- kickoff helper overlay
- second anchor for half-time
- drift detection
- whistle or audio-assisted sync
- confidence scoring from multi-point validation
