# Backend Deployment

## Recommended Platform

The simplest deployment target for this backend is Render.

This backend is a good fit for a plain Go web service because:

1. it is stateless across requests
2. it uses request-scoped in-memory SQLite
3. it does not require a separate database service

## Health Check

The backend exposes:

- `GET /health`

Expected response:

```json
{
  "status": "ok"
}
```

## Required Runtime Behavior

The backend must:

1. listen on `PORT`
2. allow frontend origins through `ALLOWED_ORIGINS`
3. keep upload and concurrency limits configurable through env vars

## Render Deploy

This repo includes:

- [`render.yaml`](/Users/palmer/compurge/render.yaml)

Render can use that file directly for deploy-as-code.

### Render Service Settings

- Runtime: `Go`
- Build Command: `go build -o bin/api ./cmd/api`
- Start Command: `./bin/api`

### Recommended Environment Variables

- `PORT`
  - Render usually injects this automatically
- `MAX_UPLOAD_SIZE_BYTES`
  - default: `10485760`
- `MAX_CONCURRENT_INGEST`
  - default: `50`
- `ALLOWED_ORIGINS`
  - required for browser frontend access

Example:

```env
ALLOWED_ORIGINS=https://your-frontend.vercel.app,http://localhost:5173
```

If you want to allow any origin during early testing:

```env
ALLOWED_ORIGINS=*
```

Do not leave `*` in place for production unless that tradeoff is intentional.

## Manual Render UI Flow

1. Push the repo to GitHub
2. In Render, create a new `Web Service`
3. Connect the GitHub repo
4. Use the commands from `render.yaml` or let Render read the file directly
5. Set `ALLOWED_ORIGINS`
6. Deploy
7. Test:
   - `GET /health`
   - `POST /parse-preview`
   - `POST /timestamps`

## Frontend Connection

After deploy, take the backend URL and set it in the frontend environment:

```env
VITE_BACKEND_URL=https://your-backend.onrender.com
```

The frontend must point to the deployed backend, not localhost.

## Post-Deploy Checks

Run these checks after deploy:

1. `GET /health` returns `200`
2. browser requests from the frontend origin pass CORS
3. `POST /parse-preview` accepts CSV upload
4. `POST /timestamps` accepts CSV or XLSX upload
5. frontend can complete setup, calibration, planning, and processing flows

