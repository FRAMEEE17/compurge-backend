# Timestamps API

## Purpose

This document describes the current backend API for generating timestamped clip ranges from event data.

The backend does not render video.
It parses event data, applies filters, resolves timeline offsets, and returns clip ranges for the frontend.

## Endpoints

`POST /timestamps`

Additional endpoint:

`POST /timestamps/json`

Preview endpoints:

- `POST /parse-preview`
- `POST /parse-preview/json`

## Content Type

`multipart/form-data`

For the JSON endpoints:

`application/json`

## Required Form Fields

- `eventData`: event data file upload (`.csv` or `.xlsx`)

## Optional Form Fields

- `requestId`: client-provided request identifier
- `period`: event period filter
- `eventType`: event type filter
- `outcomeType`: outcome filter
- `minMinute`: lower minute bound
- `maxMinute`: upper minute bound
- `minX`: lower X-coordinate bound
- `maxX`: upper X-coordinate bound
- `minY`: lower Y-coordinate bound
- `maxY`: upper Y-coordinate bound
- `minEndX`: lower ending X-coordinate bound
- `maxEndX`: upper ending X-coordinate bound
- `minEndY`: lower ending Y-coordinate bound
- `maxEndY`: upper ending Y-coordinate bound
- `minXT`: lower xT threshold
- `maxXT`: upper xT threshold
- `minProgPass`: lower progressive pass threshold
- `maxProgPass`: upper progressive pass threshold
- `minProgCarry`: lower progressive carry threshold
- `maxProgCarry`: upper progressive carry threshold
- `preRollSeconds`: seconds to include before each event
- `postRollSeconds`: seconds to include after each event
- `timelineOffset`: seconds to shift event time onto video time
- `firstHalfOffset`: half-specific override for first-half event resolution
- `secondHalfOffset`: half-specific override for second-half event resolution
- `mergeGapSeconds`: maximum gap for merging adjacent clip ranges

## Supported Period Values

- `FirstHalf`
- `SecondHalf`
- `1`
- `2`
- `H1`
- `H2`

## Response

### Success

Status:

- `200 OK`

Body:

```json
{
  "requestId": "req_1741600000000000000_1",
  "clipRanges": [
    {
      "eventTime": 24,
      "resolvedEventTime": 29,
      "resolvedStartTime": 27,
      "resolvedEndTime": 32,
      "sourceEventType": "Pass",
      "sourceOutcomeType": "Successful"
    }
  ]
}
```

Notes:

- `eventTime` is the event time on the normalized match timeline
- `resolvedEventTime` is the event time after applying `timelineOffset`
- `resolvedStartTime` and `resolvedEndTime` are the clip boundaries to use on the video timeline

### Error

Possible status codes:

- `400 Bad Request` for invalid multipart requests, missing file uploads, bad filter values, or oversized uploads
- `429 Too Many Requests` when concurrent ingest capacity is full
- `500 Internal Server Error` when backend processing fails

## Preview API

The preview API is intended for timeline calibration.
It parses the event file and returns a small normalized event list without generating clip ranges.

### Multipart Preview Endpoint

`POST /parse-preview`

Required form fields:

- `eventData`: event data file upload (`.csv` or `.xlsx`)

Optional form fields:

- `requestId`
- `limit`

Rules:

- default preview limit is `50`
- maximum preview limit is `100`

### JSON Preview Endpoint

`POST /parse-preview/json`

The request must include either:

- `eventDataCsv`
- `events`

Optional JSON fields:

- `requestId`
- `limit`

### Preview Response

```json
{
  "requestId": "req_1741600000000000000_2",
  "events": [
    {
      "minute": 0,
      "second": 24,
      "period": "FirstHalf",
      "matchSecond": 24,
      "playerName": "Player A",
      "teamName": "Team A",
      "eventType": "Pass",
      "outcomeType": "Successful",
      "xT": 0.03
    }
  ]
}
```

