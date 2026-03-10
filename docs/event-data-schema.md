# Event Data Schema

## Purpose

This document describes the event-data shapes the system can ingest, normalize, and filter.

Use this file for source-specific schema notes and normalization rules.
Do not put detailed schema reference material in `agent.md`.

## Design Rule

All external event data must be normalized into a minimal internal event model before filtering or timestamp generation.

The system must not depend on one vendor-specific schema as the universal source shape.

## Minimal Internal Event Model

Every provider-specific import should map to this core shape where possible:

- `period`
- `minute`
- `second`
- `eventType`
- `outcomeType`
- `playerId` or `playerName`
- `teamId` or `teamName`
- `x`
- `y`
- `endX`
- `endY`
- `metadata`

Notes:

- `metadata` is optional and should only hold source-specific values that are not part of the core filter model.
- `minute` and `second` may need to be converted into a single absolute event time for filtering and clip generation.
- `period` must be preserved because first-half and second-half time handling is not always equivalent across providers.

## Required for Timestamp Generation

At minimum, the system needs enough information to resolve an event time:

- `minute`
- `second`
- `period`

If one of these is missing, the parser must either:

- derive it safely from other fields
- reject the file with a clear validation error

## Required for Basic Filtering

At minimum, basic event filtering should support:

- `eventType`
- `period`

Useful but optional:

- `outcomeType`
- `playerName`
- `playerId`
- `teamName`
- `teamId`

## Source-Specific Notes

Add a section per source as needed.

Suggested pattern:

### Source: `<provider or export name>`

- File type: CSV, XLSX, or Parquet
- Typical use: player export, team export, match export
- Time fields:
- Event-type field:
- Outcome field:
- Coordinate fields:
- Known quirks:
- Example row:
- Mapping to internal model:

## Example: Local Player Export

Example file:

- [`2026-03-09T17-03_export.csv`](/Users/palmer/compurge/2026-03-09T17-03_export.csv)

Observed columns:

- `minute`
- `second`
- `period`
- `type`
- `outcomeType`
- `prog_pass`
- `prog_carry`
- `xT`
- `x`
- `y`
- `endX`
- `endY`

Interpretation:

- `minute`, `second`, and `period` are sufficient for event-time resolution
- `type` maps to `eventType`
- `outcomeType` can be used as a filter field
- `x`, `y`, `endX`, and `endY` map naturally to event location fields
- `prog_pass`, `prog_carry`, and `xT` should be treated as optional analytics fields

Possible normalized mapping:

- `period` -> `period`
- `minute` -> `minute`
- `second` -> `second`
- `type` -> `eventType`
- `outcomeType` -> `outcomeType`
- `x` -> `x`
- `y` -> `y`
- `endX` -> `endX`
- `endY` -> `endY`
- `prog_pass` -> `metadata.progPass`
- `prog_carry` -> `metadata.progCarry`
- `xT` -> `metadata.xT`

## Normalization Rules

Always apply these rules:

- trim whitespace from headers before matching
- treat header matching as case-insensitive when safe
- preserve original source headers in mapping metadata when useful
- parse unknown schemas into raw staging tables before transformation
- keep raw values as text first when source typing is inconsistent

Current header alias coverage includes:

- `period`, `half`
- `minute`, `matchMinute`, `match_minute`, `matchMin`, `min`
- `second`, `seconds`, `matchSecond`, `match_second`, `matchSec`, `sec`
- `type`, `eventType`, `event`, `actionType`, `primaryEvent`
- `outcomeType`, `outcome`, `result`, `eventOutcome`
- `playerId`, `player`, `playerIdentifier`
- `playerName`, `playerFullName`, `athleteName`
- `teamId`, `teamIdentifier`, `squadId`
- `teamName`, `team`, `squadName`
- `x`, `startX`
- `y`, `startY`
- `endX`, `to_x`, `destinationX`
- `endY`, `to_y`, `destinationY`
- `xT`, `expectedThreat`

Do not:

- hardcode assumptions that all files come from one provider
- drop `period` if the source includes it
- assume `second` is always an integer
- assume coordinates are always present

## Time Resolution Rules

Preferred approach:

1. Parse `period`
2. Parse `minute`
3. Parse `second`
4. Convert to normalized event time only after validation

If a source uses alternate time formats, document the conversion rule here.

Examples:

- `minute=8`, `second=10.5`, `period=FirstHalf`
- `minute=45`, `second=0`, `period=SecondHalf`

## Validation Rules

Reject or flag rows when:

- time fields are missing and cannot be derived
- `period` is missing and half-specific filtering is required
- row shape does not match the declared header count
- required columns for a chosen filter are absent

Do not fail the whole file for optional analytics fields unless the requested workflow depends on them.

## Filtering Notes

The normalized model should allow filters such as:

- player
- event type
- outcome
- period
- coordinate-based region filters
- analytics-threshold filters such as `xT > N`

Only promote a field into the core normalized model if it is broadly useful across multiple sources.
Keep source-specific or experimental fields in `metadata` until they are proven stable.

## Maintenance Rule

When a new provider or export format is added:

1. Add a source-specific section to this file
2. Document mapping quirks
3. Add parser tests
4. Add normalization tests
5. Do not expand `agent.md` with provider-specific detail
