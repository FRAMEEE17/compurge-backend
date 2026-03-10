# Postman Guide

## Purpose

This document explains how to test the current backend API in Postman.

## Start the API

Run the server from the repository root:

```bash
go run ./cmd/api
```

The API listens on:

`http://localhost:8080`

## Available Endpoints

- `POST /timestamps`
- `POST /timestamps/json`
- `POST /parse-preview`
- `POST /parse-preview/json`

## Import OpenAPI Into Postman

You can import the OpenAPI file directly:

- Open Postman
- Click `Import`
- Choose `File`
- Select [`docs/api/openapi.yaml`](/Users/palmer/compurge/docs/api/openapi.yaml)
- Click `Import`

Postman will create requests for:

- `POST /timestamps`
- `POST /timestamps/json`
- `POST /parse-preview`
- `POST /parse-preview/json`

After import:

- update the server URL if needed
- for `POST /timestamps`, switch to `Body -> form-data` and attach the `eventData` file manually
- for `POST /timestamps/json`, paste one of the JSON examples from this guide into `Body -> raw -> JSON`
- for `POST /parse-preview`, switch to `Body -> form-data` and attach the `eventData` file manually
- for `POST /parse-preview/json`, paste one of the preview JSON examples from this guide into `Body -> raw -> JSON`

## Option 1: Multipart Upload Request

Use this when you want to upload a CSV or XLSX file directly.

### Request Setup

- Method: `POST`
- URL: `http://localhost:8080/timestamps`
- Body: `form-data`

### Form Fields

- `eventData`
  - Type: `File`
  - Value: a `.csv` or `.xlsx` event file
- `period`
  - Type: `Text`
  - Value: `FirstHalf`
- `eventType`
  - Type: `Text`
  - Value: `Pass`
- `outcomeType`
  - Type: `Text`
  - Value: `Successful`
- `minXT`
  - Type: `Text`
  - Value: `0.01`
- `minProgPass`
  - Type: `Text`
  - Value: `10`
- `minEndX`
  - Type: `Text`
  - Value: `80`
- `minEndY`
  - Type: `Text`
  - Value: `18`
- `maxEndY`
  - Type: `Text`
  - Value: `62`
- `preRollSeconds`
  - Type: `Text`
  - Value: `2`
- `postRollSeconds`
  - Type: `Text`
  - Value: `3`
- `timelineOffset`
  - Type: `Text`
  - Value: `5`
- `firstHalfOffset`
  - Type: `Text`
  - Value: `5`
- `secondHalfOffset`
  - Type: `Text`
  - Value: `120`
- `mergeGapSeconds`
  - Type: `Text`
  - Value: `0`

### Notes

- Do not set `Content-Type` manually
- Postman will generate the correct `multipart/form-data` boundary automatically
- `.xlsx` uploads are detected by filename extension

### Expected Response

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

## Option 2: Raw JSON Request With CSV Text

Use this when you want to test the API without uploading a file.

### Request Setup

- Method: `POST`
- URL: `http://localhost:8080/timestamps/json`
- Header:
  - `Content-Type: application/json`
- Body: `raw` -> `JSON`

### Example Body

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

## Option 3: Raw JSON Request With Structured Events

Use this when the caller already has normalized event objects.

### Request Setup

- Method: `POST`
- URL: `http://localhost:8080/timestamps/json`
- Header:
  - `Content-Type: application/json`
- Body: `raw` -> `JSON`

### Example Body

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

## Option 4: Multipart Preview Request

Use this when you want a small parsed event list for timeline calibration before generating clip ranges.

### Request Setup

- Method: `POST`
- URL: `http://localhost:8080/parse-preview`
- Body: `form-data`

### Form Fields

- `eventData`
  - Type: `File`
  - Value: a `.csv` or `.xlsx` event file
- `limit`
  - Type: `Text`
  - Value: `5`

### Expected Response

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
    }
  ]
}
```

## Option 5: Raw JSON Preview Request

Use this when you want calibration preview events without uploading a file.

### Request Setup

- Method: `POST`
- URL: `http://localhost:8080/parse-preview/json`
- Header:
  - `Content-Type: application/json`
- Body: `raw` -> `JSON`

### Example Body

```json
{
  "limit": 3,
  "eventDataCsv": "minute,second,period,type,outcomeType\n0,24,FirstHalf,Pass,Successful\n1,45,FirstHalf,Carry,Successful\n46,10,SecondHalf,Pass,Successful"
}
```

## Useful Negative Tests

### Missing File

For `POST /timestamps`, omit `eventData`.

Expected:

- `400 Bad Request`

### Invalid Period

Send:

```json
{
  "eventDataCsv": "minute,second,period,type\n0,24,BadPeriod,Pass",
  "period": "BadPeriod"
}
```

Expected:

- `400 Bad Request`

### Oversized Upload

Upload a very large file to `POST /timestamps`.

Expected:

- `400 Bad Request`

### Missing JSON Event Data

Send to `POST /timestamps/json`:

```json
{
  "period": "FirstHalf",
  "eventType": "Pass"
}
```

Expected:

- `400 Bad Request`

### Invalid Preview Limit

Send to `POST /parse-preview/json`:

```json
{
  "limit": 999,
  "eventDataCsv": "minute,second,period,type,outcomeType\n0,24,FirstHalf,Pass,Successful"
}
```

Expected:

- `400 Bad Request`

## Suggested Postman Requests

- `Generate Timestamps From CSV Upload`
- `Generate Timestamps From Raw CSV JSON`
- `Generate Timestamps From Structured Events`
- `Preview Events From CSV Upload`
- `Preview Events From Raw CSV JSON`
- `Reject Missing File`
- `Reject Invalid Period`

## Related Docs

- [`docs/api/timestamps_api.md`](/Users/palmer/compurge/docs/api/timestamps_api.md)
- [`docs/event-data-schema.md`](/Users/palmer/compurge/docs/event-data-schema.md)