Notes:

- `matchSecond` is the normalized match-timeline second for calibration use
- preview endpoints do not build clip ranges
- preview endpoints are intended for event-list display before full planning

## Behavior

The current API flow is:

1. enforce request body size limit
2. parse multipart form
3. stream CSV or XLSX rows into request-scoped in-memory SQLite staging
4. normalize events into the internal event model
5. apply filters
6. build clip ranges with pre-roll, post-roll, and timeline offset
7. merge adjacent clip ranges when allowed
8. return JSON

The JSON endpoint reuses the same business logic but accepts either:

- `eventDataCsv` as raw CSV text
- `events` as already structured event objects

The preview endpoints reuse the same parsing and normalization path but stop before clip-range generation.

## Example Request

```bash
curl -X POST http://localhost:8080/timestamps \
  -F "eventData=@2026-03-09T17-03_export.csv" \
  -F "period=FirstHalf" \
  -F "eventType=Pass" \
  -F "outcomeType=Successful" \
  -F "preRollSeconds=2" \
  -F "postRollSeconds=3" \
  -F "timelineOffset=5" \
  -F "mergeGapSeconds=0"
```

## Example Response

```json
{
  "requestId": "req_1741600000000000000_1",
  "clipRanges": [
    {
      "eventTime": 24,
      "resolvedEventTime": 29,
      "resolvedStartTime": 27,
      "resolvedEndTime": 32,
      "sourceEventType": "Pass",
      "sourceOutcomeType": "Successful"
    },
    {
      "eventTime": 105,
      "resolvedEventTime": 110,
      "resolvedStartTime": 108,
      "resolvedEndTime": 113,
      "sourceEventType": "Pass",
      "sourceOutcomeType": "Successful"
    }
  ]
}
```

## Example Preview Request

```bash
curl -X POST http://localhost:8080/parse-preview \
  -F "eventData=@2026-03-09T17-03_export.csv" \
  -F "limit=5"
```

## Example Preview Response

```json
{
  "requestId": "req_1741600000000000000_2",
  "events": [
    {
      "minute": 0,
      "second": 24,
      "period": "FirstHalf",
      "matchSecond": 24,
      "eventType": "Pass",
      "outcomeType": "Successful"
    },
    {
      "minute": 1,
      "second": 45,
      "period": "FirstHalf",
      "matchSecond": 105,
      "eventType": "Carry",
      "outcomeType": "Successful"
    }
  ]
}
```

## Example: Minute and Coordinate Filter

```bash
curl -X POST http://localhost:8080/timestamps \
  -F "eventData=@2026-03-09T17-03_export.csv" \
  -F "period=FirstHalf" \
  -F "eventType=Pass" \
  -F "minMinute=0" \
  -F "maxMinute=10" \
  -F "minX=25" \
  -F "maxX=40" \
  -F "preRollSeconds=1.5" \
  -F "postRollSeconds=2.5"
```

## Example: Analytics and Tactical Filter

```bash
curl -X POST http://localhost:8080/timestamps \
  -F "eventData=@2026-03-09T17-03_export.csv" \
  -F "period=FirstHalf" \
  -F "eventType=Pass" \
  -F "outcomeType=Successful" \
  -F "minXT=0.01" \
  -F "minProgPass=10" \
  -F "minEndX=80" \
  -F "minEndY=18" \
  -F "maxEndY=62" \
  -F "firstHalfOffset=5" \
  -F "secondHalfOffset=120"
```

## Example: Client-Provided Request ID

```bash
curl -X POST http://localhost:8080/timestamps \
  -F "requestId=player_a_first_half_passes" \
  -F "eventData=@2026-03-09T17-03_export.csv" \
  -F "period=FirstHalf" \
  -F "eventType=Pass"
```

## Example JSON Request With CSV Text

```json
{
  "requestId": "player_a_first_half_passes",
  "eventDataCsv": "minute,second,period,type,outcomeType\n0,24,FirstHalf,Pass,Successful\n1,45,FirstHalf,Pass,Successful",
  "period": "FirstHalf",
  "eventType": "Pass",
  "outcomeType": "Successful",
  "minXT": 0.01,
  "minProgPass": 10,
  "minEndX": 80,
  "minEndY": 18,
  "maxEndY": 62,
  "preRollSeconds": 2,
  "postRollSeconds": 3,
  "timelineOffset": 5,
  "firstHalfOffset": 5,
  "secondHalfOffset": 120,
  "mergeGapSeconds": 0
}
```

Example curl:

```bash
curl -X POST http://localhost:8080/timestamps/json \
  -H "Content-Type: application/json" \
  -d '{
    "requestId": "player_a_first_half_passes",
    "eventDataCsv": "minute,second,period,type,outcomeType\n0,24,FirstHalf,Pass,Successful\n1,45,FirstHalf,Pass,Successful",
    "period": "FirstHalf",
    "eventType": "Pass",
    "outcomeType": "Successful",
    "minXT": 0.01,
    "minProgPass": 10,
    "minEndX": 80,
    "minEndY": 18,
    "maxEndY": 62,
    "preRollSeconds": 2,
    "postRollSeconds": 3,
    "timelineOffset": 5,
    "firstHalfOffset": 5,
    "secondHalfOffset": 120,
    "mergeGapSeconds": 0
  }'
```

## Example JSON Request With Structured Events

```json
{
  "events": [
    {
      "minute": 0,
      "second": 24,
      "period": "FirstHalf",
      "eventType": "Pass",
      "outcomeType": "Successful"
    },
    {
      "minute": 1,
      "second": 45,
      "period": "FirstHalf",
      "eventType": "Pass",
      "outcomeType": "Successful"
    }
  ],
  "period": "FirstHalf",
  "eventType": "Pass",
  "minXT": 0.01,
  "minProgPass": 10,
  "minEndX": 80,
  "preRollSeconds": 2,
  "postRollSeconds": 3,
  "firstHalfOffset": 5,
  "secondHalfOffset": 120
}
```

## Example Preview JSON Request With CSV Text

```json
{
  "limit": 3,
  "eventDataCsv": "minute,second,period,type,outcomeType\n0,24,FirstHalf,Pass,Successful\n1,45,FirstHalf,Carry,Successful\n46,10,SecondHalf,Pass,Successful"
}
```

Example curl:

```bash
curl -X POST http://localhost:8080/parse-preview/json \
  -H "Content-Type: application/json" \
  -d '{
    "limit": 3,
    "eventDataCsv": "minute,second,period,type,outcomeType\n0,24,FirstHalf,Pass,Successful\n1,45,FirstHalf,Carry,Successful\n46,10,SecondHalf,Pass,Successful"
  }'
```

## Input Notes

The current implementation supports CSV and XLSX uploads.

The normalization path requires enough fields to resolve event time:

- `period`
- `minute`
- `second`

Header order does not need to be fixed.
The backend normalizes known header aliases into the internal model.

For `POST /timestamps/json`, the request must include either:

- `eventDataCsv`
- or `events`

For `POST /parse-preview/json`, the request must also include either:

- `eventDataCsv`
- or `events`

## Current Limitations

- CSV and XLSX are implemented ingest formats
- schema alias support is partial, not vendor-complete
- no remote file fetch endpoint yet
- no persisted project or calibration storage yet
- no video access validation is performed by the backend

## Related Docs

- [`agent.md`](/Users/palmer/compurge/agent.md)
- [`docs/event-data-schema.md`](/Users/palmer/compurge/docs/event-data-schema.md)
- [`docs/manual-timeline-calibration-mvp-spec.md`](/Users/palmer/compurge/docs/manual-timeline-calibration-mvp-spec.md)
